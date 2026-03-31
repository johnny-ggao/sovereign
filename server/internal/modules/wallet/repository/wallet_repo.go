package repository

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]model.Wallet, error)
	FindByUserIDAndCurrency(ctx context.Context, userID, currency string) (*model.Wallet, error)
	FindOrCreate(ctx context.Context, userID, currency string) (*model.Wallet, error)
	UpdateBalance(ctx context.Context, id string, available, inOperation, frozen decimal.Decimal) error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) FindByUserID(ctx context.Context, userID string) ([]model.Wallet, error) {
	var wallets []model.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) FindByUserIDAndCurrency(ctx context.Context, userID, currency string) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ? AND currency = ?", userID, currency).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) FindOrCreate(ctx context.Context, userID, currency string) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.WithContext(ctx).
		Where(model.Wallet{UserID: userID, Currency: currency}).
		Attrs(model.Wallet{
			Available:   decimal.Zero,
			InOperation: decimal.Zero,
			Frozen:      decimal.Zero,
		}).
		Clauses(clause.OnConflict{DoNothing: true}).
		FirstOrCreate(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, id string, available, inOperation, frozen decimal.Decimal) error {
	return r.db.WithContext(ctx).
		Model(&model.Wallet{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"available":    available,
			"in_operation": inOperation,
			"frozen":       frozen,
		}).Error
}
