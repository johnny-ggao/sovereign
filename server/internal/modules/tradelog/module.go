package tradelog

import (
	"log/slog"

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
	utr := repository.NewUserTradeRepository(db)
	svc := service.NewTradeService(repo, utr, logger)
	h := handler.NewTradeHandler(svc)

	return &Module{Handler: h, Service: svc}
}
