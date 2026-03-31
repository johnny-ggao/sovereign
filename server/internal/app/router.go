package app

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/auth"
	"github.com/sovereign-fund/sovereign/internal/modules/dashboard"
	"github.com/sovereign-fund/sovereign/internal/modules/investment"
	"github.com/sovereign-fund/sovereign/internal/modules/premium"
	"github.com/sovereign-fund/sovereign/internal/modules/settings"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet"
	"github.com/sovereign-fund/sovereign/internal/shared/middleware"
)

func SetupRouter(a *App, ctx context.Context) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(
		middleware.Recovery(a.Logger),
		middleware.RequestID(),
		middleware.Logger(a.Logger),
		middleware.CORS(),
		middleware.RateLimitWithContext(ctx, 100, time.Second),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// Public routes
	auth.RegisterRoutes(v1, a.AuthModule.Handler, a.JWTManager)
	premium.RegisterRoutes(v1, a.PremiumModule.Handler)

	// Protected routes
	protected := v1.Group("", middleware.Auth(a.JWTManager))
	{
		dashboard.RegisterRoutes(protected, a.DashboardModule.Handler)
		wallet.RegisterRoutes(protected, a.WalletModule.Handler)
		investment.RegisterRoutes(protected, a.InvestmentModule.Handler)
		tradelog.RegisterRoutes(protected, a.TradeLogModule.Handler)
		settlement.RegisterRoutes(protected, a.SettlementModule.Handler)
		settings.RegisterRoutes(protected, a.SettingsModule.Handler)
	}

	// WebSocket routes
	ws := r.Group("/ws/v1")
	premium.RegisterWSRoutes(ws, a.PremiumModule.WSHandler)

	// Webhook routes (signature verified internally)
	webhooks := r.Group("/api/v1")
	wallet.RegisterWebhookRoutes(webhooks, a.WalletModule.Handler)

	// Internal API（供交易机器人推送套利记录）
	internal := r.Group("/api/v1/internal")
	tradelog.RegisterInternalRoutes(internal, a.TradeLogModule.Handler)

	return r
}
