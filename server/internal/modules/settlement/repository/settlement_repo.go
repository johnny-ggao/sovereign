package repository

import (
	"context"

	"github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	"gorm.io/gorm"
)

type SettlementRepository interface {
	Create(ctx context.Context, s *model.Settlement) error
	FindByUserID(ctx context.Context, userID string) ([]model.Settlement, error)
	FindByUserIDAndPeriodRange(ctx context.Context, userID, fromPeriod, toPeriod string) ([]model.Settlement, error)
	FindByID(ctx context.Context, id string) (*model.Settlement, error)
	FindByInvestmentAndPeriod(ctx context.Context, investmentID, period string) (*model.Settlement, error)
}

type settlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) SettlementRepository {
	return &settlementRepository{db: db}
}

func (r *settlementRepository) Create(ctx context.Context, s *model.Settlement) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *settlementRepository) FindByUserID(ctx context.Context, userID string) ([]model.Settlement, error) {
	var settlements []model.Settlement
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("period DESC").
		Find(&settlements).Error
	return settlements, err
}

func (r *settlementRepository) FindByUserIDAndPeriodRange(ctx context.Context, userID, fromPeriod, toPeriod string) ([]model.Settlement, error) {
	var settlements []model.Settlement
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND period >= ? AND period <= ?", userID, fromPeriod, toPeriod).
		Order("period ASC").
		Find(&settlements).Error
	return settlements, err
}

func (r *settlementRepository) FindByID(ctx context.Context, id string) (*model.Settlement, error) {
	var s model.Settlement
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settlementRepository) FindByInvestmentAndPeriod(ctx context.Context, investmentID, period string) (*model.Settlement, error) {
	var s model.Settlement
	err := r.db.WithContext(ctx).
		Where("investment_id = ? AND period = ?", investmentID, period).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}
