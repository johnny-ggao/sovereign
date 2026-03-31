package settings

import (
	"log/slog"

	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/service"
	"gorm.io/gorm"
)

type Module struct {
	Handler *handler.SettingsHandler
	Service service.SettingsService
}

func NewModule(db *gorm.DB, logger *slog.Logger) *Module {
	ur := authRepo.NewUserRepository(db)
	sr := repository.NewSettingsRepository(db)
	svc := service.NewSettingsService(ur, sr, logger)
	h := handler.NewSettingsHandler(svc)

	return &Module{Handler: h, Service: svc}
}
