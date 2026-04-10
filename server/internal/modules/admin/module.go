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
	AuthHandler        *handler.AuthHandler
	AdminUserHandler   *handler.AdminUserHandler
	AuditHandler       *handler.AuditHandler
	UserHandler        *handler.UserHandler
	DashboardHandler   *handler.DashboardHandler
	TradeHandler       *handler.TradeHandler
	TransactionHandler *handler.TransactionHandler
	AuditService       service.AuditService
	AdminRepo          repository.AdminRepository
	JWTSecret          string
}

func NewModule(db *gorm.DB, cfg config.AdminConfig, logger *slog.Logger) *Module {
	repo := repository.NewAdminRepository(db)

	authSvc := service.NewAuthService(repo, cfg.JWTSecret, cfg.JWTExpiry, logger)
	adminUserSvc := service.NewAdminUserService(repo, logger)
	userSvc := service.NewUserService(db, logger)
	auditSvc := service.NewAuditService(db)
	dashboardSvc := service.NewDashboardService(db, logger)
	tradeSvc := service.NewTradeService(db)
	transactionSvc := service.NewTransactionService(db)

	return &Module{
		AuthHandler:        handler.NewAuthHandler(authSvc),
		AdminUserHandler:   handler.NewAdminUserHandler(adminUserSvc, auditSvc),
		AuditHandler:       handler.NewAuditHandler(auditSvc),
		UserHandler:        handler.NewUserHandler(userSvc, auditSvc),
		DashboardHandler:   handler.NewDashboardHandler(dashboardSvc),
		TradeHandler:       handler.NewTradeHandler(tradeSvc, auditSvc),
		TransactionHandler: handler.NewTransactionHandler(transactionSvc),
		AuditService:       auditSvc,
		AdminRepo:          repo,
		JWTSecret:          cfg.JWTSecret,
	}
}
