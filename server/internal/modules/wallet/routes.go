package wallet

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/handler"
	"github.com/sovereign-fund/sovereign/internal/shared/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.WalletHandler) {
	wallets := rg.Group("/wallets")
	{
		wallets.GET("", h.GetWallets)
		wallets.POST("/deposit-address", h.GetDepositAddress)

		withdraw := wallets.Group("", middleware.RateLimit(10, time.Minute))
		{
			withdraw.POST("/withdraw", h.Withdraw)
		}

		wallets.GET("/addresses", h.GetWhitelistAddresses)
		wallets.POST("/addresses", h.AddWhitelistAddress)
		wallets.DELETE("/addresses/:id", h.RemoveWhitelistAddress)
	}

	txs := rg.Group("/transactions")
	{
		txs.GET("", h.GetTransactions)
		txs.GET("/:id", h.GetTransaction)
	}
}

func RegisterWebhookRoutes(rg *gin.RouterGroup, h *handler.WalletHandler) {
	rg.POST("/webhooks/wallet", h.HandleWebhook)
}
