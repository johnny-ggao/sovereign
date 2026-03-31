package dashboard

import (
	"log/slog"

	"github.com/sovereign-fund/sovereign/internal/modules/dashboard/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/dashboard/service"
	investRepo "github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	premiumRepo "github.com/sovereign-fund/sovereign/internal/modules/premium/repository"
	settlRepo "github.com/sovereign-fund/sovereign/internal/modules/settlement/repository"
	walletRepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"gorm.io/gorm"
)

type Module struct {
	Handler *handler.DashboardHandler
	Service service.DashboardService
}

func NewModule(db *gorm.DB, logger *slog.Logger) *Module {
	wr := walletRepo.NewWalletRepository(db)
	pr := premiumRepo.NewPremiumRepository(db)
	ir := investRepo.NewInvestmentRepository(db)
	sr := settlRepo.NewSettlementRepository(db)
	svc := service.NewDashboardService(wr, pr, ir, sr, logger)
	h := handler.NewDashboardHandler(svc)

	return &Module{
		Handler: h,
		Service: svc,
	}
}
