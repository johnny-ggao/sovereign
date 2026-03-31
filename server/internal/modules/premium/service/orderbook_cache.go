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
	latency time.Duration // 最近一次消息的处理延迟
}

func NewOrderBookCache() *OrderBookCache {
	return &OrderBookCache{
		books: make(map[string]*OrderBookTop),
		ts:    make(map[string]time.Time),
	}
}

func (c *OrderBookCache) Update(pair string, ob *OrderBookTop) {
	c.mu.Lock()
	now := time.Now()
	if prev, ok := c.ts[pair]; ok {
		c.latency = now.Sub(prev)
	}
	c.books[pair] = ob
	c.ts[pair] = now
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

// Latency 返回最近一次消息间隔（反映 WS 推送频率/延迟）
func (c *OrderBookCache) Latency() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latency
}
