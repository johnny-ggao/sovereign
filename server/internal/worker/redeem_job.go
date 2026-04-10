package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	investModel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	walletRepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"gorm.io/gorm"
)

type RedeemJob struct {
	db         *gorm.DB
	walletRepo walletRepo.WalletRepository
	logger     *slog.Logger
}

func NewRedeemJob(db *gorm.DB, wr walletRepo.WalletRepository, logger *slog.Logger) *RedeemJob {
	return &RedeemJob{
		db:         db,
		walletRepo: wr,
		logger:     logger,
	}
}

func (j *RedeemJob) Name() string {
	return "redeem_job"
}

func (j *RedeemJob) Run(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -3)

	var invs []investModel.Investment
	if err := j.db.WithContext(ctx).
		Where("status = ? AND end_date IS NOT NULL AND end_date <= ?", investModel.InvestStatusStopping, cutoff).
		Find(&invs).Error; err != nil {
		return err
	}

	for i := range invs {
		inv := &invs[i]

		wallet, err := j.walletRepo.FindByUserIDAndCurrency(ctx, inv.UserID, inv.Currency)
		if err != nil {
			j.logger.Error("redeem job wallet lookup failed",
				slog.String("investment_id", inv.ID),
				slog.String("user_id", inv.UserID),
				slog.String("error", err.Error()),
			)
			continue
		}

		returnAmount := inv.Amount.Add(inv.NetReturn)
		newAvailable := wallet.Available.Add(returnAmount)
		newInOperation := wallet.InOperation.Sub(inv.Amount)
		if newInOperation.IsNegative() {
			newInOperation = decimal.Zero
		}

		if err := j.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, newInOperation, wallet.Frozen); err != nil {
			j.logger.Error("redeem job wallet update failed",
				slog.String("investment_id", inv.ID),
				slog.String("user_id", inv.UserID),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := j.db.WithContext(ctx).
			Model(&investModel.Investment{}).
			Where("id = ?", inv.ID).
			Update("status", investModel.InvestStatusRedeemed).Error; err != nil {
			j.logger.Error("redeem job investment update failed",
				slog.String("investment_id", inv.ID),
				slog.String("user_id", inv.UserID),
				slog.String("error", err.Error()),
			)
			continue
		}

		j.logger.Info("redeem job processed investment",
			slog.String("investment_id", inv.ID),
			slog.String("user_id", inv.UserID),
			slog.String("currency", inv.Currency),
			slog.String("principal", inv.Amount.String()),
			slog.String("net_return", inv.NetReturn.String()),
			slog.String("total_returned", returnAmount.String()),
		)
	}

	return nil
}
