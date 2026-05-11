package service

import (
	"context"
	"fmt"

	logmodel "github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	logrepo "github.com/sovereign-fund/sovereign/internal/modules/admin/repository"
	usermodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	usermodelrepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type UserProductService interface {
	Change(ctx context.Context, userID, toType, reason, adminID, adminEmail string) error
	BulkChange(ctx context.Context, userIDs []string, toType, reason, adminID, adminEmail string) (*BulkChangeResult, error)
	History(ctx context.Context, userID string) ([]logmodel.UserProductChangeLog, error)
}

type BulkChangeResult struct {
	Succeeded []string                `json:"succeeded"`
	Failed    []BulkChangeFailedEntry `json:"failed"`
}

type BulkChangeFailedEntry struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type userProductService struct {
	uRepo usermodelrepo.UserRepository
	lRepo logrepo.UserProductChangeLogRepository
}

func NewUserProductService(uRepo usermodelrepo.UserRepository, lRepo logrepo.UserProductChangeLogRepository) UserProductService {
	return &userProductService{uRepo: uRepo, lRepo: lRepo}
}

func (s *userProductService) Change(ctx context.Context, userID, toType, reason, adminID, adminEmail string) error {
	if toType != usermodel.InvestmentTypeArbitrage && toType != usermodel.InvestmentTypeTrading {
		return apperr.ErrInvalidProductType
	}
	user, err := s.uRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.ErrAccountNotFound
	}
	from := user.InvestmentType
	if from == "" {
		from = usermodel.InvestmentTypeArbitrage
	}
	if from == toType {
		return apperr.ErrSameProductType
	}
	user.InvestmentType = toType
	if err := s.uRepo.Update(ctx, user); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if err := s.lRepo.Create(ctx, &logmodel.UserProductChangeLog{
		UserID:     userID,
		FromType:   from,
		ToType:     toType,
		AdminID:    adminID,
		AdminEmail: adminEmail,
		Reason:     reason,
	}); err != nil {
		return apperr.Wrap(apperr.ErrInternal, fmt.Errorf("write change log: %w", err))
	}
	return nil
}

func (s *userProductService) BulkChange(ctx context.Context, userIDs []string, toType, reason, adminID, adminEmail string) (*BulkChangeResult, error) {
	res := &BulkChangeResult{}
	for _, uid := range userIDs {
		if err := s.Change(ctx, uid, toType, reason, adminID, adminEmail); err != nil {
			res.Failed = append(res.Failed, BulkChangeFailedEntry{UserID: uid, Reason: err.Error()})
			continue
		}
		res.Succeeded = append(res.Succeeded, uid)
	}
	return res, nil
}

func (s *userProductService) History(ctx context.Context, userID string) ([]logmodel.UserProductChangeLog, error) {
	return s.lRepo.ListByUser(ctx, userID)
}
