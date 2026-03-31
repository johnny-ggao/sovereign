package tradelog

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/handler"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.TradeHandler) {
	trades := rg.Group("/trades")
	{
		trades.GET("", h.GetTrades)
		trades.GET("/export", h.ExportCSV)
	}
}

// RegisterInternalRoutes 内部 API，供交易机器人/外部系统推送套利交易记录
func RegisterInternalRoutes(rg *gin.RouterGroup, h *handler.TradeHandler) {
	trades := rg.Group("/trades")
	{
		trades.POST("", h.CreateTrade)
		trades.POST("/batch", h.BatchCreateTrades)
	}
}
