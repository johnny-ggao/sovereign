package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"gorm.io/gorm"
)

// TwoFAVerifier 验证 2FA 代码的接口，用于提现安全校验
type TwoFAVerifier interface {
	Verify2FA(ctx context.Context, userID, code string) (bool, error)
}

type WalletService interface {
	GetWallets(ctx context.Context, userID string) (*dto.WalletOverview, error)
	GetDepositAddress(ctx context.Context, userID string, req dto.GetDepositAddressRequest) (*dto.DepositAddressResponse, error)
	Withdraw(ctx context.Context, userID string, req dto.WithdrawRequest) (*dto.WithdrawResponse, error)

	AddWhitelistAddress(ctx context.Context, userID string, req dto.AddWhitelistAddressRequest) (*dto.WhitelistAddressResponse, error)
	GetWhitelistAddresses(ctx context.Context, userID string) ([]dto.WhitelistAddressResponse, error)
	RemoveWhitelistAddress(ctx context.Context, userID, addressID string) error

	GetTransactions(ctx context.Context, userID, txType string, page, perPage int) ([]dto.TransactionResponse, int64, error)
	GetTransaction(ctx context.Context, userID, txID string) (*dto.TransactionResponse, error)

	// HandleWebhook 处理钱包服务商回调
	HandleWebhook(ctx context.Context, payload cobo.WebhookPayload) error

	// InitWallets 为新用户创建所有币种钱包
	InitWallets(ctx context.Context, userID string, currencies []string) error

	// InitDepositAddresses 为新用户预生成所有网络的充值地址
	InitDepositAddresses(ctx context.Context, userID string, networks []string) error
}

type walletService struct {
	walletRepo  repository.WalletRepository
	addrRepo    repository.AddressRepository
	txRepo      repository.TransactionRepository
	provider    cobo.WalletProvider
	eventBus    events.Bus
	twoFA       TwoFAVerifier
	logger      *slog.Logger
	cooldown    time.Duration
}

func NewWalletService(
	walletRepo repository.WalletRepository,
	addrRepo repository.AddressRepository,
	txRepo repository.TransactionRepository,
	provider cobo.WalletProvider,
	eventBus events.Bus,
	twoFA TwoFAVerifier,
	logger *slog.Logger,
	cooldown time.Duration,
) WalletService {
	return &walletService{
		walletRepo: walletRepo,
		addrRepo:   addrRepo,
		txRepo:     txRepo,
		provider:   provider,
		eventBus:   eventBus,
		twoFA:      twoFA,
		logger:     logger,
		cooldown:   cooldown,
	}
}

func (s *walletService) GetWallets(ctx context.Context, userID string) (*dto.WalletOverview, error) {
	wallets, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	resp := make([]dto.WalletResponse, 0, len(wallets))
	totalUSDT := decimal.Zero

	for _, w := range wallets {
		resp = append(resp, dto.WalletResponse{
			ID:          w.ID,
			Currency:    w.Currency,
			Available:   w.Available,
			InOperation: w.InOperation,
			Frozen:      w.Frozen,
			Total:       w.TotalBalance(),
		})
		if w.Currency == "USDT" {
			totalUSDT = totalUSDT.Add(w.TotalBalance())
		}
	}

	return &dto.WalletOverview{
		Wallets:   resp,
		TotalUSDT: totalUSDT,
	}, nil
}

func (s *walletService) GetDepositAddress(ctx context.Context, userID string, req dto.GetDepositAddressRequest) (*dto.DepositAddressResponse, error) {
	existing, err := s.addrRepo.FindDepositAddress(ctx, userID, req.Currency, req.Network)
	if err == nil {
		return &dto.DepositAddressResponse{
			Currency: existing.Currency,
			Network:  existing.Network,
			Address:  existing.Address,
		}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	result, err := s.provider.GenerateAddress(ctx, cobo.GenerateAddressReq{
		Currency: req.Currency,
		Network:  req.Network,
		Label:    fmt.Sprintf("user_%s", userID),
	})
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("generate address: %w", err))
	}

	addr := &model.DepositAddress{
		UserID:   userID,
		Currency: req.Currency,
		Network:  req.Network,
		Address:  result.Address,
	}
	if err := s.addrRepo.CreateDepositAddress(ctx, addr); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	s.logger.Info("deposit address generated",
		slog.String("user_id", userID),
		slog.String("currency", req.Currency),
		slog.String("network", req.Network),
	)

	return &dto.DepositAddressResponse{
		Currency: req.Currency,
		Network:  req.Network,
		Address:  result.Address,
	}, nil
}

