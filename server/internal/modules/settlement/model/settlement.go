package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Settlement struct {
	ID             string          `gorm:"type:uuid;primaryKey" json:"id"`
	InvestmentID   string          `gorm:"type:uuid;index;not null" json:"investment_id"`
	UserID         string          `gorm:"type:uuid;index;not null" json:"user_id"`
	Period         string          `gorm:"type:varchar(10);not null" json:"period"` // YYYY-MM-DD
	GrossReturn    decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"gross_return"`
	PerformanceFee decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"performance_fee"`
	FeeRate        decimal.Decimal `gorm:"type:decimal(5,4);not null;default:0.5" json:"fee_rate"`
	NetReturn      decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"net_return"`
	TradeCount     int             `gorm:"not null;default:0" json:"trade_count"`
	AvgPremiumPct  decimal.Decimal `gorm:"type:decimal(8,4);default:0" json:"avg_premium_pct"`
	ReportURL      string          `gorm:"type:varchar(500)" json:"report_url"`
	SettledAt      time.Time       `gorm:"not null" json:"settled_at"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (s *Settlement) BeforeCreate(_ *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
