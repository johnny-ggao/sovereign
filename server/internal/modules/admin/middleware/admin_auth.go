package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type AdminClaims struct {
	AdminID string `json:"admin_id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAdminToken(secret string, expiry time.Duration, adminID, email, role string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := AdminClaims{
		AdminID: adminID,
		Email:   email,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "sovereign-admin",
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return token, expiresAt.Unix(), nil
}

func RequireAdmin(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Fail(c, http.StatusUnauthorized, "ADMIN_UNAUTHORIZED", "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, http.StatusUnauthorized, "ADMIN_UNAUTHORIZED", "invalid authorization format")
			return
		}

		claims, err := validateAdminToken(parts[1], secret)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "ADMIN_INVALID_TOKEN", "invalid or expired admin token")
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("admin_email", claims.Email)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		role := c.GetString("admin_role")
		if !roleSet[role] {
			response.Fail(c, http.StatusForbidden, "ADMIN_FORBIDDEN", "insufficient permissions")
			return
		}
		c.Next()
	}
}

func validateAdminToken(tokenStr, secret string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
