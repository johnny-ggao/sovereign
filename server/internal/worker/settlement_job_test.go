package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	invmodel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	settlmodel "github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	tradelogmodel "github.com/sovereign-fund/sovereign/internal/modules/tradelog/model"
	traderepo "github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	tradingmodel "github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/model"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
)

// === stubs (minimal, only methods exercised) ===

type stubInvRepo struct {
	byProduct map[string][]invmodel.Investment
}

func (s *stubInvRepo) FindAllActiveBeforeDateByProduct(_ context.Context, _ time.Time, productType string) ([]invmodel.Investment, error) {
	return s.byProduct[productType], nil
}

func (s *stubInvRepo) Update(_ context.Context, _ *invmodel.Investment) error { return nil }

func (s *stubInvRepo) Create(context.Context, *invmodel.Investment) error      { panic("unused") }
func (s *stubInvRepo) FindByID(context.Context, string) (*invmodel.Investment, error) {
	panic("unused")
}
func (s *stubInvRepo) FindByUserID(context.Context, string) ([]invmodel.Investment, error) {
	panic("unused")
}
func (s *stubInvRepo) FindActiveByUserID(context.Context, string) ([]invmodel.Investment, error) {
	panic("unused")
}
func (s *stubInvRepo) FindAllActive(context.Context) ([]invmodel.Investment, error) { panic("unused") }
func (s *stubInvRepo) FindAllActiveBeforeDate(context.Context, time.Time) ([]invmodel.Investment, error) {
	panic("unused")
}
func (s *stubInvRepo) FindByUserIDAndProduct(context.Context, string, string) ([]invmodel.Investment, error) {
	panic("unused")
}

type stubTradeRepo struct {
	summary   *traderepo.TradeSummaryResult
	err       error
	dayTrades []tradelogmodel.Trade
}

func (s *stubTradeRepo) SummarizeByPeriod(context.Context, time.Time, time.Time) (*traderepo.TradeSummaryResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.summary == nil {
		return &traderepo.TradeSummaryResult{}, nil
	}
	return s.summary, nil
}

func (s *stubTradeRepo) FindByPeriod(context.Context, time.Time, time.Time) ([]tradelogmodel.Trade, error) {
	return s.dayTrades, nil
}

func (s *stubTradeRepo) Create(context.Context, *tradelogmodel.Trade) error { panic("unused") }
func (s *stubTradeRepo) FindByUserID(context.Context, string, traderepo.TradeFilters, int, int) ([]tradelogmodel.Trade, int64, error) {
	panic("unused")
}
func (s *stubTradeRepo) FindAll(context.Context, traderepo.TradeFilters, int, int) ([]tradelogmodel.Trade, int64, error) {
	panic("unused")
}
func (s *stubTradeRepo) FindByInvestmentID(context.Context, string, int, int) ([]tradelogmodel.Trade, int64, error) {
	panic("unused")
}
func (s *stubTradeRepo) CountByInvestmentAndPeriod(context.Context, string, time.Time, time.Time) (int64, error) {
	panic("unused")
}
func (s *stubTradeRepo) SummarizeByUserID(context.Context, string, traderepo.TradeFilters) (*traderepo.TradeSummaryResult, error) {
	panic("unused")
}
func (s *stubTradeRepo) SummarizeAll(context.Context, traderepo.TradeFilters) (*traderepo.TradeSummaryResult, error) {
	panic("unused")
}

type stubTradingTradeRepo struct {
	summary   *traderepo.TradeSummaryResult
	err       error
	dayTrades []tradingmodel.TradingTrade
}

func (s *stubTradingTradeRepo) SummarizeByPeriod(context.Context, time.Time, time.Time) (*traderepo.TradeSummaryResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.summary == nil {
		return &traderepo.TradeSummaryResult{}, nil
	}
	return s.summary, nil
}

func (s *stubTradingTradeRepo) FindByPeriod(context.Context, time.Time, time.Time) ([]tradingmodel.TradingTrade, error) {
	return s.dayTrades, nil
}

func (s *stubTradingTradeRepo) Create(context.Context, *tradingmodel.TradingTrade) error {
	panic("unused")
}
func (s *stubTradingTradeRepo) BatchCreate(context.Context, []tradingmodel.TradingTrade) error {
	panic("unused")
}
func (s *stubTradingTradeRepo) FindByID(context.Context, string) (*tradingmodel.TradingTrade, error) {
	panic("unused")
}
func (s *stubTradingTradeRepo) FindAll(context.Context, traderepo.TradeFilters, int, int) ([]tradingmodel.TradingTrade, int64, error) {
	panic("unused")
}
func (s *stubTradingTradeRepo) SummarizeAll(context.Context, traderepo.TradeFilters) (*traderepo.TradeSummaryResult, error) {
	panic("unused")
}
func (s *stubTradingTradeRepo) Delete(context.Context, string) error { panic("unused") }

