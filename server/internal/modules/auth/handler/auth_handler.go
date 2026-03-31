package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type AuthHandler struct {
	authSvc service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) SendRegisterOTP(c *gin.Context) {
	var req dto.SendRegisterOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.authSvc.SendRegisterOTP(c.Request.Context(), req); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, dto.MessageResponse{Message: "OTP sent to email"})
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.authSvc.GoogleLogin(c.Request.Context(), req.IDToken, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.authSvc.Register(c.Request.Context(), req, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.authSvc.Login(c.Request.Context(), req, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		handleError(c, err)
		return
	}

	if resp.Requires2FA {
		c.JSON(http.StatusAccepted, response.Response{
			Success: true,
			Data:    resp,
		})
		return
	}

	response.OK(c, resp)
}

func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req dto.Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	userID, _ := c.Get("user_id")

	resp, err := h.authSvc.Verify2FA(c.Request.Context(), userID.(string), req.Code, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, resp)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.authSvc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, dto.MessageResponse{Message: "if the email exists, a reset code has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.authSvc.ResetPassword(c.Request.Context(), req); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, dto.MessageResponse{Message: "password has been reset successfully"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, dto.MessageResponse{Message: "logged out successfully"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	profile, err := h.authSvc.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, profile)
}

func handleError(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
