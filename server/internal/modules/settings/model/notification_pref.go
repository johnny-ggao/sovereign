package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationPref struct {
	ID               string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID           string    `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	EmailTrade       bool      `gorm:"default:true" json:"email_trade"`
	EmailDeposit     bool      `gorm:"default:true" json:"email_deposit"`
	EmailWithdraw    bool      `gorm:"default:true" json:"email_withdraw"`
	EmailSettlement  bool      `gorm:"default:true" json:"email_settlement"`
	PushPremiumAlert bool      `gorm:"default:false" json:"push_premium_alert"`
	PushTrade        bool      `gorm:"default:true" json:"push_trade"`
	PushDeposit      bool      `gorm:"default:true" json:"push_deposit"`
	PushWithdraw     bool      `gorm:"default:true" json:"push_withdraw"`
	PremiumThreshold float64   `gorm:"type:decimal(5,2);default:3.0" json:"premium_threshold"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (n *NotificationPref) BeforeCreate(_ *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
