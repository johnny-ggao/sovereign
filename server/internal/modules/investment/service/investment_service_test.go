package service

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	investRepo "github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	walletModel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	walletRepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
)

func TestRedeemMarksInvestmentStoppingWithoutTouchingWallet(t *testing.T) {
	t.Parallel()

	inv := &model.Investment{
		ID:        "investment-1",
		UserID:    "user-1",
		Amount:    decimal.NewFromInt(100),
		Currency:  "USDT",
		Status:    model.InvestStatusActive,
		NetReturn: decimal.NewFromInt(12),
		StartDate: time.Now().Add(-24 * time.Hour),
	}

	invRepo := &stubInvestmentRepository{byID: map[string]*model.Investment{inv.ID: inv}}
	walletRepo := &stubWalletRepository{}
	bus := &recordingBus{}

	svc := NewInvestmentService(
		invRepo,
		walletRepo,
		bus,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	resp, err := svc.Redeem(context.Background(), inv.UserID, dto.RedeemRequest{InvestmentID: inv.ID})
	if err != nil {
		t.Fatalf("Redeem() error = %v", err)
	}

	if resp.Status != model.InvestStatusStopping {
		t.Fatalf("Redeem() status = %q, want %q", resp.Status, model.InvestStatusStopping)
	}
	if inv.Status != model.InvestStatusStopping {
		t.Fatalf("investment status = %q, want %q", inv.Status, model.InvestStatusStopping)
	}
	if inv.EndDate == nil {
		t.Fatal("investment end date was not set")
	}
	if invRepo.updated != inv {
		t.Fatal("expected investment update to persist the modified investment")
	}
	if walletRepo.findCalls != 0 {
		t.Fatalf("wallet lookup calls = %d, want 0", walletRepo.findCalls)
	}
	if walletRepo.updateCalls != 0 {
		t.Fatalf("wallet update calls = %d, want 0", walletRepo.updateCalls)
	}
	if len(bus.published) != 1 {
		t.Fatalf("published events = %d, want 1", len(bus.published))
	}
	if bus.published[0].Type != events.InvestmentRedeemed {
		t.Fatalf("event type = %q, want %q", bus.published[0].Type, events.InvestmentRedeemed)
	}
}

type stubInvestmentRepository struct {
	byID    map[string]*model.Investment
	updated *model.Investment
}

func (s *stubInvestmentRepository) Create(context.Context, *model.Investment) error {
	panic("unexpected Create call")
}

func (s *stubInvestmentRepository) FindByID(_ context.Context, id string) (*model.Investment, error) {
	return s.byID[id], nil
}

func (s *stubInvestmentRepository) FindByUserID(context.Context, string) ([]model.Investment, error) {
	panic("unexpected FindByUserID call")
}

func (s *stubInvestmentRepository) FindActiveByUserID(context.Context, string) ([]model.Investment, error) {
	panic("unexpected FindActiveByUserID call")
}

func (s *stubInvestmentRepository) FindAllActive(context.Context) ([]model.Investment, error) {
	panic("unexpected FindAllActive call")
}

func (s *stubInvestmentRepository) Update(_ context.Context, inv *model.Investment) error {
	s.updated = inv
	return nil
}

var _ investRepo.InvestmentRepository = (*stubInvestmentRepository)(nil)

type stubWalletRepository struct {
	findCalls   int
	updateCalls int
}

func (s *stubWalletRepository) FindByUserID(context.Context, string) ([]walletModel.Wallet, error) {
	panic("unexpected FindByUserID call")
}

func (s *stubWalletRepository) FindByUserIDAndCurrency(context.Context, string, string) (*walletModel.Wallet, error) {
	s.findCalls++
	return nil, nil
}

func (s *stubWalletRepository) FindOrCreate(context.Context, string, string) (*walletModel.Wallet, error) {
	panic("unexpected FindOrCreate call")
}

func (s *stubWalletRepository) UpdateBalance(context.Context, string, decimal.Decimal, decimal.Decimal, decimal.Decimal) error {
	s.updateCalls++
	return nil
}

func (s *stubWalletRepository) AddEarnings(context.Context, string, decimal.Decimal) error {
	panic("unexpected AddEarnings call")
}

func (s *stubWalletRepository) ClaimEarnings(context.Context, string) error {
	panic("unexpected ClaimEarnings call")
}

var _ walletRepo.WalletRepository = (*stubWalletRepository)(nil)

type recordingBus struct {
	published []events.Event
}

func (b *recordingBus) Publish(_ context.Context, event events.Event) {
	b.published = append(b.published, event)
}

func (b *recordingBus) Subscribe(string, events.Handler) {}

func (b *recordingBus) Shutdown() {}
