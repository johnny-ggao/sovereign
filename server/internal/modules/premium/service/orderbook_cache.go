package service

import (
	"fmt"
	"sync"
	"time"
)

const staleThreshold = 30 * time.Second

// OrderBookCache 线程安全的订单簿缓存
type OrderBookCache struct {
	mu      sync.RWMutex
	books   map[string]*OrderBookTop
	ts      map[string]time.Time
	rtt     time.Duration
}

func NewOrderBookCache() *OrderBookCache {
	return &OrderBookCache{
		books: make(map[string]*OrderBookTop),
		ts:    make(map[string]time.Time),
	}
}

func (c *OrderBookCache) Update(pair string, ob *OrderBookTop) {
	c.mu.Lock()
	c.books[pair] = ob
	c.ts[pair] = time.Now()
	c.mu.Unlock()
}

func (c *OrderBookCache) Get(pair string) (*OrderBookTop, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ob, ok := c.books[pair]
	if !ok {
		return nil, fmt.Errorf("no orderbook for %s", pair)
	}

	if time.Since(c.ts[pair]) > staleThreshold {
		return nil, fmt.Errorf("stale orderbook for %s (age %s)", pair, time.Since(c.ts[pair]))
	}

	return ob, nil
}

// SetRTT 设置 ping/pong 测量的 RTT
func (c *OrderBookCache) SetRTT(d time.Duration) {
	c.mu.Lock()
	c.rtt = d
	c.mu.Unlock()
}

// Latency 返回 ping/pong RTT
func (c *OrderBookCache) Latency() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rtt
}
