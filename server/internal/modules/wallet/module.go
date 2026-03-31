package wallet

import (
	"log/slog"
	"time"

	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/service"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"gorm.io/gorm"
)

type Module struct {
	Handler  *handler.WalletHandler
	Service  service.WalletService
	Provider cobo.WalletProvider
}

func NewModule(db *gorm.DB, provider cobo.WalletProvider, bus events.Bus, logger *slog.Logger, cooldown time.Duration) *Module {
	walletRepo := repository.NewWalletRepository(db)
	addrRepo := repository.NewAddressRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	userRepo := authRepo.NewUserRepository(db)
	twoFA := service.NewTOTPVerifier(userRepo)

	walletSvc := service.NewWalletService(walletRepo, addrRepo, txRepo, provider, bus, twoFA, logger, cooldown)
	walletHandler := handler.NewWalletHandler(walletSvc, provider)

	return &Module{
		Handler:  walletHandler,
		Service:  walletSvc,
		Provider: provider,
	}
}
