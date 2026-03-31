package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Investment struct {
	ID             string          `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         string          `gorm:"type:uuid;index;not null" json:"user_id"`
	Amount         decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"amount"`
	Currency       string          `gorm:"type:varchar(10);not null;default:USDT" json:"currency"`
	Status         string          `gorm:"type:varchar(20);default:active" json:"status"`
	TotalReturn    decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"total_return"`
	PerformanceFee decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"performance_fee"`
	NetReturn      decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"net_return"`
	StartDate      time.Time       `gorm:"not null" json:"start_date"`
	EndDate        *time.Time      `json:"end_date"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (i *Investment) BeforeCreate(_ *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	if i.StartDate.IsZero() {
		i.StartDate = time.Now()
	}
	return nil
}

const (
	InvestStatusActive   = "active"
	InvestStatusStopping = "stopping"
	InvestStatusRedeemed = "redeemed"
)
