package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginDevice struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;index;not null" json:"user_id"`
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"`
	IP        string    `gorm:"type:varchar(45)" json:"ip"`
	Location  string    `gorm:"type:varchar(255)" json:"location"`
	LastLogin time.Time `gorm:"not null" json:"last_login"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (d *LoginDevice) BeforeCreate(_ *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
