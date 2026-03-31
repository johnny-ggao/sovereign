package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	investModel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	investRepo "github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/model"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type TradeService interface {
	GetTrades(ctx context.Context, userID string, filters dto.TradeFilterRequest, page, perPage int) (*dto.TradeListResponse, int64, error)
	CreateTrade(ctx context.Context, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	BatchCreateTrades(ctx context.Context, req dto.BatchCreateTradeRequest) (int, error)
	ExportCSV(ctx context.Context, userID string, filters dto.TradeFilterRequest) ([]byte, error)
}

type tradeService struct {
	tradeRepo  repository.TradeRepository
	investRepo investRepo.InvestmentRepository
	logger     *slog.Logger
}

func NewTradeService(repo repository.TradeRepository, ir investRepo.InvestmentRepository, logger *slog.Logger) TradeService {
	return &tradeService{tradeRepo: repo, investRepo: ir, logger: logger}
}

func (s *tradeService) CreateTrade(ctx context.Context, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
	// 校验投资存在
	inv, err := s.investRepo.FindByID(ctx, req.InvestmentID)
	if err != nil {
		return nil, apperr.New(400, "INVALID_INVESTMENT", "investment not found")
	}
	if inv.Status != investModel.InvestStatusActive {
		return nil, apperr.New(400, "INVESTMENT_NOT_ACTIVE", "investment is not active")
	}

	trade, err := s.buildTrade(req, inv.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.tradeRepo.Create(ctx, trade); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("create trade: %w", err))
	}

	// 更新投资收益
	if err := s.updateInvestmentReturns(ctx, inv); err != nil {
		s.logger.Error("update investment returns failed", slog.String("error", err.Error()))
	}

	s.logger.Info("trade created",
		slog.String("trade_id", trade.ID),
		slog.String("investment_id", trade.InvestmentID),
		slog.String("pair", trade.Pair),
		slog.String("pnl", trade.PnL.String()),
	)

	return toTradeResponse(trade), nil
}

