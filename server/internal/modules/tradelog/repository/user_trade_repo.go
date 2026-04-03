package repository

import (
	"context"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/model"
	"gorm.io/gorm"
)

type UserTradeFilters struct {
	Pair         string
	InvestmentID string
	From         time.Time
	To           time.Time
}

type UserTradeRepository interface {
	BatchCreate(ctx context.Context, trades []model.UserTrade) error
	FindByUserID(ctx context.Context, userID string, filters UserTradeFilters, limit, offset int) ([]model.UserTrade, int64, error)
	SummarizeByUserID(ctx context.Context, userID string, filters UserTradeFilters) (*TradeSummaryResult, error)
	ExistsBySettlementID(ctx context.Context, settlementID string) (bool, error)
}

type userTradeRepository struct {
	db *gorm.DB
}

func NewUserTradeRepository(db *gorm.DB) UserTradeRepository {
	return &userTradeRepository{db: db}
}

func (r *userTradeRepository) BatchCreate(ctx context.Context, trades []model.UserTrade) error {
	if len(trades) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(trades, 500).Error
}

func (r *userTradeRepository) FindByUserID(ctx context.Context, userID string, filters UserTradeFilters, limit, offset int) ([]model.UserTrade, int64, error) {
	var trades []model.UserTrade
	var total int64

	query := r.db.WithContext(ctx).Model(&model.UserTrade{}).Where("user_id = ?", userID)
	query = r.applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("executed_at DESC").Limit(limit).Offset(offset).Find(&trades).Error
	return trades, total, err
}

func (r *userTradeRepository) SummarizeByUserID(ctx context.Context, userID string, filters UserTradeFilters) (*TradeSummaryResult, error) {
	query := r.db.WithContext(ctx).Model(&model.UserTrade{}).Where("user_id = ?", userID)
	query = r.applyFilters(query, filters)

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

func (r *userTradeRepository) ExistsBySettlementID(ctx context.Context, settlementID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserTrade{}).Where("settlement_id = ?", settlementID).Count(&count).Error
	return count > 0, err
}

func (r *userTradeRepository) applyFilters(query *gorm.DB, filters UserTradeFilters) *gorm.DB {
	if filters.Pair != "" {
		query = query.Where("pair = ?", filters.Pair)
	}
	if filters.InvestmentID != "" {
		query = query.Where("investment_id = ?", filters.InvestmentID)
	}
	if !filters.From.IsZero() {
		query = query.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		query = query.Where("executed_at <= ?", filters.To)
	}
	return query
}
