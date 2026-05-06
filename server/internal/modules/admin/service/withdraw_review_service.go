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

// Stubs filled in Tasks 8-10.
func (s *withdrawReviewService) Approve(ctx context.Context, txID, adminID string) error {
	return errors.New("not implemented")
}
func (s *withdrawReviewService) Reject(ctx context.Context, txID, adminID, reason string) error {
	return errors.New("not implemented")
}
func (s *withdrawReviewService) Retry(ctx context.Context, txID, adminID string) error {
	return errors.New("not implemented")
}

// Compile-time assertions: production repos satisfy our narrow interfaces.
var _ walletReader = (walletrepo.WalletRepository)(nil)
var _ txWriter = (walletrepo.TransactionRepository)(nil)

// Keep imports live until later tasks consume them.
var _ = fmt.Sprintf
var _ = apperr.ErrInternal
var _ = (*gorm.DB)(nil)
