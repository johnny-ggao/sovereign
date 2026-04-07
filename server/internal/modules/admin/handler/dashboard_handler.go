package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DASHBOARD_STATS_FAILED", err.Error())
		return
	}

	response.OK(c, stats)
}
