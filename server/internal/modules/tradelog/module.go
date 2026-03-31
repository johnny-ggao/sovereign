package tradelog

import (
	"log/slog"

	investRepo "github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/service"
	"gorm.io/gorm"
)

type Module struct {
	Handler *handler.TradeHandler
	Service service.TradeService
}

func NewModule(db *gorm.DB, logger *slog.Logger) *Module {
	repo := repository.NewTradeRepository(db)
	ir := investRepo.NewInvestmentRepository(db)
	svc := service.NewTradeService(repo, ir, logger)
	h := handler.NewTradeHandler(svc)

	return &Module{Handler: h, Service: svc}
}
