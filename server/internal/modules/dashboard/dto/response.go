package dto

import "github.com/shopspring/decimal"

type DashboardSummary struct {
	Portfolio   PortfolioSummary  `json:"portfolio"`
	Performance PerformanceData  `json:"performance"`
	Premium     PremiumSummary   `json:"premium"`
}

type PortfolioSummary struct {
	TotalValue     decimal.Decimal `json:"total_value"`
	TotalValueUSD  decimal.Decimal `json:"total_value_usd"`
	Available      decimal.Decimal `json:"available"`
	InOperation    decimal.Decimal `json:"in_operation"`
	Frozen         decimal.Decimal `json:"frozen"`
	Currency       string          `json:"currency"`
}

type PerformanceData struct {
	CumulativeReturn    decimal.Decimal      `json:"cumulative_return"`
	CumulativeReturnPct decimal.Decimal      `json:"cumulative_return_pct"`
	AnnualizedReturn    decimal.Decimal      `json:"annualized_return_pct"`
	HighWaterMark       decimal.Decimal      `json:"high_water_mark"`
	Chart               []PerformancePoint   `json:"chart"`
}

type PerformancePoint struct {
	Date  string          `json:"date"`
	Value decimal.Decimal `json:"value"`
}

type PremiumSummary struct {
	CurrentPct  decimal.Decimal `json:"current_pct"`
	Pair        string          `json:"pair"`
	Trend       string          `json:"trend"` // up, down, flat
	LastUpdated string          `json:"last_updated"`
}

type PerformanceRequest struct {
	Period string `form:"period" binding:"omitempty,oneof=1W 1M 3M 6M 1Y ALL"`
}
