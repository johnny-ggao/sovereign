package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	Email         string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"type:varchar(255)" json:"-"`
	GoogleID      string    `gorm:"type:varchar(255);index" json:"-"`
	AvatarURL     string    `gorm:"type:varchar(500)" json:"avatar_url"`
	FullName      string    `gorm:"type:varchar(255)" json:"full_name"`
	Phone         string    `gorm:"type:varchar(50)" json:"phone"`
	Language      string    `gorm:"type:varchar(5);default:ko" json:"language"`
	KYCStatus     string    `gorm:"type:varchar(20);default:pending" json:"kyc_status"`
	TwoFASecret   string    `gorm:"type:text" json:"-"`
	TwoFAEnabled  bool      `gorm:"default:false" json:"two_fa_enabled"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

const (
	KYCStatusPending   = "pending"
	KYCStatusSubmitted = "submitted"
	KYCStatusApproved  = "approved"
	KYCStatusRejected  = "rejected"
)
