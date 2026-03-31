package investment

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/handler"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.InvestmentHandler) {
	inv := rg.Group("/investments")
	{
		inv.POST("", h.Create)
		inv.GET("", h.GetAll)
		inv.GET("/:id", h.GetByID)
		inv.POST("/redeem", h.Redeem)
	}
}
