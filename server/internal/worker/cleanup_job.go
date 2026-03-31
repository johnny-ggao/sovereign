package worker

import (
	"context"
	"log/slog"
	"time"

	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	premiumRepo "github.com/sovereign-fund/sovereign/internal/modules/premium/repository"
)

type CleanupJob struct {
	tokenRepo   authRepo.TokenRepository
	premiumRepo premiumRepo.PremiumRepository
	logger      *slog.Logger
}

func NewCleanupJob(
	tr authRepo.TokenRepository,
	pr premiumRepo.PremiumRepository,
	logger *slog.Logger,
) *CleanupJob {
	return &CleanupJob{
		tokenRepo:   tr,
		premiumRepo: pr,
		logger:      logger,
	}
}

func (j *CleanupJob) Name() string {
	return "cleanup_job"
}

func (j *CleanupJob) Run(ctx context.Context) error {
	if err := j.cleanExpiredTokens(ctx); err != nil {
		j.logger.Error("clean expired tokens failed", slog.String("error", err.Error()))
	}

	if err := j.cleanOldPremiumTicks(ctx); err != nil {
		j.logger.Error("clean old premium ticks failed", slog.String("error", err.Error()))
	}

	return nil
}

func (j *CleanupJob) cleanExpiredTokens(ctx context.Context) error {
	if err := j.tokenRepo.DeleteExpired(ctx); err != nil {
		return err
	}
	j.logger.Info("expired tokens cleaned")
	return nil
}

func (j *CleanupJob) cleanOldPremiumTicks(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, -3, 0) // keep 3 months
	if err := j.premiumRepo.DeleteOlderThan(ctx, cutoff); err != nil {
		return err
	}
	j.logger.Info("old premium ticks cleaned", slog.Time("before", cutoff))
	return nil
}
