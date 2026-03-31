package settlement

import (
	"log/slog"

	"github.com/sovereign-fund/sovereign/internal/modules/settlement/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement/service"
	"gorm.io/gorm"
)

type Module struct {
	Handler *handler.SettlementHandler
	Service service.SettlementService
}

func NewModule(db *gorm.DB, logger *slog.Logger) *Module {
	repo := repository.NewSettlementRepository(db)
	svc := service.NewSettlementService(repo, logger)
	h := handler.NewSettlementHandler(svc)

	return &Module{Handler: h, Service: svc}
}
