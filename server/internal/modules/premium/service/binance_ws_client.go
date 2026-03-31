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

const binanceWSURL = "wss://stream.binance.com:9443/stream"

type BinanceWSClient struct {
	cache  *OrderBookCache
	logger *slog.Logger
	ready  atomic.Bool
}

func NewBinanceWSClient(logger *slog.Logger) *BinanceWSClient {
	return &BinanceWSClient{
		cache:  NewOrderBookCache(),
		logger: logger,
	}
}

func (c *BinanceWSClient) Name() string         { return "binance" }
func (c *BinanceWSClient) Latency() time.Duration { return c.cache.Latency() }

func (c *BinanceWSClient) Ready() bool { return c.ready.Load() }

func (c *BinanceWSClient) GetOrderBook(_ context.Context, pair string) (*OrderBookTop, error) {
	return c.cache.Get(pair)
}

func (c *BinanceWSClient) Start(ctx context.Context) error {
	backoff := NewBackoff()

	for {
		if err := c.connect(ctx); err != nil {
			c.logger.Warn("binance ws disconnected", slog.String("error", err.Error()))
			c.ready.Store(false)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wait := backoff.Next()
		c.logger.Info("binance ws reconnecting", slog.Duration("delay", wait))
		if err := SleepWithContext(ctx, wait); err != nil {
			return err
		}
	}
}

func (c *BinanceWSClient) connect(ctx context.Context) error {
	raw, _, err := websocket.DefaultDialer.DialContext(ctx, binanceWSURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer raw.Close()
	conn := NewWSConn(raw)

	sub := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			"btcusdt@depth5@100ms",
			"ethusdt@depth5@100ms",
			"solusdt@depth5@100ms",
			"xrpusdt@depth5@100ms",
		},
		"id": 1,
	}
	if err := conn.WriteJSONSafe(sub); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	c.logger.Info("binance ws connected")

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

type binanceWSStream struct {
	Stream string          `json:"stream"` // "btcusdt@depth5@100ms"
	Data   json.RawMessage `json:"data"`
}

type binanceWSDepth struct {
	Bids [][]string `json:"bids"` // [["price","qty"],...]
	Asks [][]string `json:"asks"`
}

func (c *BinanceWSClient) handleMessage(data []byte) {
	var stream binanceWSStream
	if err := json.Unmarshal(data, &stream); err != nil || stream.Stream == "" {
		return
	}

	var depth binanceWSDepth
	if err := json.Unmarshal(stream.Data, &depth); err != nil {
		return
	}

	if len(depth.Bids) == 0 || len(depth.Asks) == 0 {
		return
	}

	bid, _ := decimal.NewFromString(depth.Bids[0][0])
	ask, _ := decimal.NewFromString(depth.Asks[0][0])

	pair := binanceStreamToPair(stream.Stream)

	c.cache.Update(pair, &OrderBookTop{
		Bid: bid,
		Ask: ask,
	})

	c.ready.Store(true)
}

// binanceStreamToPair "btcusdt@depth5@100ms" -> "BTC/KRW"
func binanceStreamToPair(stream string) string {
	symbol := strings.Split(stream, "@")[0]              // "btcusdt"
	base := strings.TrimSuffix(strings.ToUpper(symbol), "USDT") // "BTC"
	return base + "/KRW"
}
