package settings

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/handler"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handler.SettingsHandler) {
	s := rg.Group("/settings")
	{
		// Profile
		s.GET("/profile", h.GetProfile)
		s.PUT("/profile", h.UpdateProfile)

		// Security
		s.PUT("/password", h.ChangePassword)
		s.GET("/security", h.GetSecurityOverview)
		s.POST("/2fa/setup", h.Setup2FA)
		s.POST("/2fa/verify", h.Verify2FASetup)
		s.POST("/2fa/disable", h.Disable2FA)
		s.DELETE("/devices/:id", h.RevokeDevice)

		// Notifications
		s.GET("/notifications", h.GetNotificationPref)
		s.PUT("/notifications", h.UpdateNotificationPref)

		// Language
		s.PUT("/language", h.UpdateLanguage)

	}
}
