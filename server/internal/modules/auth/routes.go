package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/handler"
	"github.com/sovereign-fund/sovereign/internal/shared/middleware"
	jwtpkg "github.com/sovereign-fund/sovereign/pkg/jwt"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, jwtMgr *jwtpkg.Manager) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register/send-otp", h.SendRegisterOTP)
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/google", h.GoogleLogin)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
	}

	protected := auth.Group("", middleware.Auth(jwtMgr))
	{
		protected.POST("/verify-2fa", h.Verify2FA)
		protected.POST("/logout", h.Logout)
		protected.GET("/profile", h.GetProfile)
	}
}
