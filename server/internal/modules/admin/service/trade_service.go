package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	settlemodel "github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	trademodel "github.com/sovereign-fund/sovereign/internal/modules/tradelog/model"
	"gorm.io/gorm"
)

type TradeService interface {
	List(ctx context.Context, query dto.TradeListQuery) ([]dto.TradeListItem, int64, error)
	Stats(ctx context.Context) (*dto.TradeStats, error)
}

type tradeService struct {
	db *gorm.DB
}

func NewTradeService(db *gorm.DB) TradeService {
	return &tradeService{db: db}
}

func (s *tradeService) List(ctx context.Context, query dto.TradeListQuery) ([]dto.TradeListItem, int64, error) {
	db := s.db.WithContext(ctx).Model(&trademodel.Trade{})

	if query.Pair != "" {
		db = db.Where("pair ILIKE ?", "%"+query.Pair+"%")
	}
	if query.DateFrom != "" {
		db = db.Where("executed_at >= ?", query.DateFrom)
	}
	if query.DateTo != "" {
		db = db.Where("executed_at < ?", query.DateTo+" 23:59:59")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count trades: %w", err)
	}

	var trades []trademodel.Trade
	offset := (query.Page - 1) * query.Limit
	if err := db.Order("executed_at DESC").Offset(offset).Limit(query.Limit).Find(&trades).Error; err != nil {
		return nil, 0, fmt.Errorf("find trades: %w", err)
	}

	items := make([]dto.TradeListItem, len(trades))
	for i, t := range trades {
		items[i] = dto.TradeListItem{
			ID:           t.ID,
			Pair:         t.Pair,
			BuyExchange:  t.BuyExchange,
			SellExchange: t.SellExchange,
			BuyPrice:     t.BuyPrice.StringFixed(4),
			SellPrice:    t.SellPrice.StringFixed(4),
			Amount:       t.Amount.StringFixed(2),
			PremiumPct:   t.PremiumPct.StringFixed(2),
			PnL:          t.PnL.StringFixed(2),
			Fee:          t.Fee.StringFixed(2),
			ExecutedAt:   t.ExecutedAt.Format(time.RFC3339),
		}
	}

	return items, total, nil
}

func (s *tradeService) Stats(ctx context.Context) (*dto.TradeStats, error) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	day7 := today.AddDate(0, 0, -7)
	day30 := today.AddDate(0, 0, -30)

	var pnl1d, pnl7d, pnl30d decimal.Decimal
	var count1d, count7d, count30d int64

	if err := s.db.WithContext(ctx).Model(&trademodel.Trade{}).
		Where("executed_at >= ?", today).
		Select("COALESCE(SUM(pnl), 0)").Scan(&pnl1d).Error; err != nil {
		return nil, fmt.Errorf("sum pnl 1d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&trademodel.Trade{}).
		Where("executed_at >= ?", today).Count(&count1d).Error; err != nil {
		return nil, fmt.Errorf("count trades 1d: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&trademodel.Trade{}).
		Where("executed_at >= ?", day7).
		Select("COALESCE(SUM(pnl), 0)").Scan(&pnl7d).Error; err != nil {
		return nil, fmt.Errorf("sum pnl 7d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&trademodel.Trade{}).
		Where("executed_at >= ?", day7).Count(&count7d).Error; err != nil {
		return nil, fmt.Errorf("count trades 7d: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&trademodel.Trade{}).
		Where("executed_at >= ?", day30).
		Select("COALESCE(SUM(pnl), 0)").Scan(&pnl30d).Error; err != nil {
		return nil, fmt.Errorf("sum pnl 30d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&trademodel.Trade{}).
		Where("executed_at >= ?", day30).Count(&count30d).Error; err != nil {
		return nil, fmt.Errorf("count trades 30d: %w", err)
	}

	var userProfit1d, userProfit7d, userProfit30d decimal.Decimal
	if err := s.db.WithContext(ctx).Model(&settlemodel.Settlement{}).
		Where("settled_at >= ?", today).
		Select("COALESCE(SUM(net_return), 0)").Scan(&userProfit1d).Error; err != nil {
		return nil, fmt.Errorf("sum user profit 1d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&settlemodel.Settlement{}).
		Where("settled_at >= ?", day7).
		Select("COALESCE(SUM(net_return), 0)").Scan(&userProfit7d).Error; err != nil {
		return nil, fmt.Errorf("sum user profit 7d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&settlemodel.Settlement{}).
		Where("settled_at >= ?", day30).
		Select("COALESCE(SUM(net_return), 0)").Scan(&userProfit30d).Error; err != nil {
		return nil, fmt.Errorf("sum user profit 30d: %w", err)
	}

	return &dto.TradeStats{
		PnL1D:         pnl1d.StringFixed(2),
		PnL7D:         pnl7d.StringFixed(2),
		PnL30D:        pnl30d.StringFixed(2),
		UserProfit1D:  userProfit1d.StringFixed(2),
		UserProfit7D:  userProfit7d.StringFixed(2),
		UserProfit30D: userProfit30d.StringFixed(2),
		TradeCount1D:  count1d,
		TradeCount7D:  count7d,
		TradeCount30D: count30d,
	}, nil
}
