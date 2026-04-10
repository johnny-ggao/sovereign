package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type AuditHandler struct {
	svc service.AuditService
}

func NewAuditHandler(svc service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(c *gin.Context) {
	var req dto.AuditLogListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}

	query := service.AuditListQuery{
		Page:     req.Page,
		Limit:    req.Limit,
		Action:   req.Action,
		AdminID:  req.AdminID,
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	items, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "AUDIT_LOG_LIST_FAILED", err.Error())
		return
	}

	result := make([]dto.AuditLogResponse, len(items))
	for i, item := range items {
		result[i] = dto.AuditLogResponse{
			ID:         item.ID,
			AdminID:    item.AdminID,
			AdminEmail: item.AdminEmail,
			Action:     item.Action,
			TargetType: item.TargetType,
			TargetID:   item.TargetID,
			Detail:     item.Detail,
			IPAddress:  item.IPAddress,
			CreatedAt:  item.CreatedAt,
		}
	}

	response.Paginated(c, result, response.Meta{
		Total:   total,
		Page:    query.Page,
		PerPage: query.Limit,
	})
}
