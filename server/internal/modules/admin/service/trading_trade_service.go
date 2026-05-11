package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	settlemodel "github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	tradingmodel "github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/model"
	tradingrepo "github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/repository"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TradingTradeService interface {
	List(ctx context.Context, query dto.TradeListQuery) ([]dto.TradeListItem, int64, error)
	Stats(ctx context.Context) (*dto.TradeStats, error)
	DownloadTemplate(ctx context.Context) (*excelize.File, error)
	ImportFromExcel(ctx context.Context, file multipart.File) (int, []string, error)
	Delete(ctx context.Context, tradeID string) error
}

type tradingTradeService struct {
	db   *gorm.DB
	repo tradingrepo.TradingTradeRepository
}

func NewTradingTradeService(db *gorm.DB) TradingTradeService {
	return &tradingTradeService{db: db, repo: tradingrepo.NewTradingTradeRepository(db)}
}

func (s *tradingTradeService) List(ctx context.Context, query dto.TradeListQuery) ([]dto.TradeListItem, int64, error) {
	db := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{})

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
		return nil, 0, fmt.Errorf("count trading_trades: %w", err)
	}

	var trades []tradingmodel.TradingTrade
	offset := (query.Page - 1) * query.Limit
	if err := db.Order("executed_at DESC").Offset(offset).Limit(query.Limit).Find(&trades).Error; err != nil {
		return nil, 0, fmt.Errorf("find trading_trades: %w", err)
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
			Source:       t.Source,
			ExecutedAt:   t.ExecutedAt.Format(time.RFC3339),
		}
	}

	return items, total, nil
}

func (s *tradingTradeService) DownloadTemplate(_ context.Context) (*excelize.File, error) {
	file := excelize.NewFile()
	sheet := file.GetSheetList()[0]
	file.SetSheetName(sheet, tradeTemplateSheetName)
	for col, value := range tradeTemplateHeaders {
		if err := file.SetCellValue(tradeTemplateSheetName, fmt.Sprintf("%s1", tradeTemplateColumns[col]), value); err != nil {
			return nil, fmt.Errorf("set template header: %w", err)
		}
	}
	for col, value := range tradeTemplateSampleRow {
		if err := file.SetCellValue(tradeTemplateSheetName, fmt.Sprintf("%s2", tradeTemplateColumns[col]), value); err != nil {
			return nil, fmt.Errorf("set template sample row: %w", err)
		}
	}
	for _, width := range tradeTemplateWidths {
		if err := file.SetColWidth(tradeTemplateSheetName, width.Start, width.End, width.Width); err != nil {
			return nil, fmt.Errorf("set template column width: %w", err)
		}
	}
	return file, nil
}

func (s *tradingTradeService) ImportFromExcel(ctx context.Context, file multipart.File) (int, []string, error) {
	parsed, rowErrors, err := ParseImportRows(file)
	if err != nil {
		return 0, nil, err
	}
	if len(parsed) == 0 {
		return 0, rowErrors, nil
	}

	trading := make([]tradingmodel.TradingTrade, len(parsed))
	for i, p := range parsed {
		trading[i] = tradingmodel.TradingTrade{
			Pair:         p.Pair,
			BuyExchange:  p.BuyExchange,
			SellExchange: p.SellExchange,
			BuyPrice:     p.BuyPrice,
			SellPrice:    p.SellPrice,
			Amount:       p.Amount,
			PremiumPct:   p.PremiumPct,
			PnL:          p.PnL,
			Fee:          p.Fee,
			Source:       p.Source,
			ExecutedAt:   p.ExecutedAt,
		}
	}

	if err := s.repo.BatchCreate(ctx, trading); err != nil {
		return 0, rowErrors, fmt.Errorf("import trading_trades: %w", err)
	}
	return len(trading), rowErrors, nil
}

func (s *tradingTradeService) Stats(ctx context.Context) (*dto.TradeStats, error) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	day7 := today.AddDate(0, 0, -7)
	day30 := today.AddDate(0, 0, -30)

	var pnl1d, pnl7d, pnl30d decimal.Decimal
	var count1d, count7d, count30d int64

	if err := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{}).
		Where("executed_at >= ?", today).
		Select("COALESCE(SUM(pnl), 0)").Scan(&pnl1d).Error; err != nil {
		return nil, fmt.Errorf("sum pnl 1d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{}).
		Where("executed_at >= ?", today).Count(&count1d).Error; err != nil {
		return nil, fmt.Errorf("count trading_trades 1d: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{}).
		Where("executed_at >= ?", day7).
		Select("COALESCE(SUM(pnl), 0)").Scan(&pnl7d).Error; err != nil {
		return nil, fmt.Errorf("sum pnl 7d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{}).
		Where("executed_at >= ?", day7).Count(&count7d).Error; err != nil {
		return nil, fmt.Errorf("count trading_trades 7d: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{}).
		Where("executed_at >= ?", day30).
		Select("COALESCE(SUM(pnl), 0)").Scan(&pnl30d).Error; err != nil {
		return nil, fmt.Errorf("sum pnl 30d: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{}).
		Where("executed_at >= ?", day30).Count(&count30d).Error; err != nil {
		return nil, fmt.Errorf("count trading_trades 30d: %w", err)
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

func (s *tradingTradeService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
