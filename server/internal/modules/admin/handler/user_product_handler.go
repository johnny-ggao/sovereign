package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type UserProductHandler struct {
	svc      service.UserProductService
	auditSvc service.AuditService
}

func NewUserProductHandler(svc service.UserProductService, auditSvc service.AuditService) *UserProductHandler {
	return &UserProductHandler{svc: svc, auditSvc: auditSvc}
}

func (h *UserProductHandler) Change(c *gin.Context) {
	id := c.Param("id")
	var req dto.ChangeUserProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	adminID := c.GetString("admin_id")
	adminEmail := c.GetString("admin_email")
	err := h.svc.Change(c.Request.Context(), id, req.Type, req.Reason, adminID, adminEmail)
	h.audit(c, "change_investment_type", id, "to="+req.Type+" reason="+req.Reason)
	h.respond(c, err)
}

func (h *UserProductHandler) BulkChange(c *gin.Context) {
	var req dto.BulkChangeUserProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	adminID := c.GetString("admin_id")
	adminEmail := c.GetString("admin_email")
	res, err := h.svc.BulkChange(c.Request.Context(), req.UserIDs, req.Type, req.Reason, adminID, adminEmail)
	h.audit(c, "bulk_change_investment_type", "", fmt.Sprintf("to=%s count=%d reason=%s", req.Type, len(req.UserIDs), req.Reason))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	response.OK(c, res)
}

func (h *UserProductHandler) History(c *gin.Context) {
	id := c.Param("id")
	rows, err := h.svc.History(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "HISTORY_FAILED", err.Error())
		return
	}
	response.OK(c, rows)
}

func (h *UserProductHandler) audit(c *gin.Context, action, id, detail string) {
	if err := h.auditSvc.Log(c.Request.Context(), c.GetString("admin_id"), c.GetString("admin_email"), action, "user", id, detail, c.ClientIP()); err != nil {
		slog.Error("audit log failed", slog.String("action", action), slog.String("error", err.Error()))
	}
}

func (h *UserProductHandler) respond(c *gin.Context, err error) {
	if err == nil {
		response.OK(c, gin.H{"message": "ok"})
		return
	}
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		response.Fail(c, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
}
