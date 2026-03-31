package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDGenerated(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get(RequestIDKey)
		c.String(200, id.(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	rid := w.Header().Get(RequestIDKey)
	if rid == "" {
		t.Error("X-Request-ID header should be set")
	}
	if len(rid) < 30 {
		t.Error("generated ID should be a UUID")
	}
}

func TestRequestIDPreserved(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDKey, "my-custom-id")
	r.ServeHTTP(w, req)

	if w.Header().Get(RequestIDKey) != "my-custom-id" {
		t.Error("should preserve incoming X-Request-ID")
	}
}
