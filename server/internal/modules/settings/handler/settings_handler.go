package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type SettingsHandler struct {
	svc service.SettingsService
}

func NewSettingsHandler(svc service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	resp, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *SettingsHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	resp, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *SettingsHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "password changed"})
}

func (h *SettingsHandler) Setup2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	resp, err := h.svc.Setup2FA(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *SettingsHandler) Verify2FASetup(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.Verify2FASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if err := h.svc.Verify2FASetup(c.Request.Context(), userID, req); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "2FA enabled"})
}

func (h *SettingsHandler) Disable2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if err := h.svc.Disable2FA(c.Request.Context(), userID, req); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "2FA disabled"})
}

func (h *SettingsHandler) GetSecurityOverview(c *gin.Context) {
	userID := c.GetString("user_id")
	resp, err := h.svc.GetSecurityOverview(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *SettingsHandler) RevokeDevice(c *gin.Context) {
	userID := c.GetString("user_id")
	deviceID := c.Param("id")
	if err := h.svc.RevokeDevice(c.Request.Context(), userID, deviceID); err != nil {
		handleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *SettingsHandler) GetNotificationPref(c *gin.Context) {
	userID := c.GetString("user_id")
	resp, err := h.svc.GetNotificationPref(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *SettingsHandler) UpdateNotificationPref(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	resp, err := h.svc.UpdateNotificationPref(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *SettingsHandler) UpdateLanguage(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.UpdateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if err := h.svc.UpdateLanguage(c.Request.Context(), userID, req); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "language updated"})
}

func handleError(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
