package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/dashboard/handler"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.DashboardHandler) {
	d := rg.Group("/dashboard")
	{
		d.GET("/summary", h.GetSummary)
		d.GET("/performance", h.GetPerformance)
	}
}
