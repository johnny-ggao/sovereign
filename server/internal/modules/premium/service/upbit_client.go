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

// UpbitClient Upbit 交易所订单簿客户端
type UpbitClient struct {
	httpClient *http.Client
}

func NewUpbitClient() ExchangeClient {
	return &UpbitClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *UpbitClient) Name() string            { return "upbit" }
func (c *UpbitClient) Latency() time.Duration  { return 0 }

// pairToUpbitMarket 将内部交易对格式转为 Upbit 市场代码
// BTC/KRW → KRW-BTC
func pairToUpbitMarket(pair string) string {
	parts := strings.Split(pair, "/")
	if len(parts) != 2 {
		return pair
	}
	return parts[1] + "-" + parts[0]
}

type upbitOrderBookResp struct {
	Market         string `json:"market"`
	OrderbookUnits []struct {
		AskPrice decimal.Decimal `json:"ask_price"`
		BidPrice decimal.Decimal `json:"bid_price"`
		AskSize  decimal.Decimal `json:"ask_size"`
		BidSize  decimal.Decimal `json:"bid_size"`
	} `json:"orderbook_units"`
}

func (c *UpbitClient) GetOrderBook(ctx context.Context, pair string) (*OrderBookTop, error) {
	market := pairToUpbitMarket(pair)
	url := fmt.Sprintf("https://api.upbit.com/v1/orderbook?markets=%s", market)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upbit orderbook request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upbit api error: %d %s", resp.StatusCode, string(body))
	}

	var items []upbitOrderBookResp
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("parse upbit orderbook: %w", err)
	}

	if len(items) == 0 || len(items[0].OrderbookUnits) == 0 {
		return nil, fmt.Errorf("empty orderbook for %s", market)
	}

	top := items[0].OrderbookUnits[0]
	return &OrderBookTop{
		Bid: top.BidPrice,
		Ask: top.AskPrice,
	}, nil
}
