package repository

import (
	"context"

	"github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"gorm.io/gorm"
)

type AddressRepository interface {
	FindDepositAddress(ctx context.Context, userID, currency, network string) (*model.DepositAddress, error)
	FindDepositAddressByAddress(ctx context.Context, address string) (*model.DepositAddress, error)
	CreateDepositAddress(ctx context.Context, addr *model.DepositAddress) error

	FindWithdrawAddresses(ctx context.Context, userID string) ([]model.WithdrawAddress, error)
	FindWithdrawAddress(ctx context.Context, userID, address, network string) (*model.WithdrawAddress, error)
	CreateWithdrawAddress(ctx context.Context, addr *model.WithdrawAddress) error
	DeleteWithdrawAddress(ctx context.Context, id, userID string) error
}

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) FindDepositAddress(ctx context.Context, userID, currency, network string) (*model.DepositAddress, error) {
	var addr model.DepositAddress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND currency = ? AND network = ?", userID, currency, network).
		First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *addressRepository) FindDepositAddressByAddress(ctx context.Context, address string) (*model.DepositAddress, error) {
	var addr model.DepositAddress
	err := r.db.WithContext(ctx).
		Where("address = ?", address).
		First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *addressRepository) CreateDepositAddress(ctx context.Context, addr *model.DepositAddress) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

func (r *addressRepository) FindWithdrawAddresses(ctx context.Context, userID string) ([]model.WithdrawAddress, error) {
	var addrs []model.WithdrawAddress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true", userID).
		Order("created_at DESC").
		Find(&addrs).Error
	return addrs, err
}

func (r *addressRepository) FindWithdrawAddress(ctx context.Context, userID, address, network string) (*model.WithdrawAddress, error) {
	var addr model.WithdrawAddress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND address = ? AND network = ? AND is_active = true", userID, address, network).
		First(&addr).Error
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func (r *addressRepository) CreateWithdrawAddress(ctx context.Context, addr *model.WithdrawAddress) error {
	return r.db.WithContext(ctx).Create(addr).Error
}

func (r *addressRepository) DeleteWithdrawAddress(ctx context.Context, id, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.WithdrawAddress{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_active", false).Error
}
