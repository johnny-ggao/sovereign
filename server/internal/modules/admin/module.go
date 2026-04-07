package admin

import (
	"log/slog"

	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"gorm.io/gorm"
)

type Module struct {
	AuthHandler      *handler.AuthHandler
	AdminUserHandler *handler.AdminUserHandler
	UserHandler      *handler.UserHandler
	DashboardHandler *handler.DashboardHandler
	TradeHandler     *handler.TradeHandler
	AdminRepo        repository.AdminRepository
	JWTSecret        string
}

func NewModule(db *gorm.DB, cfg config.AdminConfig, logger *slog.Logger) *Module {
	repo := repository.NewAdminRepository(db)

	authSvc := service.NewAuthService(repo, cfg.JWTSecret, cfg.JWTExpiry, logger)
	adminUserSvc := service.NewAdminUserService(repo, logger)
	userSvc := service.NewUserService(db, logger)
	dashboardSvc := service.NewDashboardService(db, logger)
	tradeSvc := service.NewTradeService(db)

	return &Module{
		AuthHandler:      handler.NewAuthHandler(authSvc),
		AdminUserHandler: handler.NewAdminUserHandler(adminUserSvc),
		UserHandler:      handler.NewUserHandler(userSvc),
		DashboardHandler: handler.NewDashboardHandler(dashboardSvc),
		TradeHandler:     handler.NewTradeHandler(tradeSvc),
		AdminRepo:        repo,
		JWTSecret:        cfg.JWTSecret,
	}
}