type stubUserTradeRepo struct{}

func (s *stubUserTradeRepo) BatchCreate(context.Context, []tradelogmodel.UserTrade) error {
	return nil
}

func (s *stubUserTradeRepo) ExistsBySettlementID(context.Context, string) (bool, error) {
	return false, nil
}

func (s *stubUserTradeRepo) FindByUserID(context.Context, string, traderepo.UserTradeFilters, int, int) ([]tradelogmodel.UserTrade, int64, error) {
	panic("unused")
}

func (s *stubUserTradeRepo) SummarizeByUserID(context.Context, string, traderepo.UserTradeFilters) (*traderepo.TradeSummaryResult, error) {
	panic("unused")
}

type stubSettlRepo struct {
	created []settlmodel.Settlement
}

func (s *stubSettlRepo) Create(_ context.Context, st *settlmodel.Settlement) error {
	if st.ID == "" {
		st.ID = "fake-settlement-id"
	}
	s.created = append(s.created, *st)
	return nil
}

func (s *stubSettlRepo) FindByInvestmentAndPeriod(context.Context, string, string) (*settlmodel.Settlement, error) {
	return nil, nil
}

func (s *stubSettlRepo) FindByUserID(context.Context, string) ([]settlmodel.Settlement, error) {
	panic("unused")
}
func (s *stubSettlRepo) FindByUserIDAndProduct(context.Context, string, string) ([]settlmodel.Settlement, error) {
	panic("unused")
}
func (s *stubSettlRepo) FindByUserIDAndPeriodRange(context.Context, string, string, string) ([]settlmodel.Settlement, error) {
	panic("unused")
}
func (s *stubSettlRepo) FindByID(context.Context, string) (*settlmodel.Settlement, error) {
	panic("unused")
}

type earningsCall struct {
	walletID    string
	amount      decimal.Decimal
	productType string
}

type stubWalletRepoSj struct {
	walletByUser  map[string]*walletmodel.Wallet
	earningsCalls []earningsCall
}

func (s *stubWalletRepoSj) FindByUserIDAndCurrency(_ context.Context, userID, _ string) (*walletmodel.Wallet, error) {
	if w, ok := s.walletByUser[userID]; ok {
		return w, nil
	}
	return nil, errors.New("wallet not found")
}

func (s *stubWalletRepoSj) AddEarnings(_ context.Context, walletID string, amount decimal.Decimal, productType string) error {
	s.earningsCalls = append(s.earningsCalls, earningsCall{walletID, amount, productType})
	return nil
}

func (s *stubWalletRepoSj) UpdateBalance(context.Context, string, decimal.Decimal, decimal.Decimal, decimal.Decimal) error {
	return nil
}

func (s *stubWalletRepoSj) FindByUserID(context.Context, string) ([]walletmodel.Wallet, error) {
	panic("unused")
}
func (s *stubWalletRepoSj) FindOrCreate(context.Context, string, string) (*walletmodel.Wallet, error) {
	panic("unused")
}
func (s *stubWalletRepoSj) ClaimEarnings(context.Context, string) error { panic("unused") }

type spyBus struct{ events []events.Event }

func (b *spyBus) Publish(_ context.Context, e events.Event) { b.events = append(b.events, e) }
func (b *spyBus) Subscribe(string, events.Handler)          {}
func (b *spyBus) Shutdown()                                 {}

func newJobLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// === tests ===

