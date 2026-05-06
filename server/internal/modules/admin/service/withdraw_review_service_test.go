package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"gorm.io/gorm"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type stubReviewQuery struct {
	rows  []reviewListRow
	total int64
	err   error
}

func (s *stubReviewQuery) Query(ctx context.Context, q dto.WithdrawReviewListQuery) ([]reviewListRow, int64, error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	return s.rows, s.total, nil
}

type fakeProvider struct {
	withdrawErr error
	resp        *cobo.WithdrawResp
	calls       int
}

func (f *fakeProvider) Withdraw(ctx context.Context, req cobo.WithdrawReq) (*cobo.WithdrawResp, error) {
	f.calls++
	if f.withdrawErr != nil {
		return nil, f.withdrawErr
	}
	return f.resp, nil
}
func (f *fakeProvider) GenerateAddress(context.Context, cobo.GenerateAddressReq) (*cobo.GenerateAddressResp, error) {
	panic("unused")
}
func (f *fakeProvider) GetBalance(context.Context, string) (*cobo.BalanceResp, error) {
	panic("unused")
}
func (f *fakeProvider) GetTransaction(context.Context, string) (*cobo.TransactionResp, error) {
	panic("unused")
}
func (f *fakeProvider) VerifyWebhook(string, []byte) (bool, error) {
	panic("unused")
}

type stubTxRepo struct {
	byID      map[string]*walletmodel.Transaction
	updates   map[string]map[string]any
	updateErr error
}

func (s *stubTxRepo) FindByID(ctx context.Context, id string) (*walletmodel.Transaction, error) {
	if t, ok := s.byID[id]; ok {
		return t, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubTxRepo) UpdateReview(ctx context.Context, id string, fields map[string]any) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	if s.updates == nil {
		s.updates = map[string]map[string]any{}
	}
	s.updates[id] = fields
	return nil
}
func (s *stubTxRepo) UpdateExternalID(ctx context.Context, id, ext string) error { return nil }
func (s *stubTxRepo) UpdateStatus(ctx context.Context, id, status, txHash string) error {
	return nil
}

type stubWalletRepo struct {
	wallet        *walletmodel.Wallet
	lastAvailable decimal.Decimal
	lastInOp      decimal.Decimal
	lastFrozen    decimal.Decimal
	updateErr     error
}

func (s *stubWalletRepo) FindByUserIDAndCurrency(ctx context.Context, userID, currency string) (*walletmodel.Wallet, error) {
	return s.wallet, nil
}
func (s *stubWalletRepo) UpdateBalance(ctx context.Context, id string, av, op, fz decimal.Decimal) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.lastAvailable, s.lastInOp, s.lastFrozen = av, op, fz
	if s.wallet != nil {
		s.wallet.Available, s.wallet.InOperation, s.wallet.Frozen = av, op, fz
	}
	return nil
}

type spyBus struct {
	events []events.Event
}

func (s *spyBus) Publish(ctx context.Context, e events.Event) { s.events = append(s.events, e) }
func (s *spyBus) Subscribe(string, events.Handler)            {}
func (s *spyBus) Shutdown()                                   {}

// silence unused
var _ = errors.New

func TestList_ReturnsRows(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	q := &stubReviewQuery{
		rows: []reviewListRow{
			{
				ID: "tx1", UserID: "u1", UserEmail: "u@x.com",
				Currency: "USDT", Network: "TRC20",
				Amount: decimal.NewFromInt(50), Address: "Tabc",
				Status:       "pending",
				ReviewStatus: "pending_review",
				CreatedAt:    now,
			},
		},
		total: 1,
	}
	svc := NewWithdrawReviewService(q, nil, nil, nil, nil, newTestLogger())

	items, total, err := svc.List(ctx, dto.WithdrawReviewListQuery{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(items) != 1 {
		t.Fatalf("items length = %d, want 1", len(items))
	}
	if items[0].ID != "tx1" {
		t.Errorf("items[0].ID = %q, want tx1", items[0].ID)
	}
	if items[0].ReviewStatus != "pending_review" {
		t.Errorf("items[0].ReviewStatus = %q, want pending_review", items[0].ReviewStatus)
	}
}

func TestApprove_CoboSuccess_KeepsFrozenAndPublishesEvent(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Network: "TRC20", Address: "Tabc",
		Amount:       decimal.NewFromInt(50),
		Status:       "pending",
		ReviewStatus: "pending_review",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(50),
		Frozen:    decimal.NewFromInt(50),
	}}
	provider := &fakeProvider{resp: &cobo.WithdrawResp{ExternalID: "cobo-ext-1", Status: "processing"}}
	bus := &spyBus{}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, provider, bus, newTestLogger())

	if err := svc.Approve(ctx, "tx1", "admin-1"); err != nil {
		t.Fatalf("Approve error = %v", err)
	}
	if provider.calls != 1 {
		t.Errorf("provider.calls = %d, want 1", provider.calls)
	}
	if !wr.wallet.Frozen.Equal(decimal.NewFromInt(50)) {
		t.Errorf("frozen = %s, want 50 (frozen released only by webhook)", wr.wallet.Frozen.String())
	}
	upd := txR.updates["tx1"]
	if upd["review_status"] != "submitted" {
		t.Errorf("review_status = %v, want submitted", upd["review_status"])
	}
	if upd["status"] != "processing" {
		t.Errorf("status = %v, want processing", upd["status"])
	}
	if upd["external_id"] != "cobo-ext-1" {
		t.Errorf("external_id = %v, want cobo-ext-1", upd["external_id"])
	}
	if len(bus.events) != 1 {
		t.Fatalf("bus.events length = %d, want 1", len(bus.events))
	}
	if bus.events[0].Type != events.WithdrawRequested {
		t.Errorf("event type = %s, want %s", bus.events[0].Type, events.WithdrawRequested)
	}
}