func (s *tradeService) BatchCreateTrades(ctx context.Context, req dto.BatchCreateTradeRequest) (int, error) {
	// 按 investment_id 分组
	invIDs := make(map[string]bool)
	for _, t := range req.Trades {
		invIDs[t.InvestmentID] = true
	}

	// 校验所有投资存在且 active
	invMap := make(map[string]*investModel.Investment)
	for id := range invIDs {
		inv, err := s.investRepo.FindByID(ctx, id)
		if err != nil {
			return 0, apperr.New(400, "INVALID_INVESTMENT", fmt.Sprintf("investment %s not found", id))
		}
		if inv.Status != investModel.InvestStatusActive {
			return 0, apperr.New(400, "INVESTMENT_NOT_ACTIVE", fmt.Sprintf("investment %s is not active", id))
		}
		invMap[id] = inv
	}

	created := 0
	for _, req := range req.Trades {
		inv := invMap[req.InvestmentID]
		trade, err := s.buildTrade(req, inv.UserID)
		if err != nil {
			s.logger.Error("skip invalid trade", slog.String("error", err.Error()))
			continue
		}
		if err := s.tradeRepo.Create(ctx, trade); err != nil {
			s.logger.Error("create trade failed", slog.String("error", err.Error()))
			continue
		}
		created++
	}

	// 更新所有涉及的投资收益
	for _, inv := range invMap {
		if err := s.updateInvestmentReturns(ctx, inv); err != nil {
			s.logger.Error("update investment returns failed",
				slog.String("investment_id", inv.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	return created, nil
}

func (s *tradeService) buildTrade(req dto.CreateTradeRequest, userID string) (*model.Trade, error) {
	buyPrice, err := decimal.NewFromString(req.BuyPrice)
	if err != nil {
		return nil, apperr.New(400, "INVALID_PRICE", "invalid buy_price")
	}
	sellPrice, err := decimal.NewFromString(req.SellPrice)
	if err != nil {
		return nil, apperr.New(400, "INVALID_PRICE", "invalid sell_price")
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, apperr.New(400, "INVALID_AMOUNT", "invalid amount")
	}
	premiumPct, err := decimal.NewFromString(req.PremiumPct)
	if err != nil {
		return nil, apperr.New(400, "INVALID_PREMIUM", "invalid premium_pct")
	}
	pnl, err := decimal.NewFromString(req.PnL)
	if err != nil {
		return nil, apperr.New(400, "INVALID_PNL", "invalid pnl")
	}
	fee := decimal.Zero
	if req.Fee != "" {
		fee, _ = decimal.NewFromString(req.Fee)
	}
	executedAt, err := time.Parse(time.RFC3339, req.ExecutedAt)
	if err != nil {
		return nil, apperr.New(400, "INVALID_TIME", "invalid executed_at, use RFC3339 format")
	}

	return &model.Trade{
		InvestmentID: req.InvestmentID,
		UserID:       userID,
		Pair:         req.Pair,
		BuyExchange:  req.BuyExchange,
		SellExchange: req.SellExchange,
		BuyPrice:     buyPrice,
		SellPrice:    sellPrice,
		Amount:       amount,
		PremiumPct:   premiumPct,
		PnL:          pnl,
		Fee:          fee,
		ExecutedAt:   executedAt,
	}, nil
}

func (s *tradeService) updateInvestmentReturns(ctx context.Context, inv *investModel.Investment) error {
	summary, err := s.tradeRepo.SummarizeByUserID(ctx, inv.UserID, repository.TradeFilters{
		InvestmentID: inv.ID,
	})
	if err != nil {
		return err
	}

	totalPnL := decimal.NewFromFloat(summary.TotalPnL)
	perfFee := decimal.Zero
	if totalPnL.GreaterThan(decimal.Zero) {
		perfFee = totalPnL.Mul(decimal.NewFromFloat(0.5))
	}
	netReturn := totalPnL.Sub(perfFee)

	updated := *inv
	updated.TotalReturn = totalPnL
	updated.PerformanceFee = perfFee
	updated.NetReturn = netReturn

	return s.investRepo.Update(ctx, &updated)
}

func toTradeResponse(t *model.Trade) *dto.TradeResponse {
	return &dto.TradeResponse{
		ID:           t.ID,
		InvestmentID: t.InvestmentID,
		Pair:         t.Pair,
		BuyExchange:  t.BuyExchange,
		SellExchange: t.SellExchange,
		BuyPrice:     t.BuyPrice,
		SellPrice:    t.SellPrice,
		Amount:       t.Amount,
		PremiumPct:   t.PremiumPct,
		PnL:          t.PnL,
		Fee:          t.Fee,
		ExecutedAt:   t.ExecutedAt.Format(time.RFC3339),
	}
}

func (s *tradeService) GetTrades(ctx context.Context, userID string, filters dto.TradeFilterRequest, page, perPage int) (*dto.TradeListResponse, int64, error) {
	offset := (page - 1) * perPage

	repoFilters := repository.TradeFilters{
		InvestmentID: filters.InvestmentID,
		Pair:         filters.Pair,
	}
	if filters.From != "" {
		if t, err := time.Parse(time.RFC3339, filters.From); err == nil {
			repoFilters.From = t
		}
	}
	if filters.To != "" {
		if t, err := time.Parse(time.RFC3339, filters.To); err == nil {
			repoFilters.To = t
		}
	}

	trades, _, err := s.tradeRepo.FindByUserID(ctx, userID, repoFilters, perPage, offset)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.ErrInternal, err)
	}

	summary, err := s.tradeRepo.SummarizeByUserID(ctx, userID, repoFilters)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.ErrInternal, err)
	}

	resp := make([]dto.TradeResponse, 0, len(trades))
	for _, t := range trades {
		resp = append(resp, *toTradeResponse(&t))
	}

	winRate := decimal.Zero
	if summary.TotalTrades > 0 {
		winRate = decimal.NewFromInt(summary.WinCount).Div(decimal.NewFromInt(summary.TotalTrades)).Mul(decimal.NewFromInt(100))
	}

	return &dto.TradeListResponse{
		Trades: resp,
		Summary: dto.TradeSummary{
			TotalTrades:   summary.TotalTrades,
			TotalPnL:      decimal.NewFromFloat(summary.TotalPnL),
			AvgPremiumPct: decimal.NewFromFloat(summary.AvgPremium),
			WinRate:       winRate,
		},
	}, summary.TotalTrades, nil
}

func (s *tradeService) ExportCSV(ctx context.Context, userID string, filters dto.TradeFilterRequest) ([]byte, error) {
	repoFilters := repository.TradeFilters{
		InvestmentID: filters.InvestmentID,
		Pair:         filters.Pair,
	}

	trades, _, err := s.tradeRepo.FindByUserID(ctx, userID, repoFilters, 10000, 0)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	csv := "id,pair,buy_exchange,sell_exchange,buy_price,sell_price,amount,premium_pct,pnl,fee,executed_at\n"
	for _, t := range trades {
		csv += t.ID + "," + t.Pair + "," + t.BuyExchange + "," + t.SellExchange + "," +
			t.BuyPrice.String() + "," + t.SellPrice.String() + "," + t.Amount.String() + "," +
			t.PremiumPct.String() + "," + t.PnL.String() + "," + t.Fee.String() + "," +
			t.ExecutedAt.Format(time.RFC3339) + "\n"
	}

	return []byte(csv), nil
}
