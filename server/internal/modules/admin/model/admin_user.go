package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleSuperAdmin = "super_admin"
	RoleOperator   = "operator"
	RoleViewer     = "viewer"
)

type AdminUser struct {
	ID                 string     `gorm:"type:uuid;primaryKey" json:"id"`
	Email              string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash       string     `gorm:"type:varchar(255);not null" json:"-"`
	Name               string     `gorm:"type:varchar(255);not null" json:"name"`
	Role               string     `gorm:"type:varchar(20);not null;default:viewer" json:"role"`
	IsActive           bool       `gorm:"default:true" json:"is_active"`
	MustChangePassword bool       `gorm:"default:true" json:"must_change_password"`
	LastLogin          *time.Time `json:"last_login"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (a *AdminUser) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func IsValidRole(role string) bool {
	return role == RoleSuperAdmin || role == RoleOperator || role == RoleViewer
}
