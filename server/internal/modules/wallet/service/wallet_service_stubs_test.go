package service

import (
	"context"
	"errors"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"gorm.io/gorm"
)

// ---- stubWalletRepo -----------------------------------------------------

type stubWalletRepo struct {
	wallet *model.Wallet

	updateCalls    int
	lastAvailable  decimal.Decimal
	lastInOp       decimal.Decimal
	lastFrozen     decimal.Decimal
}

func (s *stubWalletRepo) FindByUserID(context.Context, string) ([]model.Wallet, error) {
	panic("stubWalletRepo.FindByUserID: unused")
}

func (s *stubWalletRepo) FindByUserIDAndCurrency(_ context.Context, userID, currency string) (*model.Wallet, error) {
	if s.wallet == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if s.wallet.UserID != userID || s.wallet.Currency != currency {
		return nil, gorm.ErrRecordNotFound
	}
	return s.wallet, nil
}

func (s *stubWalletRepo) FindOrCreate(context.Context, string, string) (*model.Wallet, error) {
	panic("stubWalletRepo.FindOrCreate: unused")
}

func (s *stubWalletRepo) UpdateBalance(_ context.Context, id string, available, inOperation, frozen decimal.Decimal) error {
	s.updateCalls++
	s.lastAvailable = available
	s.lastInOp = inOperation
	s.lastFrozen = frozen
	if s.wallet != nil && s.wallet.ID == id {
		s.wallet.Available = available
		s.wallet.InOperation = inOperation
		s.wallet.Frozen = frozen
	}
	return nil
}

func (s *stubWalletRepo) AddEarnings(context.Context, string, decimal.Decimal, string) error {
	panic("stubWalletRepo.AddEarnings: unused")
}

func (s *stubWalletRepo) ClaimEarnings(context.Context, string) error {
	panic("stubWalletRepo.ClaimEarnings: unused")
}

var _ repository.WalletRepository = (*stubWalletRepo)(nil)

// ---- stubAddressRepo ----------------------------------------------------

type stubAddressRepo struct {
	whitelist *model.WithdrawAddress
}

func (s *stubAddressRepo) FindDepositAddress(context.Context, string, string, string) (*model.DepositAddress, error) {
	panic("stubAddressRepo.FindDepositAddress: unused")
}

func (s *stubAddressRepo) FindDepositAddressByAddress(context.Context, string) (*model.DepositAddress, error) {
	panic("stubAddressRepo.FindDepositAddressByAddress: unused")
}

func (s *stubAddressRepo) CreateDepositAddress(context.Context, *model.DepositAddress) error {
	panic("stubAddressRepo.CreateDepositAddress: unused")
}

func (s *stubAddressRepo) FindWithdrawAddresses(context.Context, string) ([]model.WithdrawAddress, error) {
	panic("stubAddressRepo.FindWithdrawAddresses: unused")
}

func (s *stubAddressRepo) FindWithdrawAddress(_ context.Context, userID, address, network string) (*model.WithdrawAddress, error) {
	if s.whitelist == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if s.whitelist.UserID != userID || s.whitelist.Address != address || s.whitelist.Network != network {
		return nil, gorm.ErrRecordNotFound
	}
	return s.whitelist, nil
}

func (s *stubAddressRepo) CreateWithdrawAddress(context.Context, *model.WithdrawAddress) error {
	panic("stubAddressRepo.CreateWithdrawAddress: unused")
}

func (s *stubAddressRepo) DeleteWithdrawAddress(context.Context, string, string) error {
	panic("stubAddressRepo.DeleteWithdrawAddress: unused")
}

var _ repository.AddressRepository = (*stubAddressRepo)(nil)

// ---- stubTxRepo ---------------------------------------------------------

type stubTxRepo struct {
	mu sync.Mutex

	created          []*model.Transaction
	byID             map[string]*model.Transaction
	listByUser       []model.Transaction
	listTotal        int64
	lastReviewUpdate map[string]map[string]any

	createErr error
}

func (s *stubTxRepo) Create(_ context.Context, tx *model.Transaction) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if tx.ID == "" {
		tx.ID = "tx-" + tx.UserID + "-" + tx.Currency
	}
	s.created = append(s.created, tx)
	if s.byID == nil {
		s.byID = make(map[string]*model.Transaction)
	}
	s.byID[tx.ID] = tx
	return nil
}

