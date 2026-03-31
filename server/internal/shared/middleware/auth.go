package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
	"github.com/sovereign-fund/sovereign/pkg/jwt"
)

func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Fail(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "invalid authorization format")
			return
		}

		claims, err := jwtManager.ValidateAccess(parts[1])
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid or expired token")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
