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
	"gorm.io/gorm/clause"
)

type Wallet struct {
	ID          string          `gorm:"primaryKey"`
	UserID      string          `gorm:"not null"`
	Currency    string          `gorm:"not null"`
	Available   decimal.Decimal `gorm:"type:decimal(28,18);default:0"`
	InOperation decimal.Decimal `gorm:"type:decimal(28,18);default:0"`
	Frozen      decimal.Decimal `gorm:"type:decimal(28,18);default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Transaction struct {
	ID          string           `gorm:"primaryKey"`
	UserID      string           `gorm:"not null"`
	Type        string           `gorm:"not null"`
	Currency    string           `gorm:"not null"`
	Network     string           `gorm:"default:''"`
	Amount      decimal.Decimal  `gorm:"type:decimal(28,18);not null"`
	Fee         decimal.Decimal  `gorm:"type:decimal(28,18);default:0"`
	Address     string           `gorm:"default:''"`
	TxHash      string           `gorm:"default:''"`
	Status      string           `gorm:"default:'pending'"`
	ExternalID  string           `gorm:"default:''"`
	ConfirmedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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

	ctx := context.Background()

	// 查找第一个用户
	var userID string
	if err := db.WithContext(ctx).Raw("SELECT id FROM users LIMIT 1").Scan(&userID).Error; err != nil {
		log.Fatal("no users found:", err)
	}
	if userID == "" {
		log.Fatal("no users in database, please login first")
	}

	fmt.Printf("Seeding wallet data for user: %s\n", userID)

	seedWallets(db, ctx, userID)
	count := seedTransactions(db, ctx, userID)

	fmt.Printf("Seeded 3 wallets and %d transactions\n", count)
}

func seedWallets(db *gorm.DB, ctx context.Context, userID string) {
	wallets := []Wallet{
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Currency:    "USDT",
			Available:   decimal.NewFromFloat(52340.50),
			InOperation: decimal.NewFromFloat(25000),
			Frozen:      decimal.Zero,
		},
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Currency:    "BTC",
			Available:   decimal.NewFromFloat(1.2345),
			InOperation: decimal.NewFromFloat(0.5),
			Frozen:      decimal.Zero,
		},
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Currency:    "ETH",
			Available:   decimal.NewFromFloat(15.789),
			InOperation: decimal.NewFromFloat(5.0),
			Frozen:      decimal.NewFromFloat(2.0),
		},
	}

	for _, w := range wallets {
		result := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "currency"}},
			DoUpdates: clause.AssignmentColumns([]string{"available", "in_operation", "frozen", "updated_at"}),
		}).Create(&w)
		if result.Error != nil {
			log.Fatalf("create wallet %s: %v", w.Currency, result.Error)
		}
	}
}

func seedTransactions(db *gorm.DB, ctx context.Context, userID string) int {
	now := time.Now()
	networks := []string{"ERC20", "TRC20", "BEP20"}

	type txTemplate struct {
		txType   string
		currency string
		minAmt   float64
		maxAmt   float64
		fee      float64
	}

	templates := []txTemplate{
		{"deposit", "USDT", 1000, 20000, 0},
		{"deposit", "USDT", 5000, 50000, 0},
		{"deposit", "BTC", 0.05, 0.5, 0},
		{"deposit", "ETH", 1.0, 10.0, 0},
		{"withdraw", "USDT", 500, 10000, 1.0},
		{"withdraw", "BTC", 0.01, 0.2, 0.0001},
		{"withdraw", "ETH", 0.5, 5.0, 0.001},
	}

	addresses := map[string][]string{
		"USDT": {
			"0x742d35Cc6634C0532925a3b844Bc9e7595f2bD68",
			"TN3W4H6rK2ce4vX9YnFQHwKENnHjoxb3m9",
			"0x1234567890abcdef1234567890abcdef12345678",
		},
		"BTC": {
			"bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
			"3FZbgi29cpjq2GjdwV8eyHuJJnkLtktZc5",
			"bc1q9h5yjqk3t8x2z5m3nrgr0c5t7e8djwfsn3m6y",
		},
		"ETH": {
			"0xdAC17F958D2ee523a2206206994597C13D831ec7",
			"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
			"0x6B175474E89094C44Da98b954EedeAC495271d0F",
		},
	}

	statuses := []string{"confirmed", "confirmed", "confirmed", "confirmed", "processing", "pending"}

	var txs []Transaction

	// 生成过去 90 天的交易记录
	for i := 0; i < 45; i++ {
		tmpl := templates[rand.IntN(len(templates))]
		daysAgo := rand.IntN(90)
		hoursAgo := rand.IntN(24)
		createdAt := now.AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo) * time.Hour)

		amount := tmpl.minAmt + rand.Float64()*(tmpl.maxAmt-tmpl.minAmt)
		status := statuses[rand.IntN(len(statuses))]
		network := networks[rand.IntN(len(networks))]

		addrs := addresses[tmpl.currency]
		addr := addrs[rand.IntN(len(addrs))]

		var confirmedAt *time.Time
		txHash := ""
		if status == "confirmed" {
			t := createdAt.Add(time.Duration(5+rand.IntN(55)) * time.Minute)
			confirmedAt = &t
			txHash = fmt.Sprintf("0x%032x%032x", rand.Uint64(), rand.Uint64())
		}

		txs = append(txs, Transaction{
			ID:          uuid.New().String(),
			UserID:      userID,
			Type:        tmpl.txType,
			Currency:    tmpl.currency,
			Network:     network,
			Amount:      decimal.NewFromFloat(amount).Round(8),
			Fee:         decimal.NewFromFloat(tmpl.fee),
			Address:     addr,
			TxHash:      txHash,
			Status:      status,
			ExternalID:  fmt.Sprintf("cobo_%s", uuid.New().String()[:8]),
			ConfirmedAt: confirmedAt,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		})
	}

	if err := db.WithContext(ctx).CreateInBatches(txs, 50).Error; err != nil {
		log.Fatal("create transactions:", err)
	}

	return len(txs)
}
