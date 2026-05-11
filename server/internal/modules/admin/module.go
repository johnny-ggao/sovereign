package admin

import (
	"log/slog"

	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	authrepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	walletrepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"gorm.io/gorm"
)

type Module struct {
	AuthHandler           *handler.AuthHandler
	AdminUserHandler      *handler.AdminUserHandler
	AuditHandler          *handler.AuditHandler
	UserHandler           *handler.UserHandler
	DashboardHandler      *handler.DashboardHandler
	TradeHandler          *handler.TradeHandler
	TradingTradeHandler   *handler.TradingTradeHandler
	TransactionHandler    *handler.TransactionHandler
	WithdrawReviewHandler *handler.WithdrawReviewHandler
	UserProductHandler    *handler.UserProductHandler
	AuditService          service.AuditService
	AdminRepo             repository.AdminRepository
	JWTSecret             string
}

func NewModule(
	db *gorm.DB,
	cfg config.AdminConfig,
	logger *slog.Logger,
	walletRepo walletrepo.WalletRepository,
	txRepo walletrepo.TransactionRepository,
	coboProvider cobo.WalletProvider,
	bus events.Bus,
) *Module {
	repo := repository.NewAdminRepository(db)

	authSvc := service.NewAuthService(repo, cfg.JWTSecret, cfg.JWTExpiry, logger)
	adminUserSvc := service.NewAdminUserService(repo, logger)
	userSvc := service.NewUserService(db, logger)
	auditSvc := service.NewAuditService(db)
	dashboardSvc := service.NewDashboardService(db, logger)
	tradeSvc := service.NewTradeService(db)
	tradingTradeSvc := service.NewTradingTradeService(db)
	transactionSvc := service.NewTransactionService(db)

	withdrawReviewSvc := service.NewWithdrawReviewService(
		service.NewPGReviewQuerier(db),
		walletRepo, txRepo, coboProvider, bus, logger,
	)

	authUserRepo := authrepo.NewUserRepository(db)
	userProductChangeLogRepo := repository.NewUserProductChangeLogRepository(db)
	userProductSvc := service.NewUserProductService(authUserRepo, userProductChangeLogRepo)

	return &Module{
		AuthHandler:           handler.NewAuthHandler(authSvc),
		AdminUserHandler:      handler.NewAdminUserHandler(adminUserSvc, auditSvc),
		AuditHandler:          handler.NewAuditHandler(auditSvc),
		UserHandler:           handler.NewUserHandler(userSvc, auditSvc),
		DashboardHandler:      handler.NewDashboardHandler(dashboardSvc),
		TradeHandler:          handler.NewTradeHandler(tradeSvc, auditSvc),
		TradingTradeHandler:   handler.NewTradingTradeHandler(tradingTradeSvc, auditSvc),
		TransactionHandler:    handler.NewTransactionHandler(transactionSvc),
		WithdrawReviewHandler: handler.NewWithdrawReviewHandler(withdrawReviewSvc, auditSvc),
		UserProductHandler:    handler.NewUserProductHandler(userProductSvc, auditSvc),
		AuditService:          auditSvc,
		AdminRepo:             repo,
		JWTSecret:             cfg.JWTSecret,
	}
}
