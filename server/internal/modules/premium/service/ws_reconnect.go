package service

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Backoff 指数退避 + 抖动
type Backoff struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	attempt    int
}

func NewBackoff() *Backoff {
	return &Backoff{
		Initial:    time.Second,
		Max:        30 * time.Second,
		Multiplier: 2,
	}
}

func (b *Backoff) Next() time.Duration {
	if b.attempt == 0 {
		b.attempt++
		return b.Initial
	}

	d := b.Initial
	for i := 0; i < b.attempt; i++ {
		d = time.Duration(float64(d) * b.Multiplier)
		if d > b.Max {
			d = b.Max
			break
		}
	}
	b.attempt++

	// 加 ±25% 抖动
	jitter := float64(d) * (0.75 + rand.Float64()*0.5)
	return time.Duration(jitter)
}

func (b *Backoff) Reset() {
	b.attempt = 0
}

// WSConn 封装 websocket.Conn，提供并发安全的写操作
type WSConn struct {
	*websocket.Conn
	mu sync.Mutex
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{Conn: conn}
}

func (c *WSConn) WriteJSONSafe(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

func (c *WSConn) WritePingSafe(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.Conn.WriteMessage(websocket.PingMessage, data)
}

// PingLoop 定期发 WebSocket ping 并测量 RTT，结果写入 cache
func PingLoop(ctx context.Context, conn *WSConn, cache *OrderBookCache, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	conn.SetPongHandler(func(appData string) error {
		sent, err := time.Parse(time.RFC3339Nano, appData)
		if err == nil {
			cache.SetRTT(time.Since(sent))
		}
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WritePingSafe([]byte(time.Now().Format(time.RFC3339Nano))); err != nil {
				return
			}
		}
	}
}

// SleepWithContext 可被 context 取消的 sleep
func SleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
