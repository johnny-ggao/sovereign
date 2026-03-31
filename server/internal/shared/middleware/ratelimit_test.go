package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimitAllow(t *testing.T) {
	r := gin.New()
	r.Use(RateLimit(5, time.Second))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	for range 5 {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("request should succeed, got %d", w.Code)
		}
	}
}

func TestRateLimitBlock(t *testing.T) {
	r := gin.New()
	r.Use(RateLimit(2, time.Second))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	for i := range 4 {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "5.6.7.8:1234"
		r.ServeHTTP(w, req)

		if i >= 2 && w.Code != http.StatusTooManyRequests {
			t.Errorf("request %d should be blocked, got %d", i, w.Code)
		}
	}
}