func (s *stubTxRepo) FindByID(_ context.Context, id string) (*model.Transaction, error) {
	if tx, ok := s.byID[id]; ok {
		return tx, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubTxRepo) FindByExternalID(context.Context, string) (*model.Transaction, error) {
	panic("stubTxRepo.FindByExternalID: unused")
}

func (s *stubTxRepo) FindByUserID(_ context.Context, _ string, _ string, _ int, _ int) ([]model.Transaction, int64, error) {
	return s.listByUser, s.listTotal, nil
}

func (s *stubTxRepo) UpdateStatus(_ context.Context, id, status, txHash string) error {
	if tx, ok := s.byID[id]; ok {
		tx.Status = status
		if txHash != "" {
			tx.TxHash = txHash
		}
	}
	return nil
}

func (s *stubTxRepo) UpdateExternalID(_ context.Context, id, externalID string) error {
	if tx, ok := s.byID[id]; ok {
		tx.ExternalID = externalID
	}
	return nil
}

func (s *stubTxRepo) UpdateReview(_ context.Context, id string, fields map[string]any) error {
	if s.lastReviewUpdate == nil {
		s.lastReviewUpdate = make(map[string]map[string]any)
	}
	s.lastReviewUpdate[id] = fields
	return nil
}

var _ repository.TransactionRepository = (*stubTxRepo)(nil)

// ---- mockCoboProvider ---------------------------------------------------

type mockCoboProvider struct {
	withdrawCalls         int
	generateAddressCalls  int
	getBalanceCalls       int
	getTransactionCalls   int
	verifyWebhookCalls    int

	withdrawResp *cobo.WithdrawResp
	withdrawErr  error
}

func (m *mockCoboProvider) GenerateAddress(context.Context, cobo.GenerateAddressReq) (*cobo.GenerateAddressResp, error) {
	m.generateAddressCalls++
	return &cobo.GenerateAddressResp{Address: "mock-addr", ExternalID: "mock-ext"}, nil
}

func (m *mockCoboProvider) Withdraw(_ context.Context, _ cobo.WithdrawReq) (*cobo.WithdrawResp, error) {
	m.withdrawCalls++
	if m.withdrawErr != nil {
		return nil, m.withdrawErr
	}
	if m.withdrawResp != nil {
		return m.withdrawResp, nil
	}
	return &cobo.WithdrawResp{ExternalID: "mock-ext", Status: "pending"}, nil
}

func (m *mockCoboProvider) GetBalance(_ context.Context, currency string) (*cobo.BalanceResp, error) {
	m.getBalanceCalls++
	return &cobo.BalanceResp{Currency: currency, Available: decimal.Zero, Frozen: decimal.Zero}, nil
}

func (m *mockCoboProvider) GetTransaction(_ context.Context, externalID string) (*cobo.TransactionResp, error) {
	m.getTransactionCalls++
	return &cobo.TransactionResp{ExternalID: externalID, Status: "confirmed"}, nil
}

func (m *mockCoboProvider) VerifyWebhook(_ string, _ []byte) (bool, error) {
	m.verifyWebhookCalls++
	return true, nil
}

var _ cobo.WalletProvider = (*mockCoboProvider)(nil)

// ---- spyEventBus --------------------------------------------------------

type spyEventBus struct {
	published []events.Event
}

func (b *spyEventBus) Publish(_ context.Context, event events.Event) {
	b.published = append(b.published, event)
}

func (b *spyEventBus) Subscribe(string, events.Handler) {}

func (b *spyEventBus) Shutdown() {}

var _ events.Bus = (*spyEventBus)(nil)

// ---- stubTwoFA ----------------------------------------------------------

type stubTwoFA struct {
	ok  bool
	err error
}

func (s *stubTwoFA) Verify2FA(context.Context, string, string) (bool, error) {
	return s.ok, s.err
}

var _ TwoFAVerifier = (*stubTwoFA)(nil)

// ---- helpers ------------------------------------------------------------

// errSentinel exists to give tests a sentinel error without pulling fmt.
var errSentinel = errors.New("sentinel")
