package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// BinanceClient Binance 交易所订单簿客户端
// 返回 USDT 计价价格，由 PremiumFetcher 负责汇率转换
type BinanceClient struct {
	httpClient *http.Client
}

func NewBinanceClient() ExchangeClient {
	return &BinanceClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *BinanceClient) Name() string            { return "binance" }
func (c *BinanceClient) Latency() time.Duration  { return 0 }

// pairToBinanceSymbol 将内部交易对格式转为 Binance symbol
// BTC/KRW → BTCUSDT（Binance 以 USDT 计价）
func pairToBinanceSymbol(pair string) string {
	parts := strings.Split(pair, "/")
	if len(parts) != 2 {
		return pair
	}
	return parts[0] + "USDT"
}

type binanceOrderBookResp struct {
	Bids [][]json.RawMessage `json:"bids"` // [[price, qty], ...]
	Asks [][]json.RawMessage `json:"asks"` // [[price, qty], ...]
}

func (c *BinanceClient) GetOrderBook(ctx context.Context, pair string) (*OrderBookTop, error) {
	symbol := pairToBinanceSymbol(pair)
	url := fmt.Sprintf("https://api.binance.com/api/v3/depth?symbol=%s&limit=1", symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("binance orderbook request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance api error: %d %s", resp.StatusCode, string(body))
	}

	var ob binanceOrderBookResp
	if err := json.Unmarshal(body, &ob); err != nil {
		return nil, fmt.Errorf("parse binance orderbook: %w", err)
	}

	if len(ob.Bids) == 0 || len(ob.Asks) == 0 {
		return nil, fmt.Errorf("empty orderbook for %s", symbol)
	}

	bidPrice, err := parseRawPrice(ob.Bids[0][0])
	if err != nil {
		return nil, fmt.Errorf("parse bid price: %w", err)
	}

	askPrice, err := parseRawPrice(ob.Asks[0][0])
	if err != nil {
		return nil, fmt.Errorf("parse ask price: %w", err)
	}

	return &OrderBookTop{
		Bid: bidPrice,
		Ask: askPrice,
	}, nil
}

func parseRawPrice(raw json.RawMessage) (decimal.Decimal, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromString(s)
}
