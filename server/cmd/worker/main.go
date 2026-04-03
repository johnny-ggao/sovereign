package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sovereign-fund/sovereign/config"
	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	investRepo "github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	premiumRepo "github.com/sovereign-fund/sovereign/internal/modules/premium/repository"
	settlRepo "github.com/sovereign-fund/sovereign/internal/modules/settlement/repository"
	tradeRepo "github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	walletRepository "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/database"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/internal/worker"
	"github.com/sovereign-fund/sovereign/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	db, err := database.NewPostgres(cfg.Database, log)
	if err != nil {
		log.Error("failed to connect postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}

	_, err = database.NewRedis(cfg.Redis, log)
	if err != nil {
		log.Error("failed to connect redis", slog.String("error", err.Error()))
		os.Exit(1)
	}

	bus := events.NewBus(log)

	// Settlement job
	ir := investRepo.NewInvestmentRepository(db)
	tr := tradeRepo.NewTradeRepository(db)
	utr := tradeRepo.NewUserTradeRepository(db)
	sr := settlRepo.NewSettlementRepository(db)
	wr := walletRepository.NewWalletRepository(db)
	settlJob := worker.NewSettlementJob(ir, tr, utr, sr, wr, bus, log)

	// Cleanup job
	tokenRepo := authRepo.NewTokenRepository(db)
	pr := premiumRepo.NewPremiumRepository(db)
	cleanupJob := worker.NewCleanupJob(tokenRepo, pr, log)

	// Register jobs
	w := worker.New(log)
	w.Register(cfg.Worker.SettlementCron, settlJob)
	w.Register(cfg.Worker.CleanupCron, cleanupJob)

	if err := w.Start(); err != nil {
		log.Error("failed to start worker", slog.String("error", err.Error()))
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	w.Stop()
	bus.Shutdown()
}
