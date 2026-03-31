package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/auth"
	"github.com/sovereign-fund/sovereign/internal/modules/dashboard"
	"github.com/sovereign-fund/sovereign/internal/modules/investment"
	"github.com/sovereign-fund/sovereign/internal/modules/premium"
	"github.com/sovereign-fund/sovereign/internal/modules/settings"
	"github.com/sovereign-fund/sovereign/internal/modules/settlement"
	"github.com/sovereign-fund/sovereign/internal/modules/tradelog"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet"
	"github.com/sovereign-fund/sovereign/internal/shared/database"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	jwtpkg "github.com/sovereign-fund/sovereign/pkg/jwt"
	"github.com/sovereign-fund/sovereign/pkg/logger"
	"gorm.io/gorm"
)

type App struct {
	Config     *config.Config
	DB         *gorm.DB
	Redis      *redis.Client
	Logger     *slog.Logger
	EventBus   events.Bus
	JWTManager *jwtpkg.Manager

	AuthModule       *auth.Module
	WalletModule     *wallet.Module
	PremiumModule    *premium.Module
	DashboardModule  *dashboard.Module
	InvestmentModule *investment.Module
	TradeLogModule   *tradelog.Module
	SettlementModule *settlement.Module
	SettingsModule   *settings.Module
}

func New(cfg *config.Config) (*App, error) {
	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	db, err := database.NewPostgres(cfg.Database, log)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	rdb, err := database.NewRedis(cfg.Redis, log)
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}

	jwtMgr := jwtpkg.NewManager(cfg.JWT)
	bus := events.NewBus(log)

	var walletProvider cobo.WalletProvider
	if cfg.Cobo.UseMock {
		walletProvider = cobo.NewMockProvider()
		log.Info("using mock wallet provider")
	} else {
		var err error
		walletProvider, err = cobo.NewCoboProvider(cobo.Options{
			BaseURL:       cfg.Cobo.BaseURL,
			APISecret:     cfg.Cobo.APISecret,
			APIPubKey:     cfg.Cobo.APIPubKey,
			WalletID:      cfg.Cobo.WalletID,
			WebhookPubKey: cfg.Cobo.WebhookPubKey,
		})
		if err != nil {
			return nil, fmt.Errorf("init cobo provider: %w", err)
		}
		log.Info("using cobo wallet provider", slog.String("base_url", cfg.Cobo.BaseURL))
	}

	walletMod := wallet.NewModule(db, walletProvider, bus, log, cfg.Wallet.AddressCooldown)

	// 用户注册时自动初始化钱包 + 生成充值地址
	bus.Subscribe(events.UserRegistered, func(ctx context.Context, event events.Event) error {
		payload, ok := event.Payload.(map[string]string)
		if !ok {
			return nil
		}
		userID := payload["user_id"]
		if userID == "" {
			return nil
		}
		if err := walletMod.Service.InitWallets(ctx, userID, cfg.Wallet.Currencies); err != nil {
			return err
		}
		return walletMod.Service.InitDepositAddresses(ctx, userID, []string{"BEP20", "TRC20"})
	})

	return &App{
		Config:           cfg,
		DB:               db,
		Redis:            rdb,
		Logger:           log,
		EventBus:         bus,
		JWTManager:       jwtMgr,
		AuthModule:       auth.NewModule(db, rdb, jwtMgr, bus, cfg, log),
		WalletModule:     walletMod,
		PremiumModule:    premium.NewModule(db, log),
		DashboardModule:  dashboard.NewModule(db, log),
		InvestmentModule: investment.NewModule(db, bus, log),
		TradeLogModule:   tradelog.NewModule(db, log),
		SettlementModule: settlement.NewModule(db, log),
		SettingsModule:   settings.NewModule(db, log),
	}, nil
}