func (s *walletService) Withdraw(ctx context.Context, userID string, req dto.WithdrawRequest) (*dto.WithdrawResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperr.New(400, "INVALID_AMOUNT", "invalid withdrawal amount")
	}

	// 验证 2FA（如果用户已启用）
	if s.twoFA != nil {
		valid, err := s.twoFA.Verify2FA(ctx, userID, req.TwoFACode)
		if err != nil {
			return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("verify 2fa: %w", err))
		}
		if !valid {
			return nil, apperr.New(403, "INVALID_2FA", "invalid two-factor authentication code")
		}
	}

	whiteAddr, err := s.addrRepo.FindWithdrawAddress(ctx, userID, req.Address, req.Network)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrAddressNotWhitelisted
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if whiteAddr.InCooldown() {
		return nil, apperr.ErrAddressCooldown
	}

	wallet, err := s.walletRepo.FindByUserIDAndCurrency(ctx, userID, req.Currency)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInsufficientFunds
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if wallet.Available.LessThan(amount) {
		return nil, apperr.ErrInsufficientFunds
	}

	newAvailable := wallet.Available.Sub(amount)
	newFrozen := wallet.Frozen.Add(amount)
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, newFrozen); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	tx := &model.Transaction{
		UserID:   userID,
		Type:     model.TxTypeWithdraw,
		Currency: req.Currency,
		Network:  req.Network,
		Amount:   amount,
		Address:  req.Address,
		Status:   model.TxStatusPending,
	}
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	result, err := s.provider.Withdraw(ctx, cobo.WithdrawReq{
		Currency:  req.Currency,
		Network:   req.Network,
		Address:   req.Address,
		Amount:    amount,
		RequestID: tx.ID,
	})
	if err != nil {
		s.txRepo.UpdateStatus(ctx, tx.ID, model.TxStatusFailed, "")
		s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Available, wallet.InOperation, wallet.Frozen)
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("withdraw: %w", err))
	}

	s.txRepo.UpdateStatus(ctx, tx.ID, model.TxStatusProcessing, "")

	s.eventBus.Publish(ctx, events.Event{
		Type: events.WithdrawRequested,
		Payload: map[string]string{
			"user_id":        userID,
			"transaction_id": tx.ID,
			"external_id":    result.ExternalID,
		},
	})

	s.logger.Info("withdrawal initiated",
		slog.String("user_id", userID),
		slog.String("tx_id", tx.ID),
		slog.String("currency", req.Currency),
		slog.String("amount", amount.String()),
	)

	return &dto.WithdrawResponse{
		TransactionID: tx.ID,
		Status:        "processing",
		Message:       "withdrawal request submitted",
	}, nil
}

func (s *walletService) AddWhitelistAddress(ctx context.Context, userID string, req dto.AddWhitelistAddressRequest) (*dto.WhitelistAddressResponse, error) {
	addr := &model.WithdrawAddress{
		UserID:        userID,
		Currency:      req.Currency,
		Network:       req.Network,
		Address:       req.Address,
		Label:         req.Label,
		CooldownUntil: time.Now().Add(s.cooldown),
		IsActive:      true,
	}

	if err := s.addrRepo.CreateWithdrawAddress(ctx, addr); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	s.logger.Info("whitelist address added",
		slog.String("user_id", userID),
		slog.String("address", req.Address),
	)

	return &dto.WhitelistAddressResponse{
		ID:            addr.ID,
		Currency:      addr.Currency,
		Network:       addr.Network,
		Address:       addr.Address,
		Label:         addr.Label,
		CooldownUntil: addr.CooldownUntil.Format(time.RFC3339),
		IsActive:      addr.IsActive,
	}, nil
}

