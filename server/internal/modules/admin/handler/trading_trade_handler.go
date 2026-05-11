package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type TradingTradeHandler struct {
	svc      service.TradingTradeService
	auditSvc service.AuditService
}

func NewTradingTradeHandler(svc service.TradingTradeService, auditSvc service.AuditService) *TradingTradeHandler {
	return &TradingTradeHandler{svc: svc, auditSvc: auditSvc}
}

func (h *TradingTradeHandler) List(c *gin.Context) {
	var query dto.TradeListQuery
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
		response.Fail(c, http.StatusInternalServerError, "TRADING_LIST_FAILED", err.Error())
		return
	}

	response.Paginated(c, items, response.Meta{Total: total, Page: query.Page, PerPage: query.Limit})
}

func (h *TradingTradeHandler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_STATS_FAILED", err.Error())
		return
	}

	response.OK(c, stats)
}

func (h *TradingTradeHandler) DownloadTemplate(c *gin.Context) {
	file, err := h.svc.DownloadTemplate(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_TEMPLATE_FAILED", err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="trading-trade-import-template.xlsx"`)
	if err := file.Write(c.Writer); err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_TEMPLATE_WRITE_FAILED", err.Error())
	}
}

func (h *TradingTradeHandler) Import(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_FILE", "missing file")
		return
	}

	file, err := header.Open()
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_FILE", "unable to open uploaded file")
		return
	}
	defer file.Close()

	imported, rowErrors, err := h.svc.ImportFromExcel(c.Request.Context(), file)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_IMPORT_FAILED", err.Error())
		return
	}

	if err := h.auditSvc.Log(
		c.Request.Context(),
		c.GetString("admin_id"),
		c.GetString("admin_email"),
		"import_trading_trades",
		"trading_trade",
		"",
		fmt.Sprintf("imported=%d errors=%d", imported, len(rowErrors)),
		c.ClientIP(),
	); err != nil {
		log.Printf("failed to write audit log: %v", err)
	}

	response.OK(c, gin.H{"imported": imported, "errors": rowErrors})
}

func (h *TradingTradeHandler) Delete(c *gin.Context) {
	tradeID := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), tradeID); err != nil {
		response.Fail(c, http.StatusBadRequest, "DELETE_TRADING_FAILED", err.Error())
		return
	}

	if err := h.auditSvc.Log(
		c.Request.Context(),
		c.GetString("admin_id"),
		c.GetString("admin_email"),
		"delete_trading_trade",
		"trading_trade",
		tradeID,
		"",
		c.ClientIP(),
	); err != nil {
		log.Printf("failed to write audit log: %v", err)
	}
	response.OK(c, gin.H{"message": "交易记录已删除"})
}

