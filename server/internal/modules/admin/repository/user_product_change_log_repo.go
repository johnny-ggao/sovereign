package repository

import (
	"context"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"gorm.io/gorm"
)

type UserProductChangeLogRepository interface {
	Create(ctx context.Context, l *model.UserProductChangeLog) error
	ListByUser(ctx context.Context, userID string) ([]model.UserProductChangeLog, error)
}

type userProductChangeLogRepository struct {
	db *gorm.DB
}

func NewUserProductChangeLogRepository(db *gorm.DB) UserProductChangeLogRepository {
	return &userProductChangeLogRepository{db: db}
}

func (r *userProductChangeLogRepository) Create(ctx context.Context, l *model.UserProductChangeLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *userProductChangeLogRepository) ListByUser(ctx context.Context, userID string) ([]model.UserProductChangeLog, error) {
	var rows []model.UserProductChangeLog
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}
