package dto

import "github.com/shopspring/decimal"

type InvestmentResponse struct {
	ID             string          `json:"id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Status         string          `json:"status"`
	TotalReturn    decimal.Decimal `json:"total_return"`
	PerformanceFee decimal.Decimal `json:"performance_fee"`
	NetReturn      decimal.Decimal `json:"net_return"`
	ReturnPct      decimal.Decimal `json:"return_pct"`
	StartDate      string          `json:"start_date"`
	EndDate        *string         `json:"end_date"`
}

type InvestmentListResponse struct {
	Investments []InvestmentResponse `json:"investments"`
	Summary     InvestmentSummary    `json:"summary"`
}

type InvestmentSummary struct {
	TotalInvested  decimal.Decimal `json:"total_invested"`
	TotalReturn    decimal.Decimal `json:"total_return"`
	ActiveCount    int             `json:"active_count"`
}
