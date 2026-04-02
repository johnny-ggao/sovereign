package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Trade struct {
	ID           string          `gorm:"type:uuid;primaryKey" json:"id"`
	InvestmentID string          `gorm:"type:uuid;index;default:''" json:"investment_id"`
	UserID       string          `gorm:"type:uuid;index;default:''" json:"user_id"`
	Pair         string          `gorm:"type:varchar(20);not null" json:"pair"`
	BuyExchange  string          `gorm:"type:varchar(20);not null" json:"buy_exchange"`
	SellExchange string          `gorm:"type:varchar(20);not null" json:"sell_exchange"`
	BuyPrice     decimal.Decimal `gorm:"type:decimal(28,8);not null" json:"buy_price"`
	SellPrice    decimal.Decimal `gorm:"type:decimal(28,8);not null" json:"sell_price"`
	Amount       decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"amount"`
	PremiumPct   decimal.Decimal `gorm:"type:decimal(8,4);not null" json:"premium_pct"`
	PnL          decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"pnl"`
	Fee          decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"fee"`
	ExecutedAt   time.Time       `gorm:"not null" json:"executed_at"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (t *Trade) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
