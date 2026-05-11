package investment

import (
	"log/slog"

	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/service"
	walletRepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"gorm.io/gorm"
)

type Module struct {
	Handler *handler.InvestmentHandler
	Service service.InvestmentService
}

func NewModule(db *gorm.DB, bus events.Bus, logger *slog.Logger) *Module {
	invRepo := repository.NewInvestmentRepository(db)
	wr := walletRepo.NewWalletRepository(db)
	ur := authRepo.NewUserRepository(db)
	svc := service.NewInvestmentService(invRepo, wr, ur, bus, logger)
	h := handler.NewInvestmentHandler(svc)

	return &Module{Handler: h, Service: svc}
}
