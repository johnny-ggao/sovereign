package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/dashboard/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/dashboard/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type DashboardHandler struct {
	dashSvc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashSvc: svc}
}

func (h *DashboardHandler) GetSummary(c *gin.Context) {
	userID := c.GetString("user_id")

	summary, err := h.dashSvc.GetSummary(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, summary)
}

func (h *DashboardHandler) GetPerformance(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.PerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	perf, err := h.dashSvc.GetPerformance(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, perf)
}

func handleError(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
