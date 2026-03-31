package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					slog.Any("error", r),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
				)
				response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			}
		}()
		c.Next()
	}
}
