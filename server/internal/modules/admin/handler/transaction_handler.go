package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type TransactionHandler struct {
	svc service.TransactionService
}

func NewTransactionHandler(svc service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) List(c *gin.Context) {
	var query dto.TransactionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	items, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TX_LIST_FAILED", err.Error())
		return
	}
	response.Paginated(c, items, response.Meta{Total: total, Page: query.Page, PerPage: query.Limit})
}

func (h *TransactionHandler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TX_STATS_FAILED", err.Error())
		return
	}
	response.OK(c, stats)
}
