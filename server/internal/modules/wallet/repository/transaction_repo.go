package repository

import (
	"context"

	"github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	FindByID(ctx context.Context, id string) (*model.Transaction, error)
	FindByExternalID(ctx context.Context, externalID string) (*model.Transaction, error)
	FindByUserID(ctx context.Context, userID string, txType string, limit, offset int) ([]model.Transaction, int64, error)
	UpdateStatus(ctx context.Context, id, status, txHash string) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *transactionRepository) FindByID(ctx context.Context, id string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) FindByExternalID(ctx context.Context, externalID string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) FindByUserID(ctx context.Context, userID string, txType string, limit, offset int) ([]model.Transaction, int64, error) {
	var txs []model.Transaction
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Transaction{}).Where("user_id = ?", userID)
	if txType != "" {
		query = query.Where("type = ?", txType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&txs).Error
	return txs, total, err
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, id, status, txHash string) error {
	updates := map[string]any{"status": status}
	if txHash != "" {
		updates["tx_hash"] = txHash
	}
	if status == model.TxStatusConfirmed {
		updates["confirmed_at"] = gorm.Expr("NOW()")
	}
	return r.db.WithContext(ctx).
		Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(updates).Error
}
