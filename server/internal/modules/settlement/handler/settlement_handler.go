package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type SettlementHandler struct {
	settleSvc service.SettlementService
}

func NewSettlementHandler(svc service.SettlementService) *SettlementHandler {
	return &SettlementHandler{settleSvc: svc}
}

func (h *SettlementHandler) GetAll(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.settleSvc.GetAll(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

func (h *SettlementHandler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	resp, err := h.settleSvc.GetByID(c.Request.Context(), userID, id)
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
