package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func RateLimit(rate int, window time.Duration) gin.HandlerFunc {
	return RateLimitWithContext(context.Background(), rate, window)
}

func RateLimitWithContext(ctx context.Context, rate int, window time.Duration) gin.HandlerFunc {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	go rl.cleanup(ctx)

	return func(c *gin.Context) {
		key := c.ClientIP()

		if userID, exists := c.Get("user_id"); exists {
			key = userID.(string)
		}

		if !rl.allow(key) {
			response.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
			return
		}

		c.Next()
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists || time.Since(v.lastSeen) > rl.window {
		rl.visitors[key] = &visitor{count: 1, lastSeen: time.Now()}
		return true
	}

	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = time.Now()
	return true
}

func (rl *rateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.window*2 {
					delete(rl.visitors, key)
				}
			}
			rl.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
