package repository

import (
	"context"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"gorm.io/gorm"
)

type AdminRepository interface {
	FindByID(ctx context.Context, id string) (*model.AdminUser, error)
	FindByEmail(ctx context.Context, email string) (*model.AdminUser, error)
	FindAll(ctx context.Context) ([]model.AdminUser, error)
	Create(ctx context.Context, admin *model.AdminUser) error
	Update(ctx context.Context, admin *model.AdminUser) error
	Delete(ctx context.Context, id string) error
	UpdateLastLogin(ctx context.Context, id string) error
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) FindByID(ctx context.Context, id string) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) FindAll(ctx context.Context) ([]model.AdminUser, error) {
	var admins []model.AdminUser
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&admins).Error
	return admins, err
}

func (r *adminRepository) Create(ctx context.Context, admin *model.AdminUser) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *adminRepository) Update(ctx context.Context, admin *model.AdminUser) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *adminRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AdminUser{}).Error
}

func (r *adminRepository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Update("last_login", now).Error
}
