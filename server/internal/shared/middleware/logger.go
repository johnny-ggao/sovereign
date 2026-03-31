package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get(RequestIDKey)

		attrs := []slog.Attr{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		}

		if rid, ok := requestID.(string); ok {
			attrs = append(attrs, slog.String("request_id", rid))
		}

		if userID, exists := c.Get("user_id"); exists {
			attrs = append(attrs, slog.String("user_id", userID.(string)))
		}

		msg := "request"
		if status >= 500 {
			logger.LogAttrs(c.Request.Context(), slog.LevelError, msg, attrs...)
		} else if status >= 400 {
			logger.LogAttrs(c.Request.Context(), slog.LevelWarn, msg, attrs...)
		} else {
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, msg, attrs...)
		}
	}
}
