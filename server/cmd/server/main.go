package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/app"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/service"
	"github.com/sovereign-fund/sovereign/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to initialize app", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Cancellable context for graceful shutdown of background goroutines
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	go application.PremiumModule.Hub.Run(appCtx)

	// 启动 premium fetcher（实时行情采集 + WS 推送给前端）
	var krClients, glClients []service.ExchangeClient
	if cfg.Cobo.UseMock {
		krClients = []service.ExchangeClient{
			service.NewMockExchangeClient("upbit", 90000000, 2700000),
			service.NewMockExchangeClient("bithumb", 90000000, 2650000),
		}
		glClients = []service.ExchangeClient{
			service.NewMockExchangeClient("binance", 90000000, 0),
			service.NewMockExchangeClient("bybit", 90000000, 50000),
		}
	} else {
		upbit := service.NewUpbitWSClient(application.Logger)
		bithumb := service.NewBithumbWSClient(application.Logger)
		binance := service.NewBinanceWSClient(application.Logger)
		bybit := service.NewBybitWSClient(application.Logger)

		krClients = []service.ExchangeClient{upbit, bithumb}
		glClients = []service.ExchangeClient{binance, bybit}

		for _, ws := range []service.WSExchangeClient{upbit, bithumb, binance, bybit} {
			go ws.Start(appCtx)
		}
		// 等待 WS 就绪
		time.Sleep(2 * time.Second)
	}

	// 启动 scolkg.com 汇率获取
	rateProvider := service.NewScolkgRateProvider(application.Logger)
	go rateProvider.Start(appCtx)

	fetcher := worker.NewPremiumFetcher(application.PremiumModule.Service, krClients, glClients, rateProvider, application.Logger)
	go func() {
		ticker := time.NewTicker(cfg.Worker.PremiumFetchInterval)
		defer ticker.Stop()
		for {
			select {
			case <-appCtx.Done():
				return
			case <-ticker.C:
				fetcher.Run(appCtx)
			}
		}
	}()

	router := app.SetupRouter(application, appCtx)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		application.Logger.Info("server starting", slog.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			application.Logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Logger.Info("shutting down server...")

	appCancel()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		application.Logger.Error("server forced to shutdown", slog.String("error", err.Error()))
	}

	application.EventBus.Shutdown()
	application.Logger.Info("server stopped")
}
