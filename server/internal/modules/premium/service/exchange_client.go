package service

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// OrderBookTop 订单簿买一卖一
type OrderBookTop struct {
	Bid decimal.Decimal // 买一价（最高买价）
	Ask decimal.Decimal // 卖一价（最低卖价）
}

// ExchangeClient 交易所行情客户端接口
type ExchangeClient interface {
	GetOrderBook(ctx context.Context, pair string) (*OrderBookTop, error)
	Name() string
	Latency() time.Duration
}

// WSExchangeClient 支持 WebSocket 连接的交易所客户端
type WSExchangeClient interface {
	ExchangeClient
	Start(ctx context.Context) error
	Ready() bool
}

// MockExchangeClient 开发阶段模拟价格，带随机波动
type MockExchangeClient struct {
	name      string
	basePrice decimal.Decimal
	premium   decimal.Decimal
	mu        sync.Mutex
	lastPrice map[string]decimal.Decimal
}

func NewMockExchangeClient(name string, basePrice float64, premiumOffset float64) ExchangeClient {
	return &MockExchangeClient{
		name:      name,
		basePrice: decimal.NewFromFloat(basePrice),
		premium:   decimal.NewFromFloat(premiumOffset),
		lastPrice: make(map[string]decimal.Decimal),
	}
}

func (m *MockExchangeClient) GetOrderBook(_ context.Context, pair string) (*OrderBookTop, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	last, exists := m.lastPrice[pair]
	if !exists {
		last = m.basePrice.Add(m.premium)
	}

	pctChange := (rand.Float64() - 0.5) * 0.006
	delta := last.Mul(decimal.NewFromFloat(pctChange))
	mid := last.Add(delta).Round(0)

	base := m.basePrice.Add(m.premium)
	lowerBound := base.Mul(decimal.NewFromFloat(0.95))
	upperBound := base.Mul(decimal.NewFromFloat(1.05))
	if mid.LessThan(lowerBound) {
		mid = lowerBound
	}
	if mid.GreaterThan(upperBound) {
		mid = upperBound
	}

	m.lastPrice[pair] = mid

	// 模拟 spread: ±0.02%
	spreadHalf := mid.Mul(decimal.NewFromFloat(0.0002))
	return &OrderBookTop{
		Bid: mid.Sub(spreadHalf).Round(0),
		Ask: mid.Add(spreadHalf).Round(0),
	}, nil
}

func (m *MockExchangeClient) Name() string {
	return m.name
}

func (m *MockExchangeClient) Latency() time.Duration {
	return time.Duration(10+rand.IntN(20)) * time.Millisecond
}
