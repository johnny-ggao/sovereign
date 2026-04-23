package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	settlemodel "github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	trademodel "github.com/sovereign-fund/sovereign/internal/modules/tradelog/model"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TradeService interface {
	List(ctx context.Context, query dto.TradeListQuery) ([]dto.TradeListItem, int64, error)
	Stats(ctx context.Context) (*dto.TradeStats, error)
	DownloadTemplate(ctx context.Context) (*excelize.File, error)
	ImportFromExcel(ctx context.Context, file multipart.File) (int, []string, error)
	Delete(ctx context.Context, tradeID string) error
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
			Source:       t.Source,
			ExecutedAt:   t.ExecutedAt.Format(time.RFC3339),
		}
	}

	return items, total, nil
}

func (s *tradeService) DownloadTemplate(_ context.Context) (*excelize.File, error) {
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

func (s *tradeService) ImportFromExcel(ctx context.Context, file multipart.File) (int, []string, error) {
	trades, rowErrors, err := parseImportRows(file)
	if err != nil {
		return 0, nil, err
	}
	if len(trades) == 0 {
		return 0, rowErrors, nil
	}
	if err := s.db.WithContext(ctx).CreateInBatches(trades, tradeImportBatchSize).Error; err != nil {
		return 0, rowErrors, fmt.Errorf("import trades: %w", err)
	}
	return len(trades), rowErrors, nil
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

func (s *tradeService) Delete(ctx context.Context, tradeID string) error {
	var trade trademodel.Trade
	if err := s.db.WithContext(ctx).Where("id = ?", tradeID).First(&trade).Error; err != nil {
		return fmt.Errorf("交易记录不存在")
	}

	if trade.Source != tradeSourceImport {
		return fmt.Errorf("只能删除导入的交易记录")
	}

	var count int64
	s.db.WithContext(ctx).Table("user_trades").Where("trade_id = ?", tradeID).Count(&count)
	if count > 0 {
		return fmt.Errorf("该交易已参与结算分润，无法删除")
	}

	if err := s.db.WithContext(ctx).Where("id = ?", tradeID).Delete(&trademodel.Trade{}).Error; err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}

	return nil
}

func parseImportRows(reader io.Reader) ([]trademodel.Trade, []string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("read excel file: %w", err)
	}
	workbook, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, nil, fmt.Errorf("open excel file: %w", err)
	}
	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("excel file contains no worksheets")
	}
	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		return nil, nil, fmt.Errorf("read excel rows: %w", err)
	}

	trades := make([]trademodel.Trade, 0, len(rows))
	rowErrors := make([]string, 0)
	for index, row := range rows[1:] {
		trade, rowErr := parseTradeRow(row)
		if rowErr != nil {
			rowErrors = append(rowErrors, fmt.Sprintf("第%d行: %s", index+2, rowErr.Error()))
			continue
		}
		trades = append(trades, *trade)
	}
	return trades, rowErrors, nil
}

func parseTradeRow(row []string) (*trademodel.Trade, error) {
	pair, err := requiredCell(row, 0, "交易对")
	if err != nil {
		return nil, err
	}
	buyExchange, err := requiredCell(row, 1, "买入交易所")
	if err != nil {
		return nil, err
	}
	sellExchange, err := requiredCell(row, 2, "卖出交易所")
	if err != nil {
		return nil, err
	}

	buyPrice, err := parseDecimalCell(row, 3, "买入价格", true)
	if err != nil {
		return nil, err
	}
	sellPrice, err := parseDecimalCell(row, 4, "卖出价格", true)
	if err != nil {
		return nil, err
	}
	amount, err := parseDecimalCell(row, 5, "金额", true)
	if err != nil {
		return nil, err
	}
	premiumPct, err := parseDecimalCell(row, 6, "溢价率(%)", true)
	if err != nil {
		return nil, err
	}
	pnl, err := parseDecimalCell(row, 7, "盈亏", true)
	if err != nil {
		return nil, err
	}
	fee, err := parseDecimalCell(row, 8, "手续费", false)
	if err != nil {
		return nil, err
	}

	return &trademodel.Trade{
		Pair:         pair,
		BuyExchange:  buyExchange,
		SellExchange: sellExchange,
		BuyPrice:     buyPrice,
		SellPrice:    sellPrice,
		Amount:       amount,
		PremiumPct:   premiumPct,
		PnL:          pnl,
		Fee:          fee,
		Source:       tradeSourceImport,
		ExecutedAt:   time.Now(),
	}, nil
}

func requiredCell(row []string, index int, label string) (string, error) {
	value := strings.TrimSpace(cellValue(row, index))
	if value == "" {
		return "", fmt.Errorf("%s不能为空", label)
	}
	return value, nil
}

func parseDecimalCell(row []string, index int, label string, required bool) (decimal.Decimal, error) {
	value := strings.TrimSpace(cellValue(row, index))
	if value == "" {
		if required {
			return decimal.Zero, fmt.Errorf("%s不能为空", label)
		}
		return decimal.Zero, nil
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%s格式无效", label)
	}
	return parsed, nil
}

func cellValue(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return row[index]
}

type tradeTemplateWidth struct {
	Start string
	End   string
	Width float64
}

var tradeTemplateHeaders = []string{
	"交易对", "买入交易所", "卖出交易所", "买入价格", "卖出价格",
	"金额", "溢价率(%)", "盈亏", "手续费",
}

var tradeTemplateSampleRow = []string{
	"USDT/KRW", "Binance", "Upbit", "1.0000", "1.0350",
	"10000.00", "3.50", "350.00", "10.00",
}

var tradeTemplateColumns = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}

var tradeTemplateWidths = []tradeTemplateWidth{
	{Start: "A", End: "A", Width: 18},
	{Start: "B", End: "C", Width: 16},
	{Start: "D", End: "I", Width: 14},
}

const (
	tradeTemplateSheetName = "Trades"
	tradeImportBatchSize   = 100
	tradeSourceImport      = "import"
)
