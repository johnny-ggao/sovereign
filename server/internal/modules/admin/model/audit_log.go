package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	AdminID    string    `gorm:"type:uuid;not null;index" json:"admin_id"`
	AdminEmail string    `gorm:"type:varchar(255);not null" json:"admin_email"`
	Action     string    `gorm:"type:varchar(50);not null;index" json:"action"`
	TargetType string    `gorm:"type:varchar(50);not null" json:"target_type"`
	TargetID   string    `gorm:"type:varchar(255)" json:"target_id"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IPAddress  string    `gorm:"type:varchar(50)" json:"ip_address"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func (AuditLog) TableName() string {
	return "admin_audit_logs"
}
