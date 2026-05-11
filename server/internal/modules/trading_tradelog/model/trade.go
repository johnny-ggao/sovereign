package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TradingTrade struct {
	ID           string          `gorm:"type:uuid;primaryKey" json:"id"`
	Pair         string          `gorm:"type:varchar(20);not null" json:"pair"`
	BuyExchange  string          `gorm:"type:varchar(20)" json:"buy_exchange"`
	SellExchange string          `gorm:"type:varchar(20)" json:"sell_exchange"`
	BuyPrice     decimal.Decimal `gorm:"type:decimal(28,8)" json:"buy_price"`
	SellPrice    decimal.Decimal `gorm:"type:decimal(28,8)" json:"sell_price"`
	Amount       decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"amount"`
	PremiumPct   decimal.Decimal `gorm:"type:decimal(8,4);default:0" json:"premium_pct"`
	PnL          decimal.Decimal `gorm:"column:pnl;type:decimal(28,18);not null" json:"pnl"`
	Fee          decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"fee"`
	Source       string          `gorm:"type:varchar(20);default:import" json:"source"`
	ExecutedAt   time.Time       `gorm:"not null" json:"executed_at"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (t *TradingTrade) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

func (TradingTrade) TableName() string {
	return "trading_trades"
}
