package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "LOGIN_FAILED", err.Error())
		return
	}

	response.OK(c, resp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	adminID := c.GetString("admin_id")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), adminID, req); err != nil {
		response.Fail(c, http.StatusBadRequest, "CHANGE_PASSWORD_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "password changed"})
}
