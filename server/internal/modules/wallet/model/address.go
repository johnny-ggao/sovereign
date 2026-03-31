package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepositAddress struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;index;not null" json:"user_id"`
	Currency  string    `gorm:"type:varchar(10);not null" json:"currency"`
	Network   string    `gorm:"type:varchar(20);not null" json:"network"`
	Address   string    `gorm:"type:varchar(255);not null" json:"address"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (d *DepositAddress) BeforeCreate(_ *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

type WithdrawAddress struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        string    `gorm:"type:uuid;index;not null" json:"user_id"`
	Currency      string    `gorm:"type:varchar(10);not null" json:"currency"`
	Network       string    `gorm:"type:varchar(20);not null" json:"network"`
	Address       string    `gorm:"type:varchar(255);not null" json:"address"`
	Label         string    `gorm:"type:varchar(100)" json:"label"`
	CooldownUntil time.Time `gorm:"not null" json:"cooldown_until"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (w *WithdrawAddress) BeforeCreate(_ *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

func (w *WithdrawAddress) InCooldown() bool {
	return time.Now().Before(w.CooldownUntil)
}

const (
	NetworkERC20 = "ERC20"
	NetworkTRC20 = "TRC20"
	NetworkBEP20 = "BEP20"
)
