package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
	"github.com/sovereign-fund/sovereign/pkg/pagination"
)

type TradeHandler struct {
	tradeSvc service.TradeService
}

func NewTradeHandler(svc service.TradeService) *TradeHandler {
	return &TradeHandler{tradeSvc: svc}
}

func (h *TradeHandler) GetTrades(c *gin.Context) {
	userID := c.GetString("user_id")
	p := pagination.Parse(c)

	var filters dto.TradeFilterRequest
	if err := c.ShouldBindQuery(&filters); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, total, err := h.tradeSvc.GetTrades(c.Request.Context(), userID, filters, p.Page, p.PerPage)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, resp, response.Meta{
		Total:   total,
		Page:    p.Page,
		PerPage: p.PerPage,
	})
}

func (h *TradeHandler) CreateTrade(c *gin.Context) {
	var req dto.CreateTradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.tradeSvc.CreateTrade(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, resp)
}

func (h *TradeHandler) BatchCreateTrades(c *gin.Context) {
	var req dto.BatchCreateTradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	count, err := h.tradeSvc.BatchCreateTrades(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, gin.H{"created": count})
}

func (h *TradeHandler) ExportCSV(c *gin.Context) {
	userID := c.GetString("user_id")

	var filters dto.TradeFilterRequest
	_ = c.ShouldBindQuery(&filters)

	data, err := h.tradeSvc.ExportCSV(c.Request.Context(), userID, filters)
	if err != nil {
		handleError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=trades.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

func handleError(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
