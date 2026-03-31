package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

const upbitWSURL = "wss://api.upbit.com/websocket/v1"

type UpbitWSClient struct {
	cache  *OrderBookCache
	logger *slog.Logger
	ready  atomic.Bool
}

func NewUpbitWSClient(logger *slog.Logger) *UpbitWSClient {
	return &UpbitWSClient{
		cache:  NewOrderBookCache(),
		logger: logger,
	}
}

func (c *UpbitWSClient) Name() string    { return "upbit" }
func (c *UpbitWSClient) Latency() time.Duration { return c.cache.Latency() }

func (c *UpbitWSClient) Ready() bool { return c.ready.Load() }

func (c *UpbitWSClient) GetOrderBook(_ context.Context, pair string) (*OrderBookTop, error) {
	return c.cache.Get(pair)
}

func (c *UpbitWSClient) Start(ctx context.Context) error {
	backoff := NewBackoff()
	markets := []string{"KRW-BTC", "KRW-ETH", "KRW-SOL", "KRW-XRP", "KRW-USDT"}

	for {
		if err := c.connect(ctx, markets); err != nil {
			c.logger.Warn("upbit ws disconnected", slog.String("error", err.Error()))
			c.ready.Store(false)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wait := backoff.Next()
		c.logger.Info("upbit ws reconnecting", slog.Duration("delay", wait))
		if err := SleepWithContext(ctx, wait); err != nil {
			return err
		}
	}
}

func (c *UpbitWSClient) connect(ctx context.Context, markets []string) error {
	raw, _, err := websocket.DefaultDialer.DialContext(ctx, upbitWSURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer raw.Close()
	conn := NewWSConn(raw)

	// 订阅
	sub := []interface{}{
		map[string]string{"ticket": "sovereign"},
		map[string]interface{}{"type": "orderbook", "codes": markets},
	}
	if err := conn.WriteJSONSafe(sub); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	c.logger.Info("upbit ws connected", slog.Int("markets", len(markets)))

	go func() { <-ctx.Done(); raw.Close() }()
	go PingLoop(ctx, conn, c.cache, 10*time.Second)

	for {
		_, msg, err := raw.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		c.handleMessage(msg)
	}
}

type upbitWSOrderBook struct {
	Type           string `json:"type"`
	Code           string `json:"code"` // "KRW-BTC"
	OrderbookUnits []struct {
		AskPrice decimal.Decimal `json:"ask_price"`
		BidPrice decimal.Decimal `json:"bid_price"`
	} `json:"orderbook_units"`
}

func (c *UpbitWSClient) handleMessage(data []byte) {
	var ob upbitWSOrderBook
	if err := json.Unmarshal(data, &ob); err != nil {
		return
	}

	if ob.Type != "orderbook" || len(ob.OrderbookUnits) == 0 {
		return
	}

	pair := upbitCodeToPair(ob.Code)
	top := ob.OrderbookUnits[0]

	c.cache.Update(pair, &OrderBookTop{
		Bid: top.BidPrice,
		Ask: top.AskPrice,
	})

	c.ready.Store(true)
}

// upbitCodeToPair "KRW-BTC" -> "BTC/KRW"
func upbitCodeToPair(code string) string {
	parts := strings.Split(code, "-")
	if len(parts) != 2 {
		return code
	}
	return parts[1] + "/" + parts[0]
}
