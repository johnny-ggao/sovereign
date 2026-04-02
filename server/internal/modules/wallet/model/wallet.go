package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Wallet struct {
	ID          string          `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string          `gorm:"type:uuid;index;not null" json:"user_id"`
	Currency    string          `gorm:"type:varchar(10);not null" json:"currency"`
	Available   decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"available"`
	InOperation decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"in_operation"`
	Frozen      decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"frozen"`
	Earnings    decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"earnings"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (w *Wallet) BeforeCreate(_ *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

func (w *Wallet) TotalBalance() decimal.Decimal {
	return w.Available.Add(w.InOperation).Add(w.Frozen).Add(w.Earnings)
}
