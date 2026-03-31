package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"gorm.io/gorm"
)

type SettlementService interface {
	GetAll(ctx context.Context, userID string) (*dto.SettlementListResponse, error)
	GetByID(ctx context.Context, userID, id string) (*dto.SettlementResponse, error)
}

type settlementService struct {
	repo   repository.SettlementRepository
	logger *slog.Logger
}

func NewSettlementService(repo repository.SettlementRepository, logger *slog.Logger) SettlementService {
	return &settlementService{repo: repo, logger: logger}
}

func (s *settlementService) GetAll(ctx context.Context, userID string) (*dto.SettlementListResponse, error) {
	settlements, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	resp := make([]dto.SettlementResponse, 0, len(settlements))
	summary := dto.SettlementSummary{}

	for _, st := range settlements {
		resp = append(resp, toSettlementResponse(&st))
		summary.TotalGrossReturn = summary.TotalGrossReturn.Add(st.GrossReturn)
		summary.TotalPerformanceFee = summary.TotalPerformanceFee.Add(st.PerformanceFee)
		summary.TotalNetReturn = summary.TotalNetReturn.Add(st.NetReturn)
		summary.PeriodCount++
	}

	return &dto.SettlementListResponse{
		Settlements: resp,
		Summary:     summary,
	}, nil
}

func (s *settlementService) GetByID(ctx context.Context, userID, id string) (*dto.SettlementResponse, error) {
	st, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	if st.UserID != userID {
		return nil, apperr.ErrNotFound
	}
	r := toSettlementResponse(st)
	return &r, nil
}

func toSettlementResponse(s *model.Settlement) dto.SettlementResponse {
	return dto.SettlementResponse{
		ID:             s.ID,
		InvestmentID:   s.InvestmentID,
		Period:         s.Period,
		GrossReturn:    s.GrossReturn,
		PerformanceFee: s.PerformanceFee,
		FeeRate:        s.FeeRate,
		NetReturn:      s.NetReturn,
		TradeCount:     s.TradeCount,
		AvgPremiumPct:  s.AvgPremiumPct,
		ReportURL:      s.ReportURL,
		SettledAt:      s.SettledAt.Format(time.RFC3339),
	}
}

// CalculateFee 计算绩效费 (50%)
func CalculateFee(grossReturn decimal.Decimal, feeRate decimal.Decimal) (performanceFee, netReturn decimal.Decimal) {
	if grossReturn.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, grossReturn
	}
	performanceFee = grossReturn.Mul(feeRate)
	netReturn = grossReturn.Sub(performanceFee)
	return performanceFee, netReturn
}
