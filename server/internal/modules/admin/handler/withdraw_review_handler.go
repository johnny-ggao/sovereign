package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type WithdrawReviewHandler struct {
	svc      service.WithdrawReviewService
	auditSvc service.AuditService
}

func NewWithdrawReviewHandler(svc service.WithdrawReviewService, auditSvc service.AuditService) *WithdrawReviewHandler {
	return &WithdrawReviewHandler{svc: svc, auditSvc: auditSvc}
}

func (h *WithdrawReviewHandler) List(c *gin.Context) {
	var q dto.WithdrawReviewListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "WITHDRAW_LIST_FAILED", err.Error())
		return
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 {
		limit = 20
	}
	response.Paginated(c, items, response.Meta{Total: total, Page: page, PerPage: limit})
}

func (h *WithdrawReviewHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")
	err := h.svc.Approve(c.Request.Context(), id, adminID)
	h.audit(c, "approve_withdrawal", id, "")
	h.respond(c, err)
}

func (h *WithdrawReviewHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")
	var req dto.RejectWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	err := h.svc.Reject(c.Request.Context(), id, adminID, req.Reason)
	h.audit(c, "reject_withdrawal", id, "reason="+req.Reason)
	h.respond(c, err)
}

func (h *WithdrawReviewHandler) Retry(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")
	err := h.svc.Retry(c.Request.Context(), id, adminID)
	h.audit(c, "retry_withdrawal", id, "")
	h.respond(c, err)
}

func (h *WithdrawReviewHandler) audit(c *gin.Context, action, id, detail string) {
	if err := h.auditSvc.Log(
		c.Request.Context(),
		c.GetString("admin_id"),
		c.GetString("admin_email"),
		action, "transaction", id, detail, c.ClientIP(),
	); err != nil {
		slog.Error("audit log failed", slog.String("action", action), slog.String("error", err.Error()))
	}
}

func (h *WithdrawReviewHandler) respond(c *gin.Context, err error) {
	if err == nil {
		response.OK(c, gin.H{"message": "ok"})
		return
	}
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		response.Fail(c, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
