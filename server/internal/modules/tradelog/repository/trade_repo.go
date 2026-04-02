package repository

import (
	"context"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/model"
	"gorm.io/gorm"
)

type TradeSummaryResult struct {
	TotalTrades int64
	TotalPnL    float64
	AvgPremium  float64
	WinCount    int64
}

type TradeRepository interface {
	Create(ctx context.Context, trade *model.Trade) error
	FindByUserID(ctx context.Context, userID string, filters TradeFilters, limit, offset int) ([]model.Trade, int64, error)
	FindAll(ctx context.Context, filters TradeFilters, limit, offset int) ([]model.Trade, int64, error)
	FindByInvestmentID(ctx context.Context, investmentID string, limit, offset int) ([]model.Trade, int64, error)
	CountByInvestmentAndPeriod(ctx context.Context, investmentID string, from, to time.Time) (int64, error)
	SummarizeByUserID(ctx context.Context, userID string, filters TradeFilters) (*TradeSummaryResult, error)
	SummarizeAll(ctx context.Context, filters TradeFilters) (*TradeSummaryResult, error)
	SummarizeByPeriod(ctx context.Context, from, to time.Time) (*TradeSummaryResult, error)
}

type TradeFilters struct {
	InvestmentID string
	Pair         string
	From         time.Time
	To           time.Time
}

type tradeRepository struct {
	db *gorm.DB
}

func NewTradeRepository(db *gorm.DB) TradeRepository {
	return &tradeRepository{db: db}
}

func (r *tradeRepository) Create(ctx context.Context, trade *model.Trade) error {
	return r.db.WithContext(ctx).Create(trade).Error
}

func (r *tradeRepository) FindByUserID(ctx context.Context, userID string, filters TradeFilters, limit, offset int) ([]model.Trade, int64, error) {
	var trades []model.Trade
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Trade{}).Where("user_id = ?", userID)

	if filters.InvestmentID != "" {
		query = query.Where("investment_id = ?", filters.InvestmentID)
	}
	if filters.Pair != "" {
		query = query.Where("pair = ?", filters.Pair)
	}
	if !filters.From.IsZero() {
		query = query.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("executed_at <= ?", filters.To)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("executed_at DESC").Limit(limit).Offset(offset).Find(&trades).Error
	return trades, total, err
}

func (r *tradeRepository) FindAll(ctx context.Context, filters TradeFilters, limit, offset int) ([]model.Trade, int64, error) {
	var trades []model.Trade
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Trade{})
	if filters.Pair != "" {
		query = query.Where("pair = ?", filters.Pair)
	}
	if !filters.From.IsZero() {
		query = query.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("executed_at <= ?", filters.To)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("executed_at DESC").Limit(limit).Offset(offset).Find(&trades).Error
	return trades, total, err
}

func (r *tradeRepository) SummarizeAll(ctx context.Context, filters TradeFilters) (*TradeSummaryResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Trade{})
	if filters.Pair != "" {
		query = query.Where("pair = ?", filters.Pair)
	}
	if !filters.From.IsZero() {
		query = query.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("executed_at <= ?", filters.To)
	}

	var result TradeSummaryResult
	err := query.Select(`
		COUNT(*) as total_trades,
		COALESCE(SUM(pnl), 0) as total_pnl,
		COALESCE(AVG(premium_pct), 0) as avg_premium,
		COUNT(CASE WHEN pnl > 0 THEN 1 END) as win_count
	`).Row().Scan(&result.TotalTrades, &result.TotalPnL, &result.AvgPremium, &result.WinCount)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *tradeRepository) SummarizeByPeriod(ctx context.Context, from, to time.Time) (*TradeSummaryResult, error) {
	return r.SummarizeAll(ctx, TradeFilters{From: from, To: to})
}

func (r *tradeRepository) FindByInvestmentID(ctx context.Context, investmentID string, limit, offset int) ([]model.Trade, int64, error) {
	var trades []model.Trade
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Trade{}).Where("investment_id = ?", investmentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("executed_at DESC").Limit(limit).Offset(offset).Find(&trades).Error
	return trades, total, err
}

func (r *tradeRepository) SummarizeByUserID(ctx context.Context, userID string, filters TradeFilters) (*TradeSummaryResult, error) {
	query := r.db.WithContext(ctx).Model(&model.Trade{}).Where("user_id = ?", userID)
	if filters.InvestmentID != "" {
		query = query.Where("investment_id = ?", filters.InvestmentID)
	}
	if filters.Pair != "" {
		query = query.Where("pair = ?", filters.Pair)
	}
	if !filters.From.IsZero() {
		query = query.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("executed_at <= ?", filters.To)
	}

	var result TradeSummaryResult
	err := query.Select(`
		COUNT(*) as total_trades,
		COALESCE(SUM(pnl), 0) as total_pnl,
		COALESCE(AVG(premium_pct), 0) as avg_premium,
		COUNT(CASE WHEN pnl > 0 THEN 1 END) as win_count
	`).Row().Scan(&result.TotalTrades, &result.TotalPnL, &result.AvgPremium, &result.WinCount)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *tradeRepository) CountByInvestmentAndPeriod(ctx context.Context, investmentID string, from, to time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Trade{}).
		Where("investment_id = ? AND executed_at >= ? AND executed_at < ?", investmentID, from, to).
		Count(&count).Error
	return count, err
}
