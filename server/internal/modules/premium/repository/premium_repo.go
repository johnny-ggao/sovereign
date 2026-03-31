package repository

import (
	"context"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/premium/model"
	"gorm.io/gorm"
)

type PremiumRepository interface {
	Create(ctx context.Context, tick *model.PremiumTick) error
	CreateBatch(ctx context.Context, ticks []model.PremiumTick) error
	FindLatest(ctx context.Context) ([]model.PremiumTick, error)
	FindLatestByPair(ctx context.Context, pair string) (*model.PremiumTick, error)
	FindHistory(ctx context.Context, pair string, from, to time.Time, limit int) ([]model.PremiumTick, error)
	DeleteOlderThan(ctx context.Context, before time.Time) error
}

type premiumRepository struct {
	db *gorm.DB
}

func NewPremiumRepository(db *gorm.DB) PremiumRepository {
	return &premiumRepository{db: db}
}

func (r *premiumRepository) Create(ctx context.Context, tick *model.PremiumTick) error {
	return r.db.WithContext(ctx).Create(tick).Error
}

func (r *premiumRepository) CreateBatch(ctx context.Context, ticks []model.PremiumTick) error {
	return r.db.WithContext(ctx).Create(&ticks).Error
}

func (r *premiumRepository) FindLatest(ctx context.Context) ([]model.PremiumTick, error) {
	var ticks []model.PremiumTick

	subQuery := r.db.WithContext(ctx).
		Model(&model.PremiumTick{}).
		Select("pair, MAX(created_at) as max_created").
		Group("pair")

	err := r.db.WithContext(ctx).
		Joins("INNER JOIN (?) AS latest ON premium_ticks.pair = latest.pair AND premium_ticks.created_at = latest.max_created", subQuery).
		Find(&ticks).Error

	return ticks, err
}

func (r *premiumRepository) FindLatestByPair(ctx context.Context, pair string) (*model.PremiumTick, error) {
	var tick model.PremiumTick
	err := r.db.WithContext(ctx).
		Where("pair = ?", pair).
		Order("created_at DESC").
		First(&tick).Error
	if err != nil {
		return nil, err
	}
	return &tick, nil
}

func (r *premiumRepository) FindHistory(ctx context.Context, pair string, from, to time.Time, limit int) ([]model.PremiumTick, error) {
	var ticks []model.PremiumTick

	query := r.db.WithContext(ctx).Where("pair = ?", pair)

	if !from.IsZero() {
		query = query.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		query = query.Where("created_at <= ?", to)
	}

	if limit <= 0 || limit > 1000 {
		limit = 500
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&ticks).Error
	return ticks, err
}

func (r *premiumRepository) DeleteOlderThan(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&model.PremiumTick{}).Error
}
