package premium

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/handler"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.PremiumHandler) {
	p := rg.Group("/premium")
	{
		p.GET("/latest", h.GetLatest)
		p.GET("/history", h.GetHistory)
	}
}

func RegisterWSRoutes(rg *gin.RouterGroup, wsh *handler.WSHandler) {
	rg.GET("/premium", wsh.HandleWebSocket)
}
