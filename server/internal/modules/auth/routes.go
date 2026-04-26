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
		auth.POST("/verify-2fa", h.Verify2FA) // 公开路由：2FA 验证时尚无 token
	}

	protected := auth.Group("", middleware.Auth(jwtMgr))
	{
		protected.POST("/logout", h.Logout)
		protected.GET("/profile", h.GetProfile)
	}
}
