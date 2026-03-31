package settlement

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement/handler"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.SettlementHandler) {
	s := rg.Group("/settlements")
	{
		s.GET("", h.GetAll)
		s.GET("/:id", h.GetByID)
	}
}
