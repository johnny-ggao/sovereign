package service

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	usermodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
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
	uRepo := &stubUserRepo{}
	bus := &recordingBus{}

	svc := NewInvestmentService(
		invRepo,
		walletRepo,
		uRepo,
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
	byID             map[string]*model.Investment
	updated          *model.Investment
	created          []*model.Investment
	byUserAndProduct map[string][]model.Investment
}

func (s *stubInvestmentRepository) Create(_ context.Context, inv *model.Investment) error {
	s.created = append(s.created, inv)
	return nil
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

func (s *stubInvestmentRepository) FindAllActiveBeforeDate(context.Context, time.Time) ([]model.Investment, error) {
	panic("unexpected FindAllActiveBeforeDate call")
}

func (s *stubInvestmentRepository) FindAllActiveBeforeDateByProduct(context.Context, time.Time, string) ([]model.Investment, error) {
	return nil, nil
}

func (s *stubInvestmentRepository) FindByUserIDAndProduct(_ context.Context, userID, productType string) ([]model.Investment, error) {
	key := userID + "|" + productType
	return s.byUserAndProduct[key], nil
}

func (s *stubInvestmentRepository) Update(_ context.Context, inv *model.Investment) error {
	s.updated = inv
	return nil
}

var _ investRepo.InvestmentRepository = (*stubInvestmentRepository)(nil)

type stubWalletRepository struct {
	findCalls   int
	updateCalls int
	wallet      *walletModel.Wallet
}

func (s *stubWalletRepository) FindByUserID(context.Context, string) ([]walletModel.Wallet, error) {
	panic("unexpected FindByUserID call")
}

func (s *stubWalletRepository) FindByUserIDAndCurrency(context.Context, string, string) (*walletModel.Wallet, error) {
	s.findCalls++
	return s.wallet, nil
}

func (s *stubWalletRepository) FindOrCreate(context.Context, string, string) (*walletModel.Wallet, error) {
	panic("unexpected FindOrCreate call")
}

func (s *stubWalletRepository) UpdateBalance(context.Context, string, decimal.Decimal, decimal.Decimal, decimal.Decimal) error {
	s.updateCalls++
	return nil
}

func (s *stubWalletRepository) AddEarnings(context.Context, string, decimal.Decimal, string) error {
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

type stubUserRepo struct {
	users map[string]*usermodel.User
}

func (s *stubUserRepo) FindByID(_ context.Context, id string) (*usermodel.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return &usermodel.User{ID: id}, nil
}

func (s *stubUserRepo) FindByEmail(context.Context, string) (*usermodel.User, error) {
	panic("unexpected FindByEmail call")
}

func (s *stubUserRepo) FindByGoogleID(context.Context, string) (*usermodel.User, error) {
	panic("unexpected FindByGoogleID call")
}

func (s *stubUserRepo) Create(context.Context, *usermodel.User) error {
	panic("unexpected Create call")
}

func (s *stubUserRepo) Update(context.Context, *usermodel.User) error {
	panic("unexpected Update call")
}

func (s *stubUserRepo) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}

var _ authRepo.UserRepository = (*stubUserRepo)(nil)

func TestCreate_AssignsProductTypeFromUser(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvestmentRepository{}
	wRepo := &stubWalletRepository{wallet: &walletModel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(500),
	}}
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{
		"u1": {ID: "u1", InvestmentType: "trading"},
	}}
	bus := &recordingBus{}
	svc := NewInvestmentService(invRepo, wRepo, uRepo, bus, slog.New(slog.NewTextHandler(io.Discard, nil)))

	resp, err := svc.Create(ctx, "u1", dto.CreateInvestmentRequest{Amount: "200", Currency: "USDT"})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp == nil || resp.ProductType != "trading" {
		t.Errorf("response.ProductType mismatch, resp=%+v", resp)
	}
	if len(invRepo.created) != 1 || invRepo.created[0].ProductType != "trading" {
		t.Errorf("persisted product_type wrong: %+v", invRepo.created)
	}
}

func TestCreate_DefaultsToArbitrageWhenUserHasNoType(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvestmentRepository{}
	wRepo := &stubWalletRepository{wallet: &walletModel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(500),
	}}
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{
		"u1": {ID: "u1", InvestmentType: ""},
	}}
	svc := NewInvestmentService(invRepo, wRepo, uRepo, &recordingBus{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	resp, _ := svc.Create(ctx, "u1", dto.CreateInvestmentRequest{Amount: "200", Currency: "USDT"})
	if resp == nil || resp.ProductType != "arbitrage" {
		t.Errorf("ProductType = %v, want arbitrage", resp)
	}
}

func TestGetAll_FiltersByProductType(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvestmentRepository{
		byUserAndProduct: map[string][]model.Investment{
			"u1|trading":   {{ID: "i1", UserID: "u1", ProductType: "trading", Status: "active", Amount: decimal.NewFromInt(100)}},
			"u1|arbitrage": {{ID: "i2", UserID: "u1", ProductType: "arbitrage", Status: "active", Amount: decimal.NewFromInt(200)}},
		},
	}
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: "trading"}}}
	wRepo := &stubWalletRepository{}
	svc := NewInvestmentService(invRepo, wRepo, uRepo, &recordingBus{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	resp, err := svc.GetAll(ctx, "u1", "trading")
	if err != nil {
		t.Fatalf("GetAll trading error = %v", err)
	}
	if len(resp.Investments) != 1 || resp.Investments[0].ProductType != "trading" {
		t.Errorf("trading filter: got %+v", resp.Investments)
	}

	respArb, err := svc.GetAll(ctx, "u1", "arbitrage")
	if err != nil {
		t.Fatalf("GetAll arbitrage error = %v", err)
	}
	if len(respArb.Investments) != 1 || respArb.Investments[0].ProductType != "arbitrage" {
		t.Errorf("arbitrage filter: got %+v", respArb.Investments)
	}
}

func TestGetAll_DefaultsToUserInvestmentType(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvestmentRepository{
		byUserAndProduct: map[string][]model.Investment{
			"u1|trading":   {{ID: "i1", UserID: "u1", ProductType: "trading", Status: "active", Amount: decimal.NewFromInt(100)}},
			"u1|arbitrage": {{ID: "i2", UserID: "u1", ProductType: "arbitrage", Status: "active", Amount: decimal.NewFromInt(200)}},
		},
	}
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: "trading"}}}
	wRepo := &stubWalletRepository{}
	svc := NewInvestmentService(invRepo, wRepo, uRepo, &recordingBus{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// productType="" → default to user's InvestmentType
	resp, err := svc.GetAll(ctx, "u1", "")
	if err != nil {
		t.Fatalf("GetAll error = %v", err)
	}
	if len(resp.Investments) != 1 || resp.Investments[0].ProductType != "trading" {
		t.Errorf("default-to-user-type: got %+v", resp.Investments)
	}
}