func (s *walletService) GetWhitelistAddresses(ctx context.Context, userID string) ([]dto.WhitelistAddressResponse, error) {
	addrs, err := s.addrRepo.FindWithdrawAddresses(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	resp := make([]dto.WhitelistAddressResponse, 0, len(addrs))
	for _, a := range addrs {
		resp = append(resp, dto.WhitelistAddressResponse{
			ID:            a.ID,
			Currency:      a.Currency,
			Network:       a.Network,
			Address:       a.Address,
			Label:         a.Label,
			CooldownUntil: a.CooldownUntil.Format(time.RFC3339),
			IsActive:      a.IsActive,
		})
	}
	return resp, nil
}

func (s *walletService) RemoveWhitelistAddress(ctx context.Context, userID, addressID string) error {
	if err := s.addrRepo.DeleteWithdrawAddress(ctx, addressID, userID); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	return nil
}

func (s *walletService) GetTransactions(ctx context.Context, userID, txType string, page, perPage int) ([]dto.TransactionResponse, int64, error) {
	offset := (page - 1) * perPage
	txs, total, err := s.txRepo.FindByUserID(ctx, userID, txType, perPage, offset)
	if err != nil {
		return nil, 0, apperr.Wrap(apperr.ErrInternal, err)
	}

	resp := make([]dto.TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		r := dto.TransactionResponse{
			ID:        tx.ID,
			Type:      tx.Type,
			Currency:  tx.Currency,
			Network:   tx.Network,
			Amount:    tx.Amount,
			Fee:       tx.Fee,
			Address:   tx.Address,
			TxHash:    tx.TxHash,
			Status:    tx.Status,
			CreatedAt: tx.CreatedAt.Format(time.RFC3339),
		}
		if tx.ConfirmedAt != nil {
			t := tx.ConfirmedAt.Format(time.RFC3339)
			r.ConfirmedAt = &t
		}
		resp = append(resp, r)
	}

	return resp, total, nil
}

func (s *walletService) GetTransaction(ctx context.Context, userID, txID string) (*dto.TransactionResponse, error) {
	tx, err := s.txRepo.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if tx.UserID != userID {
		return nil, apperr.ErrNotFound
	}

	r := &dto.TransactionResponse{
		ID:        tx.ID,
		Type:      tx.Type,
		Currency:  tx.Currency,
		Network:   tx.Network,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Address:   tx.Address,
		TxHash:    tx.TxHash,
		Status:    tx.Status,
		CreatedAt: tx.CreatedAt.Format(time.RFC3339),
	}
	if tx.ConfirmedAt != nil {
		t := tx.ConfirmedAt.Format(time.RFC3339)
		r.ConfirmedAt = &t
	}
	return r, nil
}

func (s *walletService) HandleWebhook(ctx context.Context, payload cobo.WebhookPayload) error {
	tx, err := s.txRepo.FindByExternalID(ctx, payload.RequestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if payload.Type == "deposit" {
				return s.handleDepositWebhook(ctx, payload)
			}
			return nil
		}
		return err
	}

	mappedStatus := payload.Status
	// Cobo 原始状态已在 WebhookPayload 解析时映射，但做安全校验
	validStatuses := map[string]bool{
		model.TxStatusPending:    true,
		model.TxStatusProcessing: true,
		model.TxStatusConfirmed:  true,
		model.TxStatusFailed:     true,
		model.TxStatusCancelled:  true,
	}
	if !validStatuses[mappedStatus] {
		mappedStatus = model.TxStatusPending
	}

	if err := s.txRepo.UpdateStatus(ctx, tx.ID, mappedStatus, payload.TxHash); err != nil {
		return err
	}

	if mappedStatus == model.TxStatusConfirmed && tx.Type == model.TxTypeWithdraw {
		wallet, err := s.walletRepo.FindByUserIDAndCurrency(ctx, tx.UserID, tx.Currency)
		if err != nil {
			return err
		}
		newFrozen := wallet.Frozen.Sub(tx.Amount)
		if newFrozen.LessThan(decimal.Zero) {
			newFrozen = decimal.Zero
		}
		if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Available, wallet.InOperation, newFrozen); err != nil {
			return err
		}

		s.eventBus.Publish(ctx, events.Event{
			Type:    events.WithdrawCompleted,
			Payload: map[string]string{"user_id": tx.UserID, "transaction_id": tx.ID},
		})
	}

	if mappedStatus == model.TxStatusFailed && tx.Type == model.TxTypeWithdraw {
		wallet, err := s.walletRepo.FindByUserIDAndCurrency(ctx, tx.UserID, tx.Currency)
		if err != nil {
			return err
		}
		newAvailable := wallet.Available.Add(tx.Amount)
		newFrozen := wallet.Frozen.Sub(tx.Amount)
		if newFrozen.LessThan(decimal.Zero) {
			newFrozen = decimal.Zero
		}
		if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, newFrozen); err != nil {
			return err
		}

		s.eventBus.Publish(ctx, events.Event{
			Type:    events.WithdrawFailed,
			Payload: map[string]string{"user_id": tx.UserID, "transaction_id": tx.ID},
		})
	}

	return nil
}

