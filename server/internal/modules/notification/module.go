package notification

import (
	"context"
	"fmt"
	"log/slog"

	authrepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/provider"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/service"
	settingsrepo "github.com/sovereign-fund/sovereign/internal/modules/settings/repository"
	"github.com/sovereign-fund/sovereign/config"
)

type Module struct {
	Service service.NotificationService
}

func NewModule(
	cfg config.NotificationConfig,
	userRepo authrepo.UserRepository,
	settingsRepo settingsrepo.SettingsRepository,
	templateDir string,
	logger *slog.Logger,
) (*Module, error) {
	var emailProvider provider.EmailProvider

	if cfg.UseMock {
		emailProvider = provider.NewMockProvider()
		logger.Info("notification: using mock email provider")
	} else {
		var err error
		emailProvider, err = provider.NewSESProvider(context.Background(), cfg.AWSRegion, cfg.FromName, cfg.FromAddress)
		if err != nil {
			return nil, fmt.Errorf("init SES provider: %w", err)
		}
		logger.Info("notification: using AWS SES provider",
			slog.String("region", cfg.AWSRegion),
			slog.String("from", cfg.FromAddress),
		)
	}

	svc, err := service.NewNotificationService(emailProvider, userRepo, settingsRepo, templateDir, logger)
	if err != nil {
		return nil, fmt.Errorf("init notification service: %w", err)
	}

	return &Module{Service: svc}, nil
}
