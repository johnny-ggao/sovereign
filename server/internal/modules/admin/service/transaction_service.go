package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	txmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"gorm.io/gorm"
)

type TransactionService interface {
	List(ctx context.Context, query dto.TransactionListQuery) ([]dto.TransactionListItem, int64, error)
	Stats(ctx context.Context) (*dto.TransactionStats, error)
}

type transactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) TransactionService {
	return &transactionService{db: db}
}

func (s *transactionService) List(ctx context.Context, query dto.TransactionListQuery) ([]dto.TransactionListItem, int64, error) {
	type txRow struct {
		ID        string          `gorm:"column:id"`
		UserID    string          `gorm:"column:user_id"`
		UserEmail string          `gorm:"column:user_email"`
		Type      string          `gorm:"column:type"`
		Currency  string          `gorm:"column:currency"`
		Network   string          `gorm:"column:network"`
		Amount    decimal.Decimal `gorm:"column:amount"`
		Fee       decimal.Decimal `gorm:"column:fee"`
		Address   string          `gorm:"column:address"`
		TxHash    string          `gorm:"column:tx_hash"`
		Status    string          `gorm:"column:status"`
		CreatedAt time.Time       `gorm:"column:created_at"`
	}

	db := s.db.WithContext(ctx).Table("transactions").
		Select("transactions.id, transactions.user_id, users.email as user_email, transactions.type, transactions.currency, transactions.network, transactions.amount, transactions.fee, transactions.address, transactions.tx_hash, transactions.status, transactions.created_at").
		Joins("LEFT JOIN users ON users.id = transactions.user_id")

	if query.Type != "" {
		db = db.Where("transactions.type = ?", query.Type)
	}
	if query.Status != "" {
		db = db.Where("transactions.status = ?", query.Status)
	}
	if query.Search != "" {
		pattern := "%" + query.Search + "%"
		db = db.Where("users.email ILIKE ? OR transactions.address ILIKE ? OR transactions.tx_hash ILIKE ?", pattern, pattern, pattern)
	}
	if query.DateFrom != "" {
		db = db.Where("transactions.created_at >= ?", query.DateFrom)
	}
	if query.DateTo != "" {
		db = db.Where("transactions.created_at < ?", query.DateTo+" 23:59:59")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	offset := (query.Page - 1) * query.Limit
	var rows []txRow
	if err := db.Order("transactions.created_at DESC").Offset(offset).Limit(query.Limit).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("find transactions: %w", err)
	}

	items := make([]dto.TransactionListItem, len(rows))
	for i, r := range rows {
		items[i] = dto.TransactionListItem{
			ID:        r.ID,
			UserID:    r.UserID,
			UserEmail: r.UserEmail,
			Type:      r.Type,
			Currency:  r.Currency,
			Network:   r.Network,
			Amount:    r.Amount.StringFixed(2),
			Fee:       r.Fee.StringFixed(2),
			Address:   r.Address,
			TxHash:    r.TxHash,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		}
	}
	return items, total, nil
}

func (s *transactionService) Stats(ctx context.Context) (*dto.TransactionStats, error) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	day7 := today.AddDate(0, 0, -7)
	day30 := today.AddDate(0, 0, -30)

	sumAndCount := func(txType string, since time.Time) (decimal.Decimal, int64, error) {
		var sum decimal.Decimal
		var count int64
		if err := s.db.WithContext(ctx).Model(&txmodel.Transaction{}).
			Where("type = ? AND status = ? AND created_at >= ?", txType, txmodel.TxStatusConfirmed, since).
			Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error; err != nil {
			return decimal.Zero, 0, fmt.Errorf("sum %s since %v: %w", txType, since, err)
		}
		if err := s.db.WithContext(ctx).Model(&txmodel.Transaction{}).
			Where("type = ? AND status = ? AND created_at >= ?", txType, txmodel.TxStatusConfirmed, since).
			Count(&count).Error; err != nil {
			return decimal.Zero, 0, fmt.Errorf("count %s since %v: %w", txType, since, err)
		}
		return sum, count, nil
	}

	d1d, dc1d, err := sumAndCount(txmodel.TxTypeDeposit, today)
	if err != nil {
		return nil, err
	}
	d7d, dc7d, err := sumAndCount(txmodel.TxTypeDeposit, day7)
	if err != nil {
		return nil, err
	}
	d30d, dc30d, err := sumAndCount(txmodel.TxTypeDeposit, day30)
	if err != nil {
		return nil, err
	}
	w1d, wc1d, err := sumAndCount(txmodel.TxTypeWithdraw, today)
	if err != nil {
		return nil, err
	}
	w7d, wc7d, err := sumAndCount(txmodel.TxTypeWithdraw, day7)
	if err != nil {
		return nil, err
	}
	w30d, wc30d, err := sumAndCount(txmodel.TxTypeWithdraw, day30)
	if err != nil {
		return nil, err
	}

	return &dto.TransactionStats{
		Deposit1D:        d1d.StringFixed(2),
		Deposit7D:        d7d.StringFixed(2),
		Deposit30D:       d30d.StringFixed(2),
		DepositCount1D:   dc1d,
		DepositCount7D:   dc7d,
		DepositCount30D:  dc30d,
		Withdraw1D:       w1d.StringFixed(2),
		Withdraw7D:       w7d.StringFixed(2),
		Withdraw30D:      w30d.StringFixed(2),
		WithdrawCount1D:  wc1d,
		WithdrawCount7D:  wc7d,
		WithdrawCount30D: wc30d,
	}, nil
}
