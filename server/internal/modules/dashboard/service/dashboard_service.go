package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/dashboard/dto"
	investRepo "github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	premiumRepo "github.com/sovereign-fund/sovereign/internal/modules/premium/repository"
	settlRepo "github.com/sovereign-fund/sovereign/internal/modules/settlement/repository"
	walletRepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type DashboardService interface {
	GetSummary(ctx context.Context, userID string) (*dto.DashboardSummary, error)
	GetPerformance(ctx context.Context, userID string, req dto.PerformanceRequest) (*dto.PerformanceData, error)
}

type dashboardService struct {
	walletRepo  walletRepo.WalletRepository
	premiumRepo premiumRepo.PremiumRepository
	investRepo  investRepo.InvestmentRepository
	settlRepo   settlRepo.SettlementRepository
	logger      *slog.Logger
}

func NewDashboardService(
	wr walletRepo.WalletRepository,
	pr premiumRepo.PremiumRepository,
	ir investRepo.InvestmentRepository,
	sr settlRepo.SettlementRepository,
	logger *slog.Logger,
) DashboardService {
	return &dashboardService{
		walletRepo:  wr,
		premiumRepo: pr,
		investRepo:  ir,
		settlRepo:   sr,
		logger:      logger,
	}
}

func (s *dashboardService) GetSummary(ctx context.Context, userID string) (*dto.DashboardSummary, error) {
	wallets, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	portfolio := dto.PortfolioSummary{Currency: "USDT"}
	for _, w := range wallets {
		if w.Currency == "USDT" {
			portfolio.Available = portfolio.Available.Add(w.Available)
			portfolio.InOperation = portfolio.InOperation.Add(w.InOperation)
			portfolio.Frozen = portfolio.Frozen.Add(w.Frozen)
			portfolio.TotalValue = portfolio.TotalValue.Add(w.TotalBalance())
		}
	}
	portfolio.TotalValueUSD = portfolio.TotalValue

	premiumSummary := dto.PremiumSummary{
		Pair:  "BTC/KRW",
		Trend: "flat",
	}

	latestTicks, err := s.premiumRepo.FindLatest(ctx)
	if err == nil && len(latestTicks) > 0 {
		for _, t := range latestTicks {
			if t.Pair == "BTC/KRW" {
				premiumSummary.CurrentPct = t.PremiumPct
				premiumSummary.LastUpdated = t.CreatedAt.Format(time.RFC3339)
				break
			}
		}
	}

	// 收益统计
	invs, _ := s.investRepo.FindByUserID(ctx, userID)
	cumReturn := decimal.Zero
	for _, inv := range invs {
		cumReturn = cumReturn.Add(inv.NetReturn)
	}
	cumReturnPct := decimal.Zero
	if portfolio.TotalValue.GreaterThan(decimal.Zero) {
		cumReturnPct = cumReturn.Div(portfolio.TotalValue).Mul(decimal.NewFromInt(100))
	}

	perf := dto.PerformanceData{
		CumulativeReturn:    cumReturn,
		CumulativeReturnPct: cumReturnPct,
		AnnualizedReturn:    decimal.Zero,
		HighWaterMark:       portfolio.TotalValue,
		Chart:               []dto.PerformancePoint{},
	}

	return &dto.DashboardSummary{
		Portfolio:   portfolio,
		Performance: perf,
		Premium:     premiumSummary,
	}, nil
}

func (s *dashboardService) GetPerformance(ctx context.Context, userID string, req dto.PerformanceRequest) (*dto.PerformanceData, error) {
	period := req.Period
	if period == "" {
		period = "1M"
	}

	now := time.Now()
	var from time.Time
	switch period {
	case "1W":
		from = now.AddDate(0, 0, -7)
	case "1M":
		from = now.AddDate(0, -1, 0)
	case "3M":
		from = now.AddDate(0, -3, 0)
	case "6M":
		from = now.AddDate(0, -6, 0)
	case "1Y":
		from = now.AddDate(-1, 0, 0)
	case "ALL":
		from = now.AddDate(-5, 0, 0)
	default:
		from = now.AddDate(0, -1, 0)
	}

	fromPeriod := from.Format("2006-01-02")
	toPeriod := now.Format("2006-01-02")

	// 用结算记录生成收益曲线
	settlements, err := s.settlRepo.FindByUserIDAndPeriodRange(ctx, userID, fromPeriod, toPeriod)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	// 按日期聚合净收益（同一天可能有多个投资的结算）
	dailyReturn := make(map[string]decimal.Decimal)
	for _, s := range settlements {
		dailyReturn[s.Period] = dailyReturn[s.Period].Add(s.NetReturn)
	}

	// 生成连续日期的累计收益曲线
	chart := make([]dto.PerformancePoint, 0)
	cumValue := decimal.Zero
	highWaterMark := decimal.Zero

	for d := from; !d.After(now); d = d.AddDate(0, 0, 1) {
		day := d.Format("2006-01-02")
		if ret, ok := dailyReturn[day]; ok {
			cumValue = cumValue.Add(ret)
		}
		if cumValue.GreaterThan(highWaterMark) {
			highWaterMark = cumValue
		}
		chart = append(chart, dto.PerformancePoint{
			Date:  day,
			Value: cumValue,
		})
	}

	// 计算累计收益率
	invs, _ := s.investRepo.FindByUserID(ctx, userID)
	totalInvested := decimal.Zero
	for _, inv := range invs {
		totalInvested = totalInvested.Add(inv.Amount)
	}

	cumReturnPct := decimal.Zero
	if totalInvested.GreaterThan(decimal.Zero) {
		cumReturnPct = cumValue.Div(totalInvested).Mul(decimal.NewFromInt(100))
	}

	return &dto.PerformanceData{
		CumulativeReturn:    cumValue,
		CumulativeReturnPct: cumReturnPct,
		HighWaterMark:       highWaterMark,
		Chart:               chart,
	}, nil
}
