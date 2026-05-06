package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestWithdraw_QueuesPendingReviewWithoutCallingCobo(t *testing.T) {
	ctx := context.Background()
	walletRepo := &stubWalletRepo{
		wallet: &model.Wallet{
			ID:        "w1",
			UserID:    "u1",
			Currency:  "USDT",
			Available: decimal.NewFromInt(100),
			Frozen:    decimal.Zero,
		},
	}
	addrRepo := &stubAddressRepo{
		whitelist: &model.WithdrawAddress{
			UserID:        "u1",
			Currency:      "USDT",
			Network:       "TRC20",
			Address:       "Tabc",
			CooldownUntil: time.Now().Add(-time.Hour),
			IsActive:      true,
		},
	}
	txRepo := &stubTxRepo{}
	provider := &mockCoboProvider{}
	bus := &spyEventBus{}
	twoFA := &stubTwoFA{ok: true}

	svc := NewWalletService(walletRepo, addrRepo, txRepo, provider, bus, twoFA, newTestLogger(), 24*time.Hour)

	resp, err := svc.Withdraw(ctx, "u1", dto.WithdrawRequest{
		Currency:  "USDT",
		Network:   "TRC20",
		Address:   "Tabc",
		Amount:    "50",
		TwoFACode: "123456",
	})

	if err != nil {
		t.Fatalf("Withdraw() error = %v, want nil", err)
	}
	if resp == nil {
		t.Fatal("Withdraw() resp = nil, want non-nil")
	}
	if resp.TransactionID == "" {
		t.Error("Withdraw() resp.TransactionID is empty")
	}
	if resp.Status != model.ReviewStatusPendingReview {
		t.Errorf("Withdraw() resp.Status = %q, want %q", resp.Status, model.ReviewStatusPendingReview)
	}

	if !walletRepo.lastAvailable.Equal(decimal.NewFromInt(50)) {
		t.Errorf("wallet available after freeze = %s, want 50", walletRepo.lastAvailable.String())
	}
	if !walletRepo.lastFrozen.Equal(decimal.NewFromInt(50)) {
		t.Errorf("wallet frozen after freeze = %s, want 50", walletRepo.lastFrozen.String())
	}
	if walletRepo.updateCalls != 1 {
		t.Errorf("wallet UpdateBalance call count = %d, want 1", walletRepo.updateCalls)
	}

	if len(txRepo.created) != 1 {
		t.Fatalf("transactions created = %d, want 1", len(txRepo.created))
	}
	tx := txRepo.created[0]
	if tx.Status != model.TxStatusPending {
		t.Errorf("created tx.Status = %q, want %q", tx.Status, model.TxStatusPending)
	}
	if tx.ReviewStatus != model.ReviewStatusPendingReview {
		t.Errorf("created tx.ReviewStatus = %q, want %q", tx.ReviewStatus, model.ReviewStatusPendingReview)
	}
	if tx.Type != model.TxTypeWithdraw {
		t.Errorf("created tx.Type = %q, want %q", tx.Type, model.TxTypeWithdraw)
	}
	if !tx.Amount.Equal(decimal.NewFromInt(50)) {
		t.Errorf("created tx.Amount = %s, want 50", tx.Amount.String())
	}

	if provider.withdrawCalls != 0 {
		t.Errorf("provider.Withdraw calls = %d, want 0 (Cobo must not be invoked)", provider.withdrawCalls)
	}
	if len(bus.published) != 0 {
		t.Errorf("event bus publishes = %d, want 0 (no event published from Withdraw)", len(bus.published))
	}
}

