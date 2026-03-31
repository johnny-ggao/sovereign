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

const bybitWSURL = "wss://stream.bybit.com/v5/public/spot"

type BybitWSClient struct {
	cache  *OrderBookCache
	logger *slog.Logger
	ready  atomic.Bool
}

func NewBybitWSClient(logger *slog.Logger) *BybitWSClient {
	return &BybitWSClient{
		cache:  NewOrderBookCache(),
		logger: logger,
	}
}

func (c *BybitWSClient) Name() string          { return "bybit" }
func (c *BybitWSClient) Latency() time.Duration { return c.cache.Latency() }

func (c *BybitWSClient) Ready() bool { return c.ready.Load() }

func (c *BybitWSClient) GetOrderBook(_ context.Context, pair string) (*OrderBookTop, error) {
	return c.cache.Get(pair)
}

func (c *BybitWSClient) Start(ctx context.Context) error {
	backoff := NewBackoff()

	for {
		if err := c.connect(ctx); err != nil {
			c.logger.Warn("bybit ws disconnected", slog.String("error", err.Error()))
			c.ready.Store(false)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wait := backoff.Next()
		c.logger.Info("bybit ws reconnecting", slog.Duration("delay", wait))
		if err := SleepWithContext(ctx, wait); err != nil {
			return err
		}
	}
}

func (c *BybitWSClient) connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, bybitWSURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	// 订阅 level 1 orderbook（只有 best bid/ask）
	sub := map[string]interface{}{
		"op": "subscribe",
		"args": []string{
			"orderbook.1.BTCUSDT",
			"orderbook.1.ETHUSDT",
			"orderbook.1.SOLUSDT",
			"orderbook.1.XRPUSDT",
		},
	}
	if err := conn.WriteJSON(sub); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	c.logger.Info("bybit ws connected")

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// Bybit 需要每 20 秒发心跳
	go c.heartbeat(ctx, conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		c.handleMessage(msg)
	}
}

func (c *BybitWSClient) heartbeat(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteJSON(map[string]string{"op": "ping"}); err != nil {
				return
			}
		}
	}
}

type bybitWSMsg struct {
	Topic string `json:"topic"` // "orderbook.1.BTCUSDT"
	Type  string `json:"type"`  // "snapshot" or "delta"
	Data  struct {
		S string     `json:"s"` // "BTCUSDT"
		B [][]string `json:"b"` // bids [["price","size"]]
		A [][]string `json:"a"` // asks [["price","size"]]
	} `json:"data"`
}

func (c *BybitWSClient) handleMessage(data []byte) {
	var msg bybitWSMsg
	if err := json.Unmarshal(data, &msg); err != nil || msg.Topic == "" {
		return
	}

	if !strings.HasPrefix(msg.Topic, "orderbook.") {
		return
	}

	if len(msg.Data.B) == 0 || len(msg.Data.A) == 0 {
		return
	}

	bid, _ := decimal.NewFromString(msg.Data.B[0][0])
	ask, _ := decimal.NewFromString(msg.Data.A[0][0])

	pair := bybitSymbolToPair(msg.Data.S)

	c.cache.Update(pair, &OrderBookTop{Bid: bid, Ask: ask})
	c.ready.Store(true)
}

// bybitSymbolToPair "BTCUSDT" -> "BTC/KRW"
func bybitSymbolToPair(symbol string) string {
	base := strings.TrimSuffix(symbol, "USDT")
	return base + "/KRW"
}
