package dto

import "github.com/shopspring/decimal"

type TradeResponse struct {
	ID           string          `json:"id"`
	InvestmentID string          `json:"investment_id"`
	Pair         string          `json:"pair"`
	BuyExchange  string          `json:"buy_exchange"`
	SellExchange string          `json:"sell_exchange"`
	BuyPrice     decimal.Decimal `json:"buy_price"`
	SellPrice    decimal.Decimal `json:"sell_price"`
	Amount       decimal.Decimal `json:"amount"`
	PremiumPct   decimal.Decimal `json:"premium_pct"`
	PnL          decimal.Decimal `json:"pnl"`
	Fee          decimal.Decimal `json:"fee"`
	ExecutedAt   string          `json:"executed_at"`
}

type TradeListResponse struct {
	Trades  []TradeResponse `json:"trades"`
	Summary TradeSummary    `json:"summary"`
}

type TradeSummary struct {
	TotalTrades   int64           `json:"total_trades"`
	TotalPnL      decimal.Decimal `json:"total_pnl"`
	AvgPremiumPct decimal.Decimal `json:"avg_premium_pct"`
	WinRate       decimal.Decimal `json:"win_rate"`
}

type TradeFilterRequest struct {
	InvestmentID string `form:"investment_id" binding:"omitempty,uuid"`
	Pair         string `form:"pair" binding:"omitempty"`
	From         string `form:"from" binding:"omitempty"`
	To           string `form:"to" binding:"omitempty"`
}
