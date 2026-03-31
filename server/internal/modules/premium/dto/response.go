package dto

import "github.com/shopspring/decimal"

type PremiumTickResponse struct {
	Pair              string           `json:"pair"`
	KoreanPrice       decimal.Decimal  `json:"korean_price"`
	GlobalPrice       decimal.Decimal  `json:"global_price"`
	PremiumPct        decimal.Decimal  `json:"premium_pct"`
	ReversePremiumPct decimal.Decimal  `json:"reverse_premium_pct"`
	SourceKR          string           `json:"source_kr"`
	SourceGL          string           `json:"source_gl"`
	Latencies         map[string]int64 `json:"latencies,omitempty"`
	Timestamp         string           `json:"timestamp"`
}

type PremiumLatestResponse struct {
	Ticks []PremiumTickResponse `json:"ticks"`
}

type PremiumHistoryResponse struct {
	Pair   string                `json:"pair"`
	Points []PremiumTickResponse `json:"points"`
}

type WSMessage struct {
	Type string      `json:"type"` // tick, error, subscribed
	Data interface{} `json:"data"`
}
