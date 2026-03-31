package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	walletRepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"gorm.io/gorm"
)

var minInvestAmount = decimal.NewFromInt(100)

type InvestmentService interface {
	Create(ctx context.Context, userID string, req dto.CreateInvestmentRequest) (*dto.InvestmentResponse, error)
	GetAll(ctx context.Context, userID string) (*dto.InvestmentListResponse, error)
	GetByID(ctx context.Context, userID, id string) (*dto.InvestmentResponse, error)
	Redeem(ctx context.Context, userID string, req dto.RedeemRequest) (*dto.InvestmentResponse, error)
}

type investmentService struct {
	invRepo    repository.InvestmentRepository
	walletRepo walletRepo.WalletRepository
	eventBus   events.Bus
	logger     *slog.Logger
}

func NewInvestmentService(
	invRepo repository.InvestmentRepository,
	wr walletRepo.WalletRepository,
	bus events.Bus,
	logger *slog.Logger,
) InvestmentService {
	return &investmentService{
		invRepo:    invRepo,
		walletRepo: wr,
		eventBus:   bus,
		logger:     logger,
	}
}

func (s *investmentService) Create(ctx context.Context, userID string, req dto.CreateInvestmentRequest) (*dto.InvestmentResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperr.New(400, "INVALID_AMOUNT", "invalid investment amount")
	}

	if amount.LessThan(minInvestAmount) {
		return nil, apperr.ErrMinInvestment
	}

	currency := req.Currency
	if currency == "" {
		currency = "USDT"
	}

	wallet, err := s.walletRepo.FindByUserIDAndCurrency(ctx, userID, currency)
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
	newInOp := wallet.InOperation.Add(amount)
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, newInOp, wallet.Frozen); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	inv := &model.Investment{
		UserID:   userID,
		Amount:   amount,
		Currency: currency,
		Status:   model.InvestStatusActive,
	}

	if err := s.invRepo.Create(ctx, inv); err != nil {
		s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Available, wallet.InOperation, wallet.Frozen)
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("create investment: %w", err))
	}

	s.eventBus.Publish(ctx, events.Event{
		Type:    events.InvestmentCreated,
		Payload: map[string]string{"user_id": userID, "investment_id": inv.ID, "amount": amount.String()},
	})

	s.logger.Info("investment created",
		slog.String("user_id", userID),
		slog.String("investment_id", inv.ID),
		slog.String("amount", amount.String()),
	)

	return toInvestmentResponse(inv), nil
}

func (s *investmentService) GetAll(ctx context.Context, userID string) (*dto.InvestmentListResponse, error) {
	invs, err := s.invRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	resp := make([]dto.InvestmentResponse, 0, len(invs))
	summary := dto.InvestmentSummary{}

	for _, inv := range invs {
		resp = append(resp, *toInvestmentResponse(&inv))
		summary.TotalReturn = summary.TotalReturn.Add(inv.NetReturn)
		if inv.Status == model.InvestStatusActive {
			summary.TotalInvested = summary.TotalInvested.Add(inv.Amount)
			summary.ActiveCount++
		}
	}

	return &dto.InvestmentListResponse{
		Investments: resp,
		Summary:     summary,
	}, nil
}

func (s *investmentService) GetByID(ctx context.Context, userID, id string) (*dto.InvestmentResponse, error) {
	inv, err := s.invRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInvestmentNotFound
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	if inv.UserID != userID {
		return nil, apperr.ErrInvestmentNotFound
	}
	return toInvestmentResponse(inv), nil
}

func (s *investmentService) Redeem(ctx context.Context, userID string, req dto.RedeemRequest) (*dto.InvestmentResponse, error) {
	inv, err := s.invRepo.FindByID(ctx, req.InvestmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInvestmentNotFound
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	if inv.UserID != userID {
		return nil, apperr.ErrInvestmentNotFound
	}

	if inv.Status != model.InvestStatusActive {
		return nil, apperr.ErrRedeemPending
	}

	// 更新钱包余额：本金从 InOperation 退回 + 净收益到 Available
	wallet, err := s.walletRepo.FindByUserIDAndCurrency(ctx, userID, inv.Currency)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	// 本金退回 + 净收益
	returnAmount := inv.Amount.Add(inv.NetReturn)
	newAvailable := wallet.Available.Add(returnAmount)
	newInOp := wallet.InOperation.Sub(inv.Amount)
	if newInOp.LessThan(decimal.Zero) {
		newInOp = decimal.Zero
	}

	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, newInOp, wallet.Frozen); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	now := time.Now()
	inv.Status = model.InvestStatusRedeemed
	inv.EndDate = &now

	if err := s.invRepo.Update(ctx, inv); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	s.eventBus.Publish(ctx, events.Event{
		Type:    events.InvestmentRedeemed,
		Payload: map[string]string{"user_id": userID, "investment_id": inv.ID},
	})

	s.logger.Info("investment redeemed",
		slog.String("user_id", userID),
		slog.String("investment_id", inv.ID),
		slog.String("principal", inv.Amount.String()),
		slog.String("net_return", inv.NetReturn.String()),
		slog.String("total_returned", returnAmount.String()),
	)

	return toInvestmentResponse(inv), nil
}

func toInvestmentResponse(inv *model.Investment) *dto.InvestmentResponse {
	r := &dto.InvestmentResponse{
		ID:             inv.ID,
		Amount:         inv.Amount,
		Currency:       inv.Currency,
		Status:         inv.Status,
		TotalReturn:    inv.TotalReturn,
		PerformanceFee: inv.PerformanceFee,
		NetReturn:      inv.NetReturn,
		StartDate:      inv.StartDate.Format(time.RFC3339),
	}

	if inv.Amount.GreaterThan(decimal.Zero) {
		r.ReturnPct = inv.NetReturn.Div(inv.Amount).Mul(decimal.NewFromInt(100))
	}

	if inv.EndDate != nil {
		t := inv.EndDate.Format(time.RFC3339)
		r.EndDate = &t
	}

	return r
}
