package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/investment/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type InvestmentHandler struct {
	invSvc service.InvestmentService
}

func NewInvestmentHandler(svc service.InvestmentService) *InvestmentHandler {
	return &InvestmentHandler{invSvc: svc}
}

func (h *InvestmentHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.CreateInvestmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.invSvc.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, resp)
}

func (h *InvestmentHandler) GetAll(c *gin.Context) {
	userID := c.GetString("user_id")

	resp, err := h.invSvc.GetAll(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

func (h *InvestmentHandler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	resp, err := h.invSvc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

func (h *InvestmentHandler) Redeem(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.invSvc.Redeem(c.Request.Context(), userID, req)
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
