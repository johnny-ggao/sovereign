package repository

import (
	"context"

	"github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	"gorm.io/gorm"
)

type InvestmentRepository interface {
	Create(ctx context.Context, inv *model.Investment) error
	FindByID(ctx context.Context, id string) (*model.Investment, error)
	FindByUserID(ctx context.Context, userID string) ([]model.Investment, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]model.Investment, error)
	FindAllActive(ctx context.Context) ([]model.Investment, error)
	Update(ctx context.Context, inv *model.Investment) error
}

type investmentRepository struct {
	db *gorm.DB
}

func NewInvestmentRepository(db *gorm.DB) InvestmentRepository {
	return &investmentRepository{db: db}
}

func (r *investmentRepository) Create(ctx context.Context, inv *model.Investment) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *investmentRepository) FindByID(ctx context.Context, id string) (*model.Investment, error) {
	var inv model.Investment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *investmentRepository) FindByUserID(ctx context.Context, userID string) ([]model.Investment, error) {
	var invs []model.Investment
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&invs).Error
	return invs, err
}

func (r *investmentRepository) FindActiveByUserID(ctx context.Context, userID string) ([]model.Investment, error) {
	var invs []model.Investment
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.InvestStatusActive).
		Order("created_at DESC").Find(&invs).Error
	return invs, err
}

func (r *investmentRepository) FindAllActive(ctx context.Context) ([]model.Investment, error) {
	var invs []model.Investment
	err := r.db.WithContext(ctx).
		Where("status = ?", model.InvestStatusActive).
		Find(&invs).Error
	return invs, err
}

func (r *investmentRepository) Update(ctx context.Context, inv *model.Investment) error {
	return r.db.WithContext(ctx).Save(inv).Error
}