func TestApprove_CoboFailure_KeepsFrozenAndMarksSubmitFailed(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "pending_review",
		SubmitAttempts: 0,
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(50), Frozen: decimal.NewFromInt(50),
	}}
	provider := &fakeProvider{withdrawErr: errors.New("withdraw: 400 Bad Request {\"error_code\": 12007}")}
	bus := &spyBus{}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, provider, bus, newTestLogger())

	err := svc.Approve(ctx, "tx1", "admin-1")
	if err == nil {
		t.Fatal("Approve should return error to admin caller")
	}

	if !wr.wallet.Available.Equal(decimal.NewFromInt(50)) {
		t.Errorf("available = %s, want 50 (no refund on submit failure)", wr.wallet.Available.String())
	}
	if !wr.wallet.Frozen.Equal(decimal.NewFromInt(50)) {
		t.Errorf("frozen = %s, want 50 (still frozen on submit failure)", wr.wallet.Frozen.String())
	}

	upd := txR.updates["tx1"]
	if upd["review_status"] != "submit_failed" {
		t.Errorf("review_status = %v, want submit_failed", upd["review_status"])
	}
	if upd["status"] != nil {
		t.Errorf("status should not be updated on submit failure, got %v", upd["status"])
	}
	if upd["submit_attempts"] != 1 {
		t.Errorf("submit_attempts = %v, want 1", upd["submit_attempts"])
	}
	errStr, ok := upd["last_submit_error"].(string)
	if !ok || !strings.Contains(errStr, "12007") {
		t.Errorf("last_submit_error should contain 12007, got %v", upd["last_submit_error"])
	}
	if upd["last_submit_at"] == nil {
		t.Error("last_submit_at should be set")
	}

	if len(bus.events) != 0 {
		t.Errorf("bus.events should be empty on submit failure, got %d", len(bus.events))
	}
}

func TestApprove_RejectsNonPendingReview(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		ReviewStatus: "submitted",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, &stubWalletRepo{}, txR, &fakeProvider{}, &spyBus{}, newTestLogger())

	err := svc.Approve(ctx, "tx1", "admin-1")
	if err == nil {
		t.Fatal("expected error for non-pending_review tx")
	}
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != "WITHDRAW_NOT_APPROVABLE" {
		t.Errorf("error code = %v, want WITHDRAW_NOT_APPROVABLE", ae)
	}
}

func TestReject_RefundsAndPersistsReason(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "pending_review",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(50), Frozen: decimal.NewFromInt(50),
	}}
	bus := &spyBus{}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, &fakeProvider{}, bus, newTestLogger())

	if err := svc.Reject(ctx, "tx1", "admin-1", "address mismatch"); err != nil {
		t.Fatalf("Reject error = %v", err)
	}
	if !wr.wallet.Available.Equal(decimal.NewFromInt(100)) {
		t.Errorf("available = %s, want 100 (refunded)", wr.wallet.Available.String())
	}
	if !wr.wallet.Frozen.Equal(decimal.Zero) {
		t.Errorf("frozen = %s, want 0", wr.wallet.Frozen.String())
	}

	upd := txR.updates["tx1"]
	if upd["review_status"] != "rejected" {
		t.Errorf("review_status = %v, want rejected", upd["review_status"])
	}
	if upd["status"] != "cancelled" {
		t.Errorf("status = %v, want cancelled", upd["status"])
	}
	if upd["reject_reason"] != "address mismatch" {
		t.Errorf("reject_reason = %v, want 'address mismatch'", upd["reject_reason"])
	}
	if upd["reviewed_by"] != "admin-1" {
		t.Errorf("reviewed_by = %v, want admin-1", upd["reviewed_by"])
	}

	if len(bus.events) != 1 {
		t.Fatalf("bus.events length = %d, want 1", len(bus.events))
	}
	if bus.events[0].Type != events.WithdrawRejected {
		t.Errorf("event type = %s, want %s", bus.events[0].Type, events.WithdrawRejected)
	}
	payload, ok := bus.events[0].Payload.(map[string]string)
	if !ok {
		t.Fatalf("event payload type = %T, want map[string]string", bus.events[0].Payload)
	}
	if payload["reason"] != "address mismatch" {
		t.Errorf("event reason = %s, want 'address mismatch'", payload["reason"])
	}
}

func TestReject_AllowedFromSubmitFailed(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "submit_failed",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(0), Frozen: decimal.NewFromInt(50),
	}}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, &fakeProvider{}, &spyBus{}, newTestLogger())

	if err := svc.Reject(ctx, "tx1", "admin-1", "manual cleanup"); err != nil {
		t.Fatalf("Reject error = %v", err)
	}
	if !wr.wallet.Available.Equal(decimal.NewFromInt(50)) {
		t.Errorf("available = %s, want 50", wr.wallet.Available.String())
	}
	if !wr.wallet.Frozen.Equal(decimal.Zero) {
		t.Errorf("frozen = %s, want 0", wr.wallet.Frozen.String())
	}
}

func TestReject_RejectsTerminalState(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		ReviewStatus: "submitted",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, &stubWalletRepo{}, txR, &fakeProvider{}, &spyBus{}, newTestLogger())
	err := svc.Reject(ctx, "tx1", "admin-1", "x")
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != "WITHDRAW_NOT_REJECTABLE" {
		t.Errorf("error = %v, want WITHDRAW_NOT_REJECTABLE", err)
	}
}
