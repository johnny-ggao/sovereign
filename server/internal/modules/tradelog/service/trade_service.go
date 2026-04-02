package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/model"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type TradeService interface {
	GetTrades(ctx context.Context, userID string, filters dto.TradeFilterRequest, page, perPage int) (*dto.TradeListResponse, int64, error)
	GetAllTrades(ctx context.Context, filters dto.TradeFilterRequest, page, perPage int) (*dto.TradeListResponse, int64, error)
	CreateTrade(ctx context.Context, req dto.CreateTradeRequest) (*dto.TradeResponse, error)
	BatchCreateTrades(ctx context.Context, req dto.BatchCreateTradeRequest) (int, error)
	ExportCSV(ctx context.Context, userID string, filters dto.TradeFilterRequest) ([]byte, error)
}

type tradeService struct {
	tradeRepo repository.TradeRepository
	logger    *slog.Logger
}

func NewTradeService(repo repository.TradeRepository, logger *slog.Logger) TradeService {
	return &tradeService{tradeRepo: repo, logger: logger}
}

// CreateTrade 创建基金级套利交易记录（不绑定具体投资）
func (s *tradeService) CreateTrade(ctx context.Context, req dto.CreateTradeRequest) (*dto.TradeResponse, error) {
	trade, err := s.buildTrade(req)
	if err != nil {
		return nil, err
	}

	if err := s.tradeRepo.Create(ctx, trade); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("create trade: %w", err))
	}

	s.logger.Info("trade created",
		slog.String("trade_id", trade.ID),
		slog.String("pair", trade.Pair),
		slog.String("pnl", trade.PnL.String()),
	)

	return toTradeResponse(trade), nil
}

// BatchCreateTrades 批量创建基金级套利交易记录
func (s *tradeService) BatchCreateTrades(ctx context.Context, req dto.BatchCreateTradeRequest) (int, error) {
	created := 0
	for _, r := range req.Trades {
		trade, err := s.buildTrade(r)
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

	s.logger.Info("batch trades created", slog.Int("count", created))
	return created, nil
}

func (s *tradeService) buildTrade(req dto.CreateTradeRequest) (*model.Trade, error) {
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

// GetTrades 获取用户的交易记录（前端用）
func (s *tradeService) GetTrades(ctx context.Context, userID string, filters dto.TradeFilterRequest, page, perPage int) (*dto.TradeListResponse, int64, error) {
	offset := (page - 1) * perPage

	repoFilters := repository.TradeFilters{
		Pair: filters.Pair,
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

	// 基金级交易不绑定用户，查全部
	return s.queryTrades(ctx, repoFilters, perPage, offset)
}

// GetAllTrades 获取所有交易记录（基金级）
func (s *tradeService) GetAllTrades(ctx context.Context, filters dto.TradeFilterRequest, page, perPage int) (*dto.TradeListResponse, int64, error) {
	offset := (page - 1) * perPage

	repoFilters := repository.TradeFilters{
		Pair: filters.Pair,
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

	return s.queryTrades(ctx, repoFilters, perPage, offset)
}

func (s *tradeService) queryTrades(ctx context.Context, filters repository.TradeFilters, limit, offset int) (*dto.TradeListResponse, int64, error) {
	trades, _, err := s.tradeRepo.FindAll(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.ErrInternal, err)
	}

	summary, err := s.tradeRepo.SummarizeAll(ctx, filters)
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
		Pair: filters.Pair,
	}

	trades, _, err := s.tradeRepo.FindAll(ctx, repoFilters, 10000, 0)
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
