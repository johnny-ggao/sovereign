package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMockExchangeClient(t *testing.T) {
	client := NewMockExchangeClient("upbit", 90000000, 2700000)

	if client.Name() != "upbit" {
		t.Errorf("Name() = %q, want upbit", client.Name())
	}

	ob, err := client.GetOrderBook(context.Background(), "BTC/KRW")
	if err != nil {
		t.Fatalf("GetOrderBook() error = %v", err)
	}

	base := decimal.NewFromFloat(92700000)
	lower := base.Mul(decimal.NewFromFloat(0.95))
	upper := base.Mul(decimal.NewFromFloat(1.05))
	if ob.Bid.LessThan(lower) || ob.Bid.GreaterThan(upper) {
		t.Errorf("bid = %s, want between %s and %s", ob.Bid, lower, upper)
	}
	if ob.Ask.LessThan(ob.Bid) {
		t.Errorf("ask %s should be >= bid %s", ob.Ask, ob.Bid)
	}
}

func TestMockExchangeClientPriceVaries(t *testing.T) {
	client := NewMockExchangeClient("upbit", 90000000, 0)
	ctx := context.Background()

	ob1, _ := client.GetOrderBook(ctx, "BTC/KRW")
	ob2, _ := client.GetOrderBook(ctx, "BTC/KRW")
	ob3, _ := client.GetOrderBook(ctx, "BTC/KRW")

	if ob1.Bid.Equal(ob2.Bid) && ob2.Bid.Equal(ob3.Bid) {
		t.Error("mock client should produce varying prices")
	}
}

func TestPremiumCalculation(t *testing.T) {
	kr := NewMockExchangeClient("upbit", 92700000, 0)
	gl := NewMockExchangeClient("binance", 90000000, 0)

	krOB, _ := kr.GetOrderBook(context.Background(), "BTC/KRW")
	glOB, _ := gl.GetOrderBook(context.Background(), "BTC/KRW")

	premium := krOB.Bid.Sub(glOB.Ask).Div(glOB.Ask).Mul(decimal.NewFromInt(100))

	if premium.LessThan(decimal.NewFromFloat(-5)) || premium.GreaterThan(decimal.NewFromFloat(10)) {
		t.Errorf("premium = %s%%, want between -5%% and 10%%", premium)
	}
}
