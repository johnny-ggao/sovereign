package repository

import (
	"context"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/model"
	tradelogrepo "github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	"gorm.io/gorm"
)

type TradingTradeRepository interface {
	Create(ctx context.Context, t *model.TradingTrade) error
	FindByID(ctx context.Context, id string) (*model.TradingTrade, error)
	FindAll(ctx context.Context, filters tradelogrepo.TradeFilters, limit, offset int) ([]model.TradingTrade, int64, error)
	SummarizeByPeriod(ctx context.Context, from, to time.Time) (*tradelogrepo.TradeSummaryResult, error)
	SummarizeAll(ctx context.Context, filters tradelogrepo.TradeFilters) (*tradelogrepo.TradeSummaryResult, error)
	FindByPeriod(ctx context.Context, from, to time.Time) ([]model.TradingTrade, error)
	Delete(ctx context.Context, id string) error
	BatchCreate(ctx context.Context, trades []model.TradingTrade) error
}

type tradingTradeRepository struct {
	db *gorm.DB
}

func NewTradingTradeRepository(db *gorm.DB) TradingTradeRepository {
	return &tradingTradeRepository{db: db}
}

func (r *tradingTradeRepository) Create(ctx context.Context, t *model.TradingTrade) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *tradingTradeRepository) BatchCreate(ctx context.Context, trades []model.TradingTrade) error {
	if len(trades) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(trades, 100).Error
}

func (r *tradingTradeRepository) FindByID(ctx context.Context, id string) (*model.TradingTrade, error) {
	var t model.TradingTrade
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tradingTradeRepository) FindAll(ctx context.Context, filters tradelogrepo.TradeFilters, limit, offset int) ([]model.TradingTrade, int64, error) {
	var trades []model.TradingTrade
	var total int64
	q := r.db.WithContext(ctx).Model(&model.TradingTrade{})
	if filters.Pair != "" {
		q = q.Where("pair = ?", filters.Pair)
	}
	if !filters.From.IsZero() {
		q = q.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		q = q.Where("executed_at <= ?", filters.To)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("executed_at DESC").Limit(limit).Offset(offset).Find(&trades).Error
	return trades, total, err
}

func (r *tradingTradeRepository) SummarizeAll(ctx context.Context, filters tradelogrepo.TradeFilters) (*tradelogrepo.TradeSummaryResult, error) {
	q := r.db.WithContext(ctx).Model(&model.TradingTrade{})
	if !filters.From.IsZero() {
		q = q.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		q = q.Where("executed_at <= ?", filters.To)
	}
	var result tradelogrepo.TradeSummaryResult
	err := q.Select(`
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

func (r *tradingTradeRepository) SummarizeByPeriod(ctx context.Context, from, to time.Time) (*tradelogrepo.TradeSummaryResult, error) {
	return r.SummarizeAll(ctx, tradelogrepo.TradeFilters{From: from, To: to})
}

func (r *tradingTradeRepository) FindByPeriod(ctx context.Context, from, to time.Time) ([]model.TradingTrade, error) {
	var trades []model.TradingTrade
	err := r.db.WithContext(ctx).
		Where("executed_at >= ? AND executed_at < ?", from, to).
		Order("executed_at ASC").
		Find(&trades).Error
	return trades, err
}

func (r *tradingTradeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.TradingTrade{}).Error
}
