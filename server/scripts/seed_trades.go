// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/shared/database"
	"github.com/sovereign-fund/sovereign/pkg/logger"
	"gorm.io/gorm"
)

type Trade struct {
	ID           string          `gorm:"primaryKey"`
	InvestmentID string          `gorm:"not null"`
	UserID       string          `gorm:"not null"`
	Pair         string          `gorm:"not null"`
	BuyExchange  string          `gorm:"not null"`
	SellExchange string          `gorm:"not null"`
	BuyPrice     decimal.Decimal `gorm:"type:decimal(28,8);not null"`
	SellPrice    decimal.Decimal `gorm:"type:decimal(28,8);not null"`
	Amount       decimal.Decimal `gorm:"type:decimal(28,18);not null"`
	PremiumPct   decimal.Decimal `gorm:"type:decimal(8,4);not null"`
	PnL          decimal.Decimal `gorm:"column:pnl;type:decimal(28,18);not null"`
	Fee          decimal.Decimal `gorm:"type:decimal(28,18);default:0"`
	ExecutedAt   time.Time       `gorm:"not null"`
	CreatedAt    time.Time
}

type Investment struct {
	ID     string `gorm:"primaryKey"`
	UserID string
	Status string
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	l := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.NewPostgres(cfg.Database, l)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 查找 active 的投资
	var investments []Investment
	if err := db.WithContext(ctx).Table("investments").Where("status = 'active'").Find(&investments).Error; err != nil {
		log.Fatal("find investments:", err)
	}

	if len(investments) == 0 {
		log.Fatal("no active investments found")
	}

	fmt.Printf("Found %d active investments\n", len(investments))

	pairs := []string{"BTC/KRW", "ETH/KRW", "SOL/KRW", "XRP/KRW"}
	buyExchanges := []string{"binance", "bybit"}
	sellExchanges := []string{"upbit", "bithumb"}
	now := time.Now()

	var trades []Trade
	for _, inv := range investments {
		// 每个投资生成 30-50 笔交易，分布在过去 30 天
		tradeCount := 30 + rand.IntN(20)
		for i := 0; i < tradeCount; i++ {
			pair := pairs[rand.IntN(len(pairs))]
			buyEx := buyExchanges[rand.IntN(len(buyExchanges))]
			sellEx := sellExchanges[rand.IntN(len(sellExchanges))]

			// 随机时间：过去 30 天
			daysAgo := rand.IntN(30)
			hoursAgo := rand.IntN(24)
			executedAt := now.AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo) * time.Hour)

			// 根据交易对生成合理的买入价（USDT 计价 * 汇率 = KRW）
			var buyPrice float64
			switch pair {
			case "BTC/KRW":
				buyPrice = 130000000 + rand.Float64()*5000000
			case "ETH/KRW":
				buyPrice = 5200000 + rand.Float64()*200000
			case "SOL/KRW":
				buyPrice = 350000 + rand.Float64()*20000
			case "XRP/KRW":
				buyPrice = 4500 + rand.Float64()*300
			}

			// 溢价 0.5% ~ 3.5%
			premiumPct := 0.5 + rand.Float64()*3.0
			sellPrice := buyPrice * (1 + premiumPct/100)

			// 交易金额（USDT）
			amount := 50 + rand.Float64()*200

			// PnL = amount * premiumPct / 100
			pnl := amount * premiumPct / 100

			// 手续费 = PnL * 0.1%
			fee := pnl * 0.001

			trades = append(trades, Trade{
				ID:           uuid.New().String(),
				InvestmentID: inv.ID,
				UserID:       inv.UserID,
				Pair:         pair,
				BuyExchange:  buyEx,
				SellExchange: sellEx,
				BuyPrice:     decimal.NewFromFloat(buyPrice).Round(0),
				SellPrice:    decimal.NewFromFloat(sellPrice).Round(0),
				Amount:       decimal.NewFromFloat(amount).Round(6),
				PremiumPct:   decimal.NewFromFloat(premiumPct).Round(4),
				PnL:          decimal.NewFromFloat(pnl).Round(6),
				Fee:          decimal.NewFromFloat(fee).Round(6),
				ExecutedAt:   executedAt,
				CreatedAt:    executedAt,
			})
		}
	}

	if err := db.WithContext(ctx).CreateInBatches(trades, 100).Error; err != nil {
		log.Fatal("create trades:", err)
	}

	// 更新投资的收益
	updateInvestmentReturns(db, ctx, investments)

	fmt.Printf("Seeded %d trades for %d investments\n", len(trades), len(investments))
}

func updateInvestmentReturns(db *gorm.DB, ctx context.Context, investments []Investment) {
	for _, inv := range investments {
		var totalPnL, totalFee decimal.Decimal
		db.WithContext(ctx).Table("trades").
			Where("investment_id = ?", inv.ID).
			Select("COALESCE(SUM(pnl), 0) as total_pnl, COALESCE(SUM(fee), 0) as total_fee").
			Row().Scan(&totalPnL, &totalFee)

		perfFee := decimal.Zero
		if totalPnL.GreaterThan(decimal.Zero) {
			perfFee = totalPnL.Mul(decimal.NewFromFloat(0.5))
		}
		netReturn := totalPnL.Sub(perfFee)

		db.WithContext(ctx).Table("investments").
			Where("id = ?", inv.ID).
			Updates(map[string]interface{}{
				"total_return":   totalPnL,
				"performance_fee": perfFee,
				"net_return":     netReturn,
			})

		fmt.Printf("  Investment %s: PnL=%s, Fee=%s, Net=%s\n", inv.ID[:8], totalPnL, perfFee, netReturn)
	}
}
