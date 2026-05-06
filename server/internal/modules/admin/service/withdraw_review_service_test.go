package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
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
