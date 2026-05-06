package service

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
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
