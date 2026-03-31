// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/model"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/database"
	"github.com/sovereign-fund/sovereign/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config:", err)
	}

	l := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.NewPostgres(cfg.Database, l)
	if err != nil {
		log.Fatal("connect db:", err)
	}

	repo := repository.NewPremiumRepository(db)
	ctx := context.Background()

	pairs := []struct {
		pair    string
		krBase  float64
		glBase  float64
		srcKR   string
		srcGL   string
	}{
		{"BTC/KRW", 92700000, 90000000, "upbit", "binance"},
		{"ETH/KRW", 5150000, 5000000, "upbit", "binance"},
	}

	now := time.Now()
	// 生成过去 7 天的历史数据，每 2 分钟一条
	points := 7 * 24 * 30 // 7 天 * 每小时 30 条
	var ticks []model.PremiumTick

	for _, p := range pairs {
		krPrice := p.krBase
		glPrice := p.glBase

		for i := points; i > 0; i-- {
			ts := now.Add(-time.Duration(i) * 2 * time.Minute)

			// 随机游走
			krPrice += krPrice * (rand.Float64() - 0.5) * 0.004
			glPrice += glPrice * (rand.Float64() - 0.5) * 0.003

			// 限制范围
			if krPrice < p.krBase*0.92 {
				krPrice = p.krBase * 0.92
			}
			if krPrice > p.krBase*1.08 {
				krPrice = p.krBase * 1.08
			}
			if glPrice < p.glBase*0.92 {
				glPrice = p.glBase * 0.92
			}
			if glPrice > p.glBase*1.08 {
				glPrice = p.glBase * 1.08
			}

			kr := decimal.NewFromFloat(krPrice).Round(0)
			gl := decimal.NewFromFloat(glPrice).Round(0)
			pct := kr.Sub(gl).Div(gl).Mul(decimal.NewFromInt(100)).Round(4)

			ticks = append(ticks, model.PremiumTick{
				Pair:        p.pair,
				KoreanPrice: kr,
				GlobalPrice: gl,
				PremiumPct:  pct,
				SourceKR:    p.srcKR,
				SourceGL:    p.srcGL,
				CreatedAt:   ts,
			})
		}
	}

	// 分批插入
	batchSize := 500
	for i := 0; i < len(ticks); i += batchSize {
		end := i + batchSize
		if end > len(ticks) {
			end = len(ticks)
		}
		if err := repo.CreateBatch(ctx, ticks[i:end]); err != nil {
			log.Fatal("insert batch:", err)
		}
	}

	fmt.Printf("Seeded %d premium ticks for %d pairs\n", len(ticks), len(pairs))
}
