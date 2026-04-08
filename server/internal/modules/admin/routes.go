package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/middleware"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

func RegisterRoutes(r *gin.RouterGroup, m *Module) {
	admin := r.Group("/admin")

	// Auth routes
	auth := admin.Group("/auth")
	{
		auth.POST("/login", m.AuthHandler.Login)
		auth.POST("/change-password", middleware.RequireAdmin(m.JWTSecret), m.AuthHandler.ChangePassword)
		auth.GET("/me", middleware.RequireAdmin(m.JWTSecret), func(c *gin.Context) {
			adminID := c.GetString("admin_id")
			adminUser, err := m.AdminRepo.FindByID(c.Request.Context(), adminID)
			if err != nil {
				response.Fail(c, http.StatusUnauthorized, "ADMIN_NOT_FOUND", "admin user not found")
				return
			}
			response.OK(c, adminUser)
		})
	}

	// Protected routes (all require admin auth)
	protected := admin.Group("", middleware.RequireAdmin(m.JWTSecret))

	// Dashboard
	protected.GET("/dashboard/stats",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
		m.DashboardHandler.Stats,
	)

	// Investments
	protected.GET("/investments",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
		m.UserHandler.ListInvestments,
	)

	// Trades
	protected.GET("/trades",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
		m.TradeHandler.List,
	)
	protected.GET("/trades/template",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
		m.TradeHandler.DownloadTemplate,
	)
	protected.POST("/trades/import",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
		m.TradeHandler.ImportTrades,
	)
	protected.GET("/trades/stats",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
		m.TradeHandler.Stats,
	)
	protected.DELETE("/trades/:id",
		middleware.RequireRole(model.RoleSuperAdmin),
		m.TradeHandler.Delete,
	)

	// Transactions
	protected.GET("/transactions",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
		m.TransactionHandler.List,
	)
	protected.GET("/transactions/stats",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
		m.TransactionHandler.Stats,
	)

	// User management
	users := protected.Group("/users")
	{
		users.GET("",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
			m.UserHandler.List,
		)
		users.GET("/:id",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
			m.UserHandler.Detail,
		)
		users.PUT("/:id",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.UserHandler.Update,
		)
		users.POST("/:id/disable",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.UserHandler.Disable,
		)
		users.POST("/:id/enable",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.UserHandler.Enable,
		)
		users.POST("/:id/reset-password",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.UserHandler.ResetPassword,
		)
		users.POST("/:id/reset-2fa",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.UserHandler.Reset2FA,
		)
		users.POST("/:id/adjust-balance",
			middleware.RequireRole(model.RoleSuperAdmin),
			m.UserHandler.AdjustBalance,
		)
	}

	// Admin user management
	adminUsers := protected.Group("/admin-users")
	{
		adminUsers.GET("",
			middleware.RequireRole(model.RoleSuperAdmin),
			m.AdminUserHandler.List,
		)
		adminUsers.POST("",
			middleware.RequireRole(model.RoleSuperAdmin),
			m.AdminUserHandler.Create,
		)
		adminUsers.PUT("/:id",
			middleware.RequireRole(model.RoleSuperAdmin),
			m.AdminUserHandler.Update,
		)
		adminUsers.DELETE("/:id",
			middleware.RequireRole(model.RoleSuperAdmin),
			m.AdminUserHandler.Delete,
		)
	}
}
