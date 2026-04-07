package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	authmodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	"github.com/sovereign-fund/sovereign/internal/modules/notification/provider"
	settingsmodel "github.com/sovereign-fund/sovereign/internal/modules/settings/model"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
)

// ---------------------------------------------------------------------------
// Stub: UserRepository
// ---------------------------------------------------------------------------

type stubUserRepo struct {
	user *authmodel.User
	err  error
}

func (s *stubUserRepo) FindByID(_ context.Context, _ string) (*authmodel.User, error) {
	return s.user, s.err
}
func (s *stubUserRepo) FindByEmail(_ context.Context, _ string) (*authmodel.User, error) {
	return nil, nil
}
func (s *stubUserRepo) FindByGoogleID(_ context.Context, _ string) (*authmodel.User, error) {
	return nil, nil
}
func (s *stubUserRepo) Create(_ context.Context, _ *authmodel.User) error  { return nil }
func (s *stubUserRepo) Update(_ context.Context, _ *authmodel.User) error  { return nil }
func (s *stubUserRepo) ExistsByEmail(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// ---------------------------------------------------------------------------
// Stub: SettingsRepository
// ---------------------------------------------------------------------------

type stubSettingsRepo struct {
	pref *settingsmodel.NotificationPref
	err  error
}

func (s *stubSettingsRepo) FindNotificationPref(_ context.Context, _ string) (*settingsmodel.NotificationPref, error) {
	return s.pref, s.err
}
func (s *stubSettingsRepo) UpsertNotificationPref(_ context.Context, _ *settingsmodel.NotificationPref) error {
	return nil
}
func (s *stubSettingsRepo) FindLoginDevices(_ context.Context, _ string) ([]settingsmodel.LoginDevice, error) {
	return nil, nil
}
func (s *stubSettingsRepo) UpsertLoginDevice(_ context.Context, _ *settingsmodel.LoginDevice) error {
	return nil
}
func (s *stubSettingsRepo) DeleteLoginDevice(_ context.Context, _, _ string) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newTestService(t *testing.T, ep provider.EmailProvider, ur *stubUserRepo, sr *stubSettingsRepo) NotificationService {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc, err := NewNotificationService(ep, ur, sr, "../template/emails", logger)
	if err != nil {
		t.Fatalf("NewNotificationService: %v", err)
	}
	return svc
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHandleDepositConfirmedSendsEmail(t *testing.T) {
	mock := &provider.MockProvider{}
	userRepo := &stubUserRepo{
		user: &authmodel.User{
			ID:       "u1",
			Email:    "alice@example.com",
			Language: "en",
		},
	}
	settingsRepo := &stubSettingsRepo{
		pref: &settingsmodel.NotificationPref{
			UserID:       "u1",
			EmailDeposit: true,
		},
	}

	svc := newTestService(t, mock, userRepo, settingsRepo)

	evt := events.Event{
		Type: events.DepositConfirmed,
		Payload: map[string]string{
			"user_id":  "u1",
			"amount":   "1000",
			"currency": "USDT",
			"network":  "Ethereum",
			"tx_hash":  "0xabc123",
		},
	}

	if err := svc.HandleDepositConfirmed(context.Background(), evt); err != nil {
		t.Fatalf("HandleDepositConfirmed returned error: %v", err)
	}

	if len(mock.Sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(mock.Sent))
	}

	sent := mock.Sent[0]
	if sent.To != "alice@example.com" {
		t.Errorf("expected To=alice@example.com, got %s", sent.To)
	}
	if sent.Subject == "" {
		t.Error("expected non-empty subject")
	}
	if sent.HTML == "" {
		t.Error("expected non-empty HTML body")
	}
}

func TestHandleDepositConfirmedSkipsWhenDisabled(t *testing.T) {
	mock := &provider.MockProvider{}
	userRepo := &stubUserRepo{
		user: &authmodel.User{
			ID:       "u2",
			Email:    "bob@example.com",
			Language: "en",
		},
	}
	settingsRepo := &stubSettingsRepo{
		pref: &settingsmodel.NotificationPref{
			UserID:       "u2",
			EmailDeposit: false,
		},
	}

	svc := newTestService(t, mock, userRepo, settingsRepo)

	evt := events.Event{
		Type: events.DepositConfirmed,
		Payload: map[string]string{
			"user_id":  "u2",
			"amount":   "500",
			"currency": "USDC",
			"network":  "Polygon",
			"tx_hash":  "0xdef456",
		},
	}

	if err := svc.HandleDepositConfirmed(context.Background(), evt); err != nil {
		t.Fatalf("HandleDepositConfirmed returned error: %v", err)
	}

	if len(mock.Sent) != 0 {
		t.Fatalf("expected 0 emails sent, got %d", len(mock.Sent))
	}
}

func TestHandleDepositConfirmedSkipsEmptyEmail(t *testing.T) {
	mock := &provider.MockProvider{}
	userRepo := &stubUserRepo{
		user: &authmodel.User{
			ID:       "u3",
			Email:    "",
			Language: "en",
		},
	}
	settingsRepo := &stubSettingsRepo{
		pref: &settingsmodel.NotificationPref{
			UserID:       "u3",
			EmailDeposit: true,
		},
	}

	svc := newTestService(t, mock, userRepo, settingsRepo)

	evt := events.Event{
		Type: events.DepositConfirmed,
		Payload: map[string]string{
			"user_id":  "u3",
			"amount":   "200",
			"currency": "USDT",
			"network":  "Tron",
			"tx_hash":  "0xghi789",
		},
	}

	if err := svc.HandleDepositConfirmed(context.Background(), evt); err != nil {
		t.Fatalf("HandleDepositConfirmed returned error: %v", err)
	}

	if len(mock.Sent) != 0 {
		t.Fatalf("expected 0 emails sent, got %d", len(mock.Sent))
	}
}
