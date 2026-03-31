package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type PremiumHandler struct {
	premiumSvc service.PremiumService
}

func NewPremiumHandler(svc service.PremiumService) *PremiumHandler {
	return &PremiumHandler{premiumSvc: svc}
}

func (h *PremiumHandler) GetLatest(c *gin.Context) {
	resp, err := h.premiumSvc.GetLatest(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *PremiumHandler) GetHistory(c *gin.Context) {
	var req dto.PremiumHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.premiumSvc.GetHistory(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func handleError(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
