package cobo

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMockGenerateAddress(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.GenerateAddress(context.Background(), GenerateAddressReq{
		Currency: "USDT",
		Network:  "ERC20",
		Label:    "test",
	})
	if err != nil {
		t.Fatalf("GenerateAddress() error = %v", err)
	}
	if resp.Address == "" {
		t.Error("address should not be empty")
	}
	if resp.ExternalID == "" {
		t.Error("external ID should not be empty")
	}
}

func TestMockWithdraw(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.Withdraw(context.Background(), WithdrawReq{
		Currency:  "USDT",
		Network:   "ERC20",
		Address:   "0xabc",
		Amount:    decimal.NewFromFloat(100),
		RequestID: "req-123",
	})
	if err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if resp.Status != "pending" {
		t.Errorf("status = %q, want pending", resp.Status)
	}
}

func TestMockGetBalance(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.GetBalance(context.Background(), "USDT")
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if resp.Currency != "USDT" {
		t.Errorf("currency = %q, want USDT", resp.Currency)
	}
	if resp.Available.LessThanOrEqual(decimal.Zero) {
		t.Error("available should be > 0")
	}
}

func TestMockVerifyWebhook(t *testing.T) {
	p := NewMockProvider()
	valid, err := p.VerifyWebhook("sig", []byte("payload"))
	if err != nil {
		t.Fatalf("VerifyWebhook() error = %v", err)
	}
	if !valid {
		t.Error("mock should always return true")
	}
}
