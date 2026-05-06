package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	walletrepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"gorm.io/gorm"
)

// Narrow interfaces — only methods this service uses.
type walletReader interface {
	FindByUserIDAndCurrency(ctx context.Context, userID, currency string) (*walletmodel.Wallet, error)
	UpdateBalance(ctx context.Context, id string, available, inOperation, frozen decimal.Decimal) error
}

type txWriter interface {
	FindByID(ctx context.Context, id string) (*walletmodel.Transaction, error)
	UpdateReview(ctx context.Context, id string, fields map[string]any) error
	UpdateExternalID(ctx context.Context, id, externalID string) error
	UpdateStatus(ctx context.Context, id, status, txHash string) error
}

type reviewListRow struct {
	ID              string          `gorm:"column:id"`
	UserID          string          `gorm:"column:user_id"`
	UserEmail       string          `gorm:"column:user_email"`
	Currency        string          `gorm:"column:currency"`
	Network         string          `gorm:"column:network"`
	Amount          decimal.Decimal `gorm:"column:amount"`
	Address         string          `gorm:"column:address"`
	Status          string          `gorm:"column:status"`
	ReviewStatus    string          `gorm:"column:review_status"`
	RejectReason    string          `gorm:"column:reject_reason"`
	SubmitAttempts  int             `gorm:"column:submit_attempts"`
	LastSubmitError string          `gorm:"column:last_submit_error"`
	LastSubmitAt    *time.Time      `gorm:"column:last_submit_at"`
	ReviewedBy      string          `gorm:"column:reviewed_by"`
	ReviewedAt      *time.Time      `gorm:"column:reviewed_at"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	ExternalID      string          `gorm:"column:external_id"`
	TxHash          string          `gorm:"column:tx_hash"`
}

type reviewQuerier interface {
	Query(ctx context.Context, q dto.WithdrawReviewListQuery) ([]reviewListRow, int64, error)
}

type WithdrawReviewService interface {
	List(ctx context.Context, q dto.WithdrawReviewListQuery) ([]dto.WithdrawReviewItem, int64, error)
	Approve(ctx context.Context, txID, adminID string) error
	Reject(ctx context.Context, txID, adminID, reason string) error
	Retry(ctx context.Context, txID, adminID string) error
}

type withdrawReviewService struct {
	q        reviewQuerier
	walletR  walletReader
	txR      txWriter
	provider cobo.WalletProvider
	bus      events.Bus
	logger   *slog.Logger
}

func NewWithdrawReviewService(
	q reviewQuerier,
	walletR walletReader,
	txR txWriter,
	provider cobo.WalletProvider,
	bus events.Bus,
	logger *slog.Logger,
) WithdrawReviewService {
	return &withdrawReviewService{q: q, walletR: walletR, txR: txR, provider: provider, bus: bus, logger: logger}
}

func (s *withdrawReviewService) List(ctx context.Context, q dto.WithdrawReviewListQuery) ([]dto.WithdrawReviewItem, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	rows, total, err := s.q.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	items := make([]dto.WithdrawReviewItem, len(rows))
	for i, r := range rows {
		items[i] = dto.WithdrawReviewItem{
			ID: r.ID, UserID: r.UserID, UserEmail: r.UserEmail,
			Currency: r.Currency, Network: r.Network, Amount: r.Amount,
			Address: r.Address, Status: r.Status, ReviewStatus: r.ReviewStatus,
			RejectReason: r.RejectReason, SubmitAttempts: r.SubmitAttempts,
			LastSubmitError: r.LastSubmitError, ReviewedBy: r.ReviewedBy,
			CreatedAt:  r.CreatedAt.Format(time.RFC3339),
			ExternalID: r.ExternalID, TxHash: r.TxHash,
		}
		if r.LastSubmitAt != nil {
			t := r.LastSubmitAt.Format(time.RFC3339)
			items[i].LastSubmitAt = &t
		}
		if r.ReviewedAt != nil {
			t := r.ReviewedAt.Format(time.RFC3339)
			items[i].ReviewedAt = &t
		}
	}
	return items, total, nil
}

func (s *withdrawReviewService) Approve(ctx context.Context, txID, adminID string) error {
	tx, err := s.txR.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if tx.Type != walletmodel.TxTypeWithdraw || tx.ReviewStatus != walletmodel.ReviewStatusPendingReview {
		return apperr.ErrWithdrawNotApprovable
	}
	return s.submitToCobo(ctx, tx, adminID)
}

// submitToCobo is shared by Approve and Retry. On Cobo success it transitions
// the ticket to submitted/processing. On Cobo failure it transitions to
// submit_failed WITHOUT touching wallet balances — money stays frozen so the
// admin can retry.
func (s *withdrawReviewService) submitToCobo(ctx context.Context, tx *walletmodel.Transaction, adminID string) error {
	now := time.Now()
	resp, coboErr := s.provider.Withdraw(ctx, cobo.WithdrawReq{
		Currency:  tx.Currency,
		Network:   tx.Network,
		Address:   tx.Address,
		Amount:    tx.Amount,
		RequestID: tx.ID,
	})

	if coboErr != nil {
		s.logger.Error("cobo withdraw failed",
			slog.String("error", coboErr.Error()),
			slog.String("tx_id", tx.ID),
			slog.String("admin_id", adminID),
		)
		fields := map[string]any{
			"review_status":     walletmodel.ReviewStatusSubmitFailed,
			"submit_attempts":   tx.SubmitAttempts + 1,
			"last_submit_error": coboErr.Error(),
			"last_submit_at":    now,
			"reviewed_by":       adminID,
			"reviewed_at":       now,
		}
		if upErr := s.txR.UpdateReview(ctx, tx.ID, fields); upErr != nil {
			return apperr.Wrap(apperr.ErrInternal, fmt.Errorf("persist submit_failed: %w (cobo err: %v)", upErr, coboErr))
		}
		return apperr.Wrap(apperr.ErrCoboSubmitFailed, coboErr)
	}

	fields := map[string]any{
		"review_status":   walletmodel.ReviewStatusSubmitted,
		"status":          walletmodel.TxStatusProcessing,
		"external_id":     resp.ExternalID,
		"submit_attempts": tx.SubmitAttempts + 1,
		"last_submit_at":  now,
		"reviewed_by":     adminID,
		"reviewed_at":     now,
	}
	if err := s.txR.UpdateReview(ctx, tx.ID, fields); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.bus.Publish(ctx, events.Event{
		Type: events.WithdrawRequested,
		Payload: map[string]string{
			"user_id":        tx.UserID,
			"transaction_id": tx.ID,
			"external_id":    resp.ExternalID,
		},
	})

	s.logger.Info("withdrawal submitted to cobo",
		slog.String("tx_id", tx.ID),
		slog.String("external_id", resp.ExternalID),
		slog.String("admin_id", adminID),
	)
	return nil
}

func (s *withdrawReviewService) Reject(ctx context.Context, txID, adminID, reason string) error {
	tx, err := s.txR.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if tx.Type != walletmodel.TxTypeWithdraw {
		return apperr.ErrWithdrawNotRejectable
	}
	switch tx.ReviewStatus {
	case walletmodel.ReviewStatusPendingReview, walletmodel.ReviewStatusSubmitFailed:
		// allowed
	default:
		return apperr.ErrWithdrawNotRejectable
	}

	wallet, err := s.walletR.FindByUserIDAndCurrency(ctx, tx.UserID, tx.Currency)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	origAvailable, origFrozen := wallet.Available, wallet.Frozen
	newAvailable := wallet.Available.Add(tx.Amount)
	newFrozen := wallet.Frozen.Sub(tx.Amount)
	if newFrozen.LessThan(decimal.Zero) {
		newFrozen = decimal.Zero
	}
	if err := s.walletR.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, newFrozen); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	now := time.Now()
	if err := s.txR.UpdateReview(ctx, tx.ID, map[string]any{
		"review_status": walletmodel.ReviewStatusRejected,
		"status":        walletmodel.TxStatusCancelled,
		"reject_reason": reason,
		"reviewed_by":   adminID,
		"reviewed_at":   now,
	}); err != nil {
		// Best-effort revert of refund if persistence fails.
		_ = s.walletR.UpdateBalance(ctx, wallet.ID, origAvailable, wallet.InOperation, origFrozen)
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.bus.Publish(ctx, events.Event{
		Type: events.WithdrawRejected,
		Payload: map[string]string{
			"user_id":        tx.UserID,
			"transaction_id": tx.ID,
			"amount":         tx.Amount.String(),
			"currency":       tx.Currency,
			"reason":         reason,
		},
	})
	return nil
}

func (s *withdrawReviewService) Retry(ctx context.Context, txID, adminID string) error {
	tx, err := s.txR.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if tx.Type != walletmodel.TxTypeWithdraw || tx.ReviewStatus != walletmodel.ReviewStatusSubmitFailed {
		return apperr.ErrWithdrawNotRetriable
	}
	return s.submitToCobo(ctx, tx, adminID)
}

type PGReviewQuerier struct {
	db *gorm.DB
}

func NewPGReviewQuerier(db *gorm.DB) *PGReviewQuerier { return &PGReviewQuerier{db: db} }

func (q *PGReviewQuerier) Query(ctx context.Context, p dto.WithdrawReviewListQuery) ([]reviewListRow, int64, error) {
	db := q.db.WithContext(ctx).Table("transactions").
		Select(`transactions.id, transactions.user_id, users.email AS user_email,
			transactions.currency, transactions.network, transactions.amount, transactions.address,
			transactions.status, transactions.review_status, transactions.reject_reason,
			transactions.submit_attempts, transactions.last_submit_error, transactions.last_submit_at,
			transactions.reviewed_by, transactions.reviewed_at, transactions.created_at,
			transactions.external_id, transactions.tx_hash`).
		Joins("LEFT JOIN users ON users.id = transactions.user_id").
		Where("transactions.type = ?", "withdraw")

	if p.ReviewStatus != "" {
		db = db.Where("transactions.review_status = ?", p.ReviewStatus)
	}
	if p.UserID != "" {
		db = db.Where("transactions.user_id = ?", p.UserID)
	}
	if p.DateFrom != "" {
		db = db.Where("transactions.created_at >= ?", p.DateFrom)
	}
	if p.DateTo != "" {
		db = db.Where("transactions.created_at < ?", p.DateTo+" 23:59:59")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	var rows []reviewListRow
	offset := (p.Page - 1) * p.Limit
	if err := db.Order("transactions.created_at DESC").Offset(offset).Limit(p.Limit).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("find: %w", err)
	}
	return rows, total, nil
}

// Compile-time assertions: production repos satisfy our narrow interfaces.
var _ walletReader = (walletrepo.WalletRepository)(nil)
var _ txWriter = (walletrepo.TransactionRepository)(nil)
var _ reviewQuerier = (*PGReviewQuerier)(nil)
