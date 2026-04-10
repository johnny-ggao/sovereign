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

type AdminUserHandler struct {
	svc      service.AdminUserService
	auditSvc service.AuditService
}

func NewAdminUserHandler(svc service.AdminUserService, auditSvc service.AuditService) *AdminUserHandler {
	return &AdminUserHandler{svc: svc, auditSvc: auditSvc}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	admins, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "ADMIN_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, admins)
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	var req dto.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	admin, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "CREATE_ADMIN_FAILED", err.Error())
		return
	}

	detail := fmt.Sprintf("email=%s, role=%s", admin.Email, admin.Role)
	if err := h.auditSvc.Log(
		c.Request.Context(),
		c.GetString("admin_id"),
		c.GetString("admin_email"),
		"create_admin",
		"admin",
		admin.ID,
		detail,
		c.ClientIP(),
	); err != nil {
		log.Printf("failed to write audit log: %v", err)
	}

	response.Created(c, admin)
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	admin, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "UPDATE_ADMIN_FAILED", err.Error())
		return
	}

	response.OK(c, admin)
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	currentAdminID := c.GetString("admin_id")

	if err := h.svc.Delete(c.Request.Context(), id, currentAdminID); err != nil {
		response.Fail(c, http.StatusBadRequest, "DELETE_ADMIN_FAILED", err.Error())
		return
	}

	if err := h.auditSvc.Log(
		c.Request.Context(),
		c.GetString("admin_id"),
		c.GetString("admin_email"),
		"delete_admin",
		"admin",
		id,
		"",
		c.ClientIP(),
	); err != nil {
		log.Printf("failed to write audit log: %v", err)
	}

	response.NoContent(c)
}
