package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

// InternalAuth 内部 API 认证中间件：API Key + IP 白名单
func InternalAuth(apiKey string, allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. IP 白名单检查
		if len(allowedIPs) > 0 {
			clientIP := c.ClientIP()
			allowed := false
			for _, ip := range allowedIPs {
				if ip == "*" {
					allowed = true
					break
				}
				// 支持 CIDR 格式（如 172.31.0.0/16）
				if strings.Contains(ip, "/") {
					_, ipNet, err := net.ParseCIDR(ip)
					if err == nil && ipNet.Contains(net.ParseIP(clientIP)) {
						allowed = true
						break
					}
				} else if ip == clientIP {
					allowed = true
					break
				}
			}
			if !allowed {
				response.Fail(c, http.StatusForbidden, "IP_FORBIDDEN", "access denied from "+clientIP)
				return
			}
		}

		// 2. API Key 检查
		if apiKey != "" {
			key := c.GetHeader("X-Internal-Key")
			if key == "" {
				key = c.Query("key")
			}
			if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
				response.Fail(c, http.StatusUnauthorized, "INVALID_API_KEY", "invalid or missing internal API key")
				return
			}
		}

		c.Next()
	}
}
