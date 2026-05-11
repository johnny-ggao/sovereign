package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProductChangeLog struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     string    `gorm:"type:uuid;index;not null" json:"user_id"`
	FromType   string    `gorm:"type:varchar(20);not null" json:"from_type"`
	ToType     string    `gorm:"type:varchar(20);not null" json:"to_type"`
	AdminID    string    `gorm:"type:uuid;not null" json:"admin_id"`
	AdminEmail string    `gorm:"type:varchar(255);not null" json:"admin_email"`
	Reason     string    `gorm:"type:varchar(500);not null" json:"reason"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (l *UserProductChangeLog) BeforeCreate(_ *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

func (UserProductChangeLog) TableName() string {
	return "user_product_change_logs"
}
