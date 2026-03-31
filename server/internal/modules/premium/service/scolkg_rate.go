package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const scolkgPollingURL = "https://ticker1.cryprice.com/socket.io/?EIO=4&transport=polling"

// ScolkgRateProvider 从 scolkg.com (ticker1.cryprice.com) 获取 USDT/KRW 汇率
type ScolkgRateProvider struct {
	mu     sync.RWMutex
	rate   decimal.Decimal
	logger *slog.Logger
}

func NewScolkgRateProvider(logger *slog.Logger) *ScolkgRateProvider {
	return &ScolkgRateProvider{
		rate:   decimal.NewFromFloat(1500), // 初始 fallback
		logger: logger,
	}
}

func (p *ScolkgRateProvider) Rate() decimal.Decimal {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.rate
}

// Start 启动定期轮询获取汇率
func (p *ScolkgRateProvider) Start(ctx context.Context) {
	// 立即获取一次
	p.fetch()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.fetch()
		}
	}
}

func (p *ScolkgRateProvider) fetch() {
	rate, err := p.fetchRate()
	if err != nil {
		p.logger.Warn("scolkg rate fetch failed", slog.String("error", err.Error()))
		return
	}

	p.mu.Lock()
	p.rate = rate
	p.mu.Unlock()

	p.logger.Debug("scolkg USDT/KRW rate updated", slog.String("rate", rate.String()))
}

func (p *ScolkgRateProvider) fetchRate() (decimal.Decimal, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 1. 握手获取 sid
	resp, err := client.Get(scolkgPollingURL)
	if err != nil {
		return decimal.Zero, fmt.Errorf("handshake: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 解析 sid：响应格式 "0{...}"
	if len(bodyStr) < 2 || bodyStr[0] != '0' {
		return decimal.Zero, fmt.Errorf("unexpected handshake: %s", bodyStr[:min(len(bodyStr), 100)])
	}

	var handshake struct {
		SID string `json:"sid"`
	}
	if err := json.Unmarshal([]byte(bodyStr[1:]), &handshake); err != nil {
		return decimal.Zero, fmt.Errorf("parse handshake: %w", err)
	}

	sid := handshake.SID

	// 2. 发送连接确认 (40)
	pollURL := fmt.Sprintf("%s&sid=%s", scolkgPollingURL, sid)
	req, _ := http.NewRequest("POST", pollURL, strings.NewReader("40"))
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	resp2, err := client.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("connect: %w", err)
	}
	resp2.Body.Close()

	// 3. 多次轮询直到获取到 usdtprice
	for attempt := 0; attempt < 5; attempt++ {
		time.Sleep(time.Duration(500+attempt*300) * time.Millisecond)

		resp3, err := client.Get(pollURL)
		if err != nil {
			continue
		}

		data, _ := io.ReadAll(resp3.Body)
		resp3.Body.Close()
		dataStr := string(data)

		// 解析 usdtprice response
		// 格式: ...42["usdtprice response",1529.405]...
		idx := strings.Index(dataStr, `"usdtprice response",`)
		if idx == -1 {
			continue
		}

		start := idx + len(`"usdtprice response",`)
		end := strings.Index(dataStr[start:], "]")
		if end == -1 {
			continue
		}

		rateStr := strings.TrimSpace(dataStr[start : start+end])
		rate, err := decimal.NewFromString(rateStr)
		if err != nil {
			continue
		}

		return rate, nil
	}

	return decimal.Zero, fmt.Errorf("usdtprice not found after 5 polling attempts")
}

