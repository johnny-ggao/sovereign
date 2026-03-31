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

const bithumbWSURL = "wss://pubwss.bithumb.com/pub/ws"

type BithumbWSClient struct {
	cache  *OrderBookCache
	logger *slog.Logger
	ready  atomic.Bool
}

func NewBithumbWSClient(logger *slog.Logger) *BithumbWSClient {
	return &BithumbWSClient{
		cache:  NewOrderBookCache(),
		logger: logger,
	}
}

func (c *BithumbWSClient) Name() string          { return "bithumb" }
func (c *BithumbWSClient) Latency() time.Duration { return c.cache.Latency() }

func (c *BithumbWSClient) Ready() bool { return c.ready.Load() }

func (c *BithumbWSClient) GetOrderBook(_ context.Context, pair string) (*OrderBookTop, error) {
	return c.cache.Get(pair)
}

func (c *BithumbWSClient) Start(ctx context.Context) error {
	backoff := NewBackoff()

	for {
		if err := c.connect(ctx); err != nil {
			c.logger.Warn("bithumb ws disconnected", slog.String("error", err.Error()))
			c.ready.Store(false)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wait := backoff.Next()
		c.logger.Info("bithumb ws reconnecting", slog.Duration("delay", wait))
		if err := SleepWithContext(ctx, wait); err != nil {
			return err
		}
	}
}

func (c *BithumbWSClient) connect(ctx context.Context) error {
	raw, _, err := websocket.DefaultDialer.DialContext(ctx, bithumbWSURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer raw.Close()
	conn := NewWSConn(raw)

	sub := map[string]interface{}{
		"type":      "orderbookdepth",
		"symbols":   []string{"BTC_KRW", "ETH_KRW", "SOL_KRW", "XRP_KRW"},
		"tickTypes": []string{"1H"},
	}
	if err := conn.WriteJSONSafe(sub); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	c.logger.Info("bithumb ws connected")

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

type bithumbWSMsg struct {
	Type    string `json:"type"`
	Content struct {
		List []struct {
			Symbol    string `json:"symbol"`    // "BTC_KRW"
			OrderType string `json:"orderType"` // "bid" or "ask"
			Price     string `json:"price"`
			Quantity  string `json:"quantity"`
		} `json:"list"`
	} `json:"content"`
}

func (c *BithumbWSClient) handleMessage(data []byte) {
	var msg bithumbWSMsg
	if err := json.Unmarshal(data, &msg); err != nil || msg.Type != "orderbookdepth" {
		return
	}

	// 按 symbol 收集 best bid/ask
	bids := make(map[string]decimal.Decimal)
	asks := make(map[string]decimal.Decimal)

	for _, item := range msg.Content.List {
		price, err := decimal.NewFromString(item.Price)
		if err != nil {
			continue
		}

		switch item.OrderType {
		case "bid":
			if cur, ok := bids[item.Symbol]; !ok || price.GreaterThan(cur) {
				bids[item.Symbol] = price
			}
		case "ask":
			if cur, ok := asks[item.Symbol]; !ok || price.LessThan(cur) {
				asks[item.Symbol] = price
			}
		}
	}

	for symbol, bid := range bids {
		ask, ok := asks[symbol]
		if !ok {
			continue
		}
		pair := bithumbSymbolToPair(symbol)
		c.cache.Update(pair, &OrderBookTop{Bid: bid, Ask: ask})
		c.ready.Store(true)
	}
}

// bithumbSymbolToPair "BTC_KRW" -> "BTC/KRW"
func bithumbSymbolToPair(symbol string) string {
	return strings.ReplaceAll(symbol, "_", "/")
}
