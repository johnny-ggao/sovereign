package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type PremiumTick struct {
	ID                uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Pair              string          `gorm:"type:varchar(20);not null;index" json:"pair"`
	KoreanPrice       decimal.Decimal `gorm:"type:decimal(28,8);not null" json:"korean_price"`
	GlobalPrice       decimal.Decimal `gorm:"type:decimal(28,8);not null" json:"global_price"`
	PremiumPct        decimal.Decimal `gorm:"type:decimal(8,4);not null" json:"premium_pct"`
	ReversePremiumPct decimal.Decimal `gorm:"type:decimal(8,4);not null;default:0" json:"reverse_premium_pct"`
	SourceKR          string          `gorm:"type:varchar(20)" json:"source_kr"`
	SourceGL          string          `gorm:"type:varchar(20)" json:"source_gl"`
	CreatedAt         time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
}

func (PremiumTick) TableName() string {
	return "premium_ticks"
}

type PremiumSnapshot struct {
	Pair              string            `json:"pair"`
	KoreanPrice       decimal.Decimal   `json:"korean_price"`
	GlobalPrice       decimal.Decimal   `json:"global_price"`
	PremiumPct        decimal.Decimal   `json:"premium_pct"`
	ReversePremiumPct decimal.Decimal   `json:"reverse_premium_pct"`
	SourceKR          string            `json:"source_kr"`
	SourceGL          string            `json:"source_gl"`
	Latencies         map[string]int64  `json:"latencies,omitempty"` // exchange -> ms
	Timestamp         time.Time         `json:"timestamp"`
}

const (
	PairBTCKRW  = "BTC/KRW"
	PairETHKRW  = "ETH/KRW"
	PairSOLKRW  = "SOL/KRW"
	PairXRPKRW  = "XRP/KRW"
	PairUSDTKRW = "USDT/KRW"

	ExchangeUpbit   = "upbit"
	ExchangeBithumb = "bithumb"
	ExchangeBinance = "binance"
	ExchangeBybit   = "bybit"
)