func TestSettleProduct_IsolatesArbitrageFromTrading(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	invRepo := &stubInvRepo{byProduct: map[string][]invmodel.Investment{
		invmodel.ProductTypeArbitrage: {{ID: "a1", UserID: "ua", Amount: decimal.NewFromInt(100), Currency: "USDT", ProductType: invmodel.ProductTypeArbitrage, Status: invmodel.InvestStatusActive}},
		invmodel.ProductTypeTrading:   {{ID: "t1", UserID: "ut", Amount: decimal.NewFromInt(100), Currency: "USDT", ProductType: invmodel.ProductTypeTrading, Status: invmodel.InvestStatusActive}},
	}}
	arbTradeRepo := &stubTradeRepo{summary: &traderepo.TradeSummaryResult{TotalTrades: 1, TotalPnL: 100}}
	tradingTradeRepo := &stubTradingTradeRepo{summary: &traderepo.TradeSummaryResult{TotalTrades: 1, TotalPnL: 200}}
	walletRepo := &stubWalletRepoSj{walletByUser: map[string]*walletmodel.Wallet{
		"ua": {ID: "wa", UserID: "ua", Currency: "USDT"},
		"ut": {ID: "wt", UserID: "ut", Currency: "USDT"},
	}}
	settl := &stubSettlRepo{}

	job := NewSettlementJob(invRepo, arbTradeRepo, tradingTradeRepo, &stubUserTradeRepo{}, settl, walletRepo, &spyBus{}, newJobLogger())
	if err := job.RunForDate(ctx, day); err != nil {
		t.Fatalf("RunForDate error = %v", err)
	}

	// Verify earnings calls split by productType
	arbCalls, tradingCalls := 0, 0
	for _, c := range walletRepo.earningsCalls {
		if c.productType == invmodel.ProductTypeArbitrage && c.walletID == "wa" {
			arbCalls++
		}
		if c.productType == invmodel.ProductTypeTrading && c.walletID == "wt" {
			tradingCalls++
		}
		if c.productType == invmodel.ProductTypeArbitrage && c.walletID == "wt" {
			t.Errorf("trading wallet got arbitrage earnings")
		}
		if c.productType == invmodel.ProductTypeTrading && c.walletID == "wa" {
			t.Errorf("arbitrage wallet got trading earnings")
		}
	}
	if arbCalls != 1 {
		t.Errorf("expected 1 arb earnings call, got %d", arbCalls)
	}
	if tradingCalls != 1 {
		t.Errorf("expected 1 trading earnings call, got %d", tradingCalls)
	}

	// Each settlement record should have the right product_type
	if len(settl.created) != 2 {
		t.Fatalf("expected 2 settlements, got %d", len(settl.created))
	}
	for _, s := range settl.created {
		if s.UserID == "ua" && s.ProductType != invmodel.ProductTypeArbitrage {
			t.Errorf("ua settlement product_type=%s, want arbitrage", s.ProductType)
		}
		if s.UserID == "ut" && s.ProductType != invmodel.ProductTypeTrading {
			t.Errorf("ut settlement product_type=%s, want trading", s.ProductType)
		}
	}
}

func TestSettleProduct_SkipWhenNoActiveInvestments(t *testing.T) {
	ctx := context.Background()
	job := NewSettlementJob(
		&stubInvRepo{byProduct: map[string][]invmodel.Investment{}},
		&stubTradeRepo{}, &stubTradingTradeRepo{}, &stubUserTradeRepo{}, &stubSettlRepo{},
		&stubWalletRepoSj{}, &spyBus{}, newJobLogger(),
	)
	if err := job.RunForDate(ctx, time.Now()); err != nil {
		t.Fatalf("error = %v", err)
	}
}

func TestSettleProduct_OneProductFailureDoesNotBlockOther(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	invRepo := &stubInvRepo{byProduct: map[string][]invmodel.Investment{
		invmodel.ProductTypeArbitrage: {{ID: "a1", UserID: "ua", Amount: decimal.NewFromInt(100), Currency: "USDT", ProductType: invmodel.ProductTypeArbitrage, Status: invmodel.InvestStatusActive}},
		invmodel.ProductTypeTrading:   {{ID: "t1", UserID: "ut", Amount: decimal.NewFromInt(100), Currency: "USDT", ProductType: invmodel.ProductTypeTrading, Status: invmodel.InvestStatusActive}},
	}}
	arbTradeRepo := &stubTradeRepo{err: errors.New("simulated arb failure")}
	tradingTradeRepo := &stubTradingTradeRepo{summary: &traderepo.TradeSummaryResult{TotalTrades: 1, TotalPnL: 100}}
	walletRepo := &stubWalletRepoSj{walletByUser: map[string]*walletmodel.Wallet{
		"ua": {ID: "wa", UserID: "ua", Currency: "USDT"},
		"ut": {ID: "wt", UserID: "ut", Currency: "USDT"},
	}}
	job := NewSettlementJob(invRepo, arbTradeRepo, tradingTradeRepo, &stubUserTradeRepo{}, &stubSettlRepo{}, walletRepo, &spyBus{}, newJobLogger())

	// First error should propagate (arb failed) but trading should still settle
	_ = job.RunForDate(ctx, day)

	tradingSettled := false
	for _, c := range walletRepo.earningsCalls {
		if c.productType == invmodel.ProductTypeTrading && c.walletID == "wt" {
			tradingSettled = true
			break
		}
	}
	if !tradingSettled {
		t.Error("trading should still settle even when arbitrage fails")
	}
}
