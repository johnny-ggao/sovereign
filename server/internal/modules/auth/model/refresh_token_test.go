package model

import (
	"testing"
	"time"
)

func TestRefreshTokenIsExpired(t *testing.T) {
	expired := RefreshToken{ExpiresAt: time.Now().Add(-1 * time.Hour)}
	if !expired.IsExpired() {
		t.Error("should be expired")
	}

	valid := RefreshToken{ExpiresAt: time.Now().Add(1 * time.Hour)}
	if valid.IsExpired() {
		t.Error("should not be expired")
	}
}
