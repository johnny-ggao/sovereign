package service

import (
	"context"
	"errors"
	"log/slog"

	authrepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/provider"
	ntmpl "github.com/sovereign-fund/sovereign/internal/modules/notification/template"
	settingsmodel "github.com/sovereign-fund/sovereign/internal/modules/settings/model"
	settingsrepo "github.com/sovereign-fund/sovereign/internal/modules/settings/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"gorm.io/gorm"
)

// NotificationService handles event-driven email notifications.
type NotificationService interface {
	HandleDepositConfirmed(ctx context.Context, event events.Event) error
	HandleWithdrawCompleted(ctx context.Context, event events.Event) error
	HandleWithdrawFailed(ctx context.Context, event events.Event) error
	HandleSettlementCreated(ctx context.Context, event events.Event) error
	SendOTP(ctx context.Context, email, lang, templateName, code, expiresIn string) error
}

type notificationService struct {
	emailProvider provider.EmailProvider
	userRepo      authrepo.UserRepository
	settingsRepo  settingsrepo.SettingsRepository
	renderer      *ntmpl.Renderer
	logger        *slog.Logger
}

// NewNotificationService creates a NotificationService.
// It initialises the template renderer from templateDir.
func NewNotificationService(
	emailProvider provider.EmailProvider,
	userRepo authrepo.UserRepository,
	settingsRepo settingsrepo.SettingsRepository,
	templateDir string,
	logger *slog.Logger,
) (NotificationService, error) {
	renderer, err := ntmpl.NewRenderer(templateDir)
	if err != nil {
		return nil, err
	}
	return &notificationService{
		emailProvider: emailProvider,
		userRepo:      userRepo,
		settingsRepo:  settingsRepo,
		renderer:      renderer,
		logger:        logger,
	}, nil
}

// ---------------------------------------------------------------------------
// Event handlers
// ---------------------------------------------------------------------------

func (s *notificationService) HandleDepositConfirmed(ctx context.Context, event events.Event) error {
	return s.sendIfEnabled(ctx, event, "deposit_confirmed", func(pref *settingsmodel.NotificationPref) bool {
		return pref.EmailDeposit
	}, func(payload map[string]string) map[string]string {
		return map[string]string{
			"Amount":   payload["amount"],
			"Currency": payload["currency"],
			"Network":  payload["network"],
			"TxHash":   payload["tx_hash"],
		}
	})
}

func (s *notificationService) HandleWithdrawCompleted(ctx context.Context, event events.Event) error {
	return s.sendIfEnabled(ctx, event, "withdraw_completed", func(pref *settingsmodel.NotificationPref) bool {
		return pref.EmailWithdraw
	}, func(payload map[string]string) map[string]string {
		return map[string]string{
			"Amount":    payload["amount"],
			"Currency":  payload["currency"],
			"Network":   payload["network"],
			"ToAddress": payload["to_address"],
			"TxHash":    payload["tx_hash"],
		}
	})
}

func (s *notificationService) HandleWithdrawFailed(ctx context.Context, event events.Event) error {
	return s.sendIfEnabled(ctx, event, "withdraw_failed", func(pref *settingsmodel.NotificationPref) bool {
		return pref.EmailWithdraw
	}, func(payload map[string]string) map[string]string {
		return map[string]string{
			"Amount":   payload["amount"],
			"Currency": payload["currency"],
			"Reason":   payload["reason"],
		}
	})
}

func (s *notificationService) HandleSettlementCreated(ctx context.Context, event events.Event) error {
	return s.sendIfEnabled(ctx, event, "settlement_created", func(pref *settingsmodel.NotificationPref) bool {
		return pref.EmailSettlement
	}, func(payload map[string]string) map[string]string {
		return map[string]string{
			"Date":      payload["period"],
			"TotalPnL":  payload["total_pnl"],
			"UserShare": payload["net_return"],
			"FeeRate":   payload["fee_rate"],
		}
	})
}

// SendOTP sends an OTP email directly (no preference check).
func (s *notificationService) SendOTP(ctx context.Context, email, lang, templateName, code, expiresIn string) error {
	data := map[string]string{
		"OTPCode":   code,
		"ExpiresIn": expiresIn,
	}

	subject, html, err := s.renderer.Render(templateName, lang, data)
	if err != nil {
		s.logger.Error("failed to render OTP template",
			slog.String("template", templateName),
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := s.emailProvider.Send(ctx, provider.SendInput{
		To:      email,
		Subject: subject,
		HTML:    html,
	}); err != nil {
		s.logger.Error("failed to send OTP email",
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// ---------------------------------------------------------------------------
// Core helper
// ---------------------------------------------------------------------------

// sendIfEnabled is the shared logic for all event-driven notifications.
//
//  1. Extract user_id from payload
//  2. Look up user (email + language)
//  3. Check notification preference
//  4. Render template
//  5. Send email
//
// Errors are logged but never returned (never block the main flow).
func (s *notificationService) sendIfEnabled(
	ctx context.Context,
	event events.Event,
	templateName string,
	isEnabled func(*settingsmodel.NotificationPref) bool,
	buildData func(map[string]string) map[string]string,
) error {
	payload, ok := event.Payload.(map[string]string)
	if !ok {
		s.logger.Warn("invalid event payload type",
			slog.String("event_type", event.Type),
		)
		return nil
	}

	userID := payload["user_id"]
	if userID == "" {
		s.logger.Warn("event payload missing user_id",
			slog.String("event_type", event.Type),
		)
		return nil
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to find user for notification",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil
	}

	if user.Email == "" {
		s.logger.Warn("user has no email, skipping notification",
			slog.String("user_id", userID),
		)
		return nil
	}

	pref, err := s.settingsRepo.FindNotificationPref(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Default: all email notifications enabled
			pref = &settingsmodel.NotificationPref{
				EmailTrade:      true,
				EmailDeposit:    true,
				EmailWithdraw:   true,
				EmailSettlement: true,
			}
		} else {
			s.logger.Error("failed to find notification pref",
				slog.String("user_id", userID),
				slog.String("error", err.Error()),
			)
			return nil
		}
	}

	if !isEnabled(pref) {
		s.logger.Info("notification disabled by user preference",
			slog.String("user_id", userID),
			slog.String("template", templateName),
		)
		return nil
	}

	data := buildData(payload)

	subject, html, err := s.renderer.Render(templateName, user.Language, data)
	if err != nil {
		s.logger.Error("failed to render notification template",
			slog.String("template", templateName),
			slog.String("error", err.Error()),
		)
		return nil
	}

	if err := s.emailProvider.Send(ctx, provider.SendInput{
		To:      user.Email,
		Subject: subject,
		HTML:    html,
	}); err != nil {
		s.logger.Error("failed to send notification email",
			slog.String("email", user.Email),
			slog.String("template", templateName),
			slog.String("error", err.Error()),
		)
		return nil
	}

	s.logger.Info("notification email sent",
		slog.String("email", user.Email),
		slog.String("template", templateName),
	)

	return nil
}