func TestCancelWithdraw_RestoresFrozen(t *testing.T) {
	ctx := context.Background()
	walletRepo := &stubWalletRepo{
		wallet: &model.Wallet{
			ID:        "w1",
			UserID:    "u1",
			Currency:  "USDT",
			Available: decimal.NewFromInt(50),
			Frozen:    decimal.NewFromInt(50),
		},
	}
	txRepo := &stubTxRepo{
		byID: map[string]*model.Transaction{
			"tx1": {
				ID:           "tx1",
				UserID:       "u1",
				Type:         model.TxTypeWithdraw,
				Currency:     "USDT",
				Amount:       decimal.NewFromInt(50),
				Status:       model.TxStatusPending,
				ReviewStatus: model.ReviewStatusPendingReview,
			},
		},
	}
	svc := NewWalletService(walletRepo, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, newTestLogger(), 24*time.Hour)

	if err := svc.CancelWithdraw(ctx, "u1", "tx1"); err != nil {
		t.Fatalf("CancelWithdraw() error = %v, want nil", err)
	}

	if !walletRepo.lastAvailable.Equal(decimal.NewFromInt(100)) {
		t.Errorf("wallet available after refund = %s, want 100", walletRepo.lastAvailable.String())
	}
	if !walletRepo.lastFrozen.Equal(decimal.Zero) {
		t.Errorf("wallet frozen after refund = %s, want 0", walletRepo.lastFrozen.String())
	}

	updated, ok := txRepo.lastReviewUpdate["tx1"]
	if !ok {
		t.Fatal("UpdateReview was not called for tx1")
	}
	if got := updated["review_status"]; got != model.ReviewStatusCancelled {
		t.Errorf("review_status = %v, want %v", got, model.ReviewStatusCancelled)
	}
	if got := updated["status"]; got != model.TxStatusCancelled {
		t.Errorf("status = %v, want %v", got, model.TxStatusCancelled)
	}
}

func TestCancelWithdraw_RejectsWrongOwner(t *testing.T) {
	ctx := context.Background()
	txRepo := &stubTxRepo{byID: map[string]*model.Transaction{
		"tx1": {
			ID:           "tx1",
			UserID:       "other",
			Type:         model.TxTypeWithdraw,
			ReviewStatus: model.ReviewStatusPendingReview,
		},
	}}
	svc := NewWalletService(&stubWalletRepo{}, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, newTestLogger(), 24*time.Hour)

	err := svc.CancelWithdraw(ctx, "u1", "tx1")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("CancelWithdraw() error = %v, want ErrNotFound", err)
	}
}

func TestGetTransactions_IncludesReviewFields(t *testing.T) {
	ctx := context.Background()
	rejectedAt := time.Now()
	txRepo := &stubTxRepo{
		listByUser: []model.Transaction{
			{
				ID: "tx1", UserID: "u1", Type: model.TxTypeWithdraw,
				Currency: "USDT", Network: "TRC20",
				Amount: decimal.NewFromInt(50), Address: "Tabc",
				Status:       model.TxStatusCancelled,
				ReviewStatus: model.ReviewStatusRejected,
				RejectReason: "address mismatch",
				ReviewedAt:   &rejectedAt,
			},
		},
		listTotal: 1,
	}
	svc := NewWalletService(&stubWalletRepo{}, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, newTestLogger(), 24*time.Hour)

	items, total, err := svc.GetTransactions(ctx, "u1", "withdraw", 1, 10)
	if err != nil {
		t.Fatalf("GetTransactions error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(items) != 1 {
		t.Fatalf("items length = %d, want 1", len(items))
	}
	if items[0].ReviewStatus != "rejected" {
		t.Errorf("ReviewStatus = %q, want rejected", items[0].ReviewStatus)
	}
	if items[0].RejectReason != "address mismatch" {
		t.Errorf("RejectReason = %q, want 'address mismatch'", items[0].RejectReason)
	}
}

func TestCancelWithdraw_RejectsNonPendingReview(t *testing.T) {
	ctx := context.Background()
	txRepo := &stubTxRepo{byID: map[string]*model.Transaction{
		"tx1": {
			ID:           "tx1",
			UserID:       "u1",
			Type:         model.TxTypeWithdraw,
			ReviewStatus: model.ReviewStatusSubmitted,
		},
	}}
	svc := NewWalletService(&stubWalletRepo{}, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, newTestLogger(), 24*time.Hour)

	err := svc.CancelWithdraw(ctx, "u1", "tx1")
	var ae *apperr.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("CancelWithdraw() error = %v, want *apperr.AppError", err)
	}
	if ae.Code != "WITHDRAW_NOT_CANCELLABLE" {
		t.Errorf("AppError.Code = %q, want %q", ae.Code, "WITHDRAW_NOT_CANCELLABLE")
	}
}
