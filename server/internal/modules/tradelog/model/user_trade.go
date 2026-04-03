package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserTrade struct {
	ID           string          `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       string          `gorm:"type:uuid;index;not null" json:"user_id"`
	InvestmentID string          `gorm:"type:uuid;index;not null" json:"investment_id"`
	TradeID      *string         `gorm:"type:uuid" json:"trade_id"`
	SettlementID *string         `gorm:"type:uuid" json:"settlement_id"`
	Pair         string          `gorm:"type:varchar(20);not null" json:"pair"`
	BuyExchange  string          `gorm:"type:varchar(20);not null" json:"buy_exchange"`
	SellExchange string          `gorm:"type:varchar(20);not null" json:"sell_exchange"`
	BuyPrice     decimal.Decimal `gorm:"type:decimal(38,8);not null" json:"buy_price"`
	SellPrice    decimal.Decimal `gorm:"type:decimal(38,8);not null" json:"sell_price"`
	Amount       decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"amount"`
	PremiumPct   decimal.Decimal `gorm:"type:decimal(8,4);not null" json:"premium_pct"`
	PnL          decimal.Decimal `gorm:"column:pnl;type:decimal(28,18);not null" json:"pnl"`
	Fee          decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"fee"`
	Ratio        decimal.Decimal `gorm:"type:decimal(10,8);not null" json:"ratio"`
	ExecutedAt   time.Time       `gorm:"not null" json:"executed_at"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (t *UserTrade) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

func (UserTrade) TableName() string {
	return "user_trades"
}
