package model

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestTotalBalance(t *testing.T) {
	w := Wallet{
		Available:   decimal.NewFromFloat(1000),
		InOperation: decimal.NewFromFloat(500),
		Frozen:      decimal.NewFromFloat(200),
	}

	total := w.TotalBalance()
	expected := decimal.NewFromFloat(1700)

	if !total.Equal(expected) {
		t.Errorf("TotalBalance() = %s, want %s", total, expected)
	}
}

func TestWithdrawAddressInCooldown(t *testing.T) {
	addr := WithdrawAddress{
		CooldownUntil: time.Now().Add(1 * time.Hour),
	}
	if !addr.InCooldown() {
		t.Error("should be in cooldown")
	}

	addr.CooldownUntil = time.Now().Add(-1 * time.Hour)
	if addr.InCooldown() {
		t.Error("should not be in cooldown")
	}
}
