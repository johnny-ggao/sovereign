package premium

import (
	"log/slog"

	"github.com/sovereign-fund/sovereign/internal/modules/premium/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/service"
	"gorm.io/gorm"
)

type Module struct {
	Handler   *handler.PremiumHandler
	WSHandler *handler.WSHandler
	Service   service.PremiumService
	Hub       *service.Hub
}

func NewModule(db *gorm.DB, logger *slog.Logger) *Module {
	repo := repository.NewPremiumRepository(db)
	hub := service.NewHub(logger)
	svc := service.NewPremiumService(repo, hub, logger)
	h := handler.NewPremiumHandler(svc)
	wsh := handler.NewWSHandler(hub, logger)

	return &Module{
		Handler:   h,
		WSHandler: wsh,
		Service:   svc,
		Hub:       hub,
	}
}
