package jwt

import (
	"testing"
	"time"

	"github.com/sovereign-fund/sovereign/config"
)

func newTestManager() *Manager {
	return NewManager(config.JWTConfig{
		AccessSecret:  "test-access-secret-32-characters!",
		RefreshSecret: "test-refresh-secret-32-characters",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "sovereign-test",
	})
}

func TestGenerateAndValidate(t *testing.T) {
	mgr := newTestManager()

	pair, err := mgr.Generate("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("tokens should not be empty")
	}

	claims, err := mgr.ValidateAccess(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccess() error = %v", err)
	}

	if claims.UserID != "user-123" {
		t.Fatalf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Email != "test@example.com" {
		t.Fatalf("Email = %q, want %q", claims.Email, "test@example.com")
	}

	claims, err = mgr.ValidateRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ValidateRefresh() error = %v", err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("UserID = %q, want %q", claims.UserID, "user-123")
	}
}

func TestValidateInvalidToken(t *testing.T) {
	mgr := newTestManager()

	_, err := mgr.ValidateAccess("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestAccessTokenCannotValidateAsRefresh(t *testing.T) {
	mgr := newTestManager()

	pair, _ := mgr.Generate("user-123", "test@example.com")

	_, err := mgr.ValidateRefresh(pair.AccessToken)
	if err == nil {
		t.Fatal("access token should not validate as refresh token")
	}
}
