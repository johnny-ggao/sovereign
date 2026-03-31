package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateFeeProfit(t *testing.T) {
	gross := decimal.NewFromFloat(1000)
	rate := decimal.NewFromFloat(0.5)

	fee, net := CalculateFee(gross, rate)

	expectedFee := decimal.NewFromFloat(500)
	expectedNet := decimal.NewFromFloat(500)

	if !fee.Equal(expectedFee) {
		t.Errorf("fee = %s, want %s", fee, expectedFee)
	}
	if !net.Equal(expectedNet) {
		t.Errorf("net = %s, want %s", net, expectedNet)
	}
}

func TestCalculateFeeLoss(t *testing.T) {
	gross := decimal.NewFromFloat(-200)
	rate := decimal.NewFromFloat(0.5)

	fee, net := CalculateFee(gross, rate)

	if !fee.Equal(decimal.Zero) {
		t.Errorf("fee should be 0 on loss, got %s", fee)
	}
	if !net.Equal(gross) {
		t.Errorf("net should equal gross on loss, got %s", net)
	}
}

func TestCalculateFeeZero(t *testing.T) {
	fee, net := CalculateFee(decimal.Zero, decimal.NewFromFloat(0.5))

	if !fee.Equal(decimal.Zero) {
		t.Errorf("fee = %s, want 0", fee)
	}
	if !net.Equal(decimal.Zero) {
		t.Errorf("net = %s, want 0", net)
	}
}
