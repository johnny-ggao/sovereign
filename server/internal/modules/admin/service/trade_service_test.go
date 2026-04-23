package service

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func TestDownloadTemplateIncludesHeadersAndSampleRow(t *testing.T) {
	svc := &tradeService{}

	file, err := svc.DownloadTemplate(context.Background())
	if err != nil {
		t.Fatalf("DownloadTemplate() error = %v", err)
	}

	rows, err := file.GetRows(file.GetSheetList()[0])
	if err != nil {
		t.Fatalf("GetRows() error = %v", err)
	}

	if len(rows) < 2 {
		t.Fatalf("expected at least 2 rows, got %d", len(rows))
	}

	if rows[0][0] != "交易对" || rows[1][0] != "USDT/KRW" {
		t.Fatalf("unexpected template rows: %#v", rows[:2])
	}
}

func TestParseImportRowsReturnsTradesAndRowErrors(t *testing.T) {
	workbook := newImportWorkbook([][]string{
		{"交易对", "买入交易所", "卖出交易所", "买入价格", "卖出价格", "金额", "溢价率(%)", "盈亏", "手续费"},
		{"USDT/KRW", "Binance", "Upbit", "1.0000", "1.0350", "10000.00", "3.50", "350.00", "10.00"},
		{"BTC/USDT", "", "Upbit", "80000", "81000", "1.00", "1.25", "1000", "5"},
	})

	trades, rowErrors, err := parseImportRows(bytes.NewReader(workbook.Bytes()))
	if err != nil {
		t.Fatalf("parseImportRows() error = %v", err)
	}

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if len(rowErrors) != 1 || !strings.Contains(rowErrors[0], "第3行") {
		t.Fatalf("unexpected row errors: %#v", rowErrors)
	}
	if trades[0].Source != "import" {
		t.Fatalf("expected source=import, got %q", trades[0].Source)
	}
	if !trades[0].Amount.Equal(decimal.RequireFromString("10000.00")) {
		t.Fatalf("unexpected amount: %s", trades[0].Amount.String())
	}
}

func newImportWorkbook(rows [][]string) *bytes.Buffer {
	file := excelize.NewFile()
	sheet := file.GetSheetList()[0]
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell := string(rune('A'+colIndex)) + strconv.Itoa(rowIndex+1)
			_ = file.SetCellValue(sheet, cell, value)
		}
	}
	buffer := &bytes.Buffer{}
	_ = file.Write(buffer)
	return buffer
}