func (s *walletService) handleDepositWebhook(ctx context.Context, payload cobo.WebhookPayload) error {
	// 通过充值地址反查用户
	depositAddr, err := s.addrRepo.FindDepositAddressByAddress(ctx, payload.Address)
	if err != nil {
		s.logger.Warn("deposit address not found",
			slog.String("address", payload.Address),
			slog.String("error", err.Error()),
		)
		return nil
	}

	tx := &model.Transaction{
		UserID:     depositAddr.UserID,
		Type:       model.TxTypeDeposit,
		Currency:   payload.Currency,
		Network:    depositAddr.Network,
		Amount:     payload.Amount,
		Fee:        payload.Fee,
		TxHash:     payload.TxHash,
		Address:    payload.Address,
		Status:     payload.Status,
		ExternalID: payload.ID,
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return err
	}

	// 充值确认后更新钱包余额
	if payload.Status == "confirmed" {
		wallet, err := s.walletRepo.FindOrCreate(ctx, depositAddr.UserID, payload.Currency)
		if err != nil {
			return fmt.Errorf("find wallet for deposit: %w", err)
		}

		newAvailable := wallet.Available.Add(payload.Amount)
		if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, wallet.Frozen); err != nil {
			return fmt.Errorf("update balance for deposit: %w", err)
		}

		s.eventBus.Publish(ctx, events.Event{
			Type: events.DepositConfirmed,
			Payload: map[string]string{
				"user_id":        depositAddr.UserID,
				"transaction_id": tx.ID,
				"currency":       payload.Currency,
				"amount":         payload.Amount.String(),
			},
		})

		s.logger.Info("deposit confirmed",
			slog.String("user_id", depositAddr.UserID),
			slog.String("currency", payload.Currency),
			slog.String("amount", payload.Amount.String()),
		)
	}

	return nil
}

func (s *walletService) InitWallets(ctx context.Context, userID string, currencies []string) error {
	for _, currency := range currencies {
		if _, err := s.walletRepo.FindOrCreate(ctx, userID, currency); err != nil {
			return apperr.Wrap(apperr.ErrInternal, fmt.Errorf("init wallet %s: %w", currency, err))
		}
	}
	return nil
}

func (s *walletService) InitDepositAddresses(ctx context.Context, userID string, networks []string) error {
	for _, network := range networks {
		_, err := s.GetDepositAddress(ctx, userID, dto.GetDepositAddressRequest{
			Currency: "USDT",
			Network:  network,
		})
		if err != nil {
			s.logger.Error("init deposit address failed",
				slog.String("user_id", userID),
				slog.String("network", network),
				slog.String("error", err.Error()),
			)
			// 不中断，继续生成下一个网络的地址
		}
	}
	return nil
}
