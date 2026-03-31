package dto

import "github.com/shopspring/decimal"

type SettlementResponse struct {
	ID             string          `json:"id"`
	InvestmentID   string          `json:"investment_id"`
	Period         string          `json:"period"`
	GrossReturn    decimal.Decimal `json:"gross_return"`
	PerformanceFee decimal.Decimal `json:"performance_fee"`
	FeeRate        decimal.Decimal `json:"fee_rate"`
	NetReturn      decimal.Decimal `json:"net_return"`
	TradeCount     int             `json:"trade_count"`
	AvgPremiumPct  decimal.Decimal `json:"avg_premium_pct"`
	ReportURL      string          `json:"report_url"`
	SettledAt      string          `json:"settled_at"`
}

type SettlementListResponse struct {
	Settlements []SettlementResponse `json:"settlements"`
	Summary     SettlementSummary    `json:"summary"`
}

type SettlementSummary struct {
	TotalGrossReturn    decimal.Decimal `json:"total_gross_return"`
	TotalPerformanceFee decimal.Decimal `json:"total_performance_fee"`
	TotalNetReturn      decimal.Decimal `json:"total_net_return"`
	PeriodCount         int             `json:"period_count"`
}
