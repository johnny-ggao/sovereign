package worker

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/model"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/service"
)

type stubPremiumService struct {
	ticks []model.PremiumSnapshot
}

func (s *stubPremiumService) GetLatest(_ context.Context) (interface{}, error)  { return nil, nil }
func (s *stubPremiumService) GetHistory(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, nil
}
func (s *stubPremiumService) SaveTick(_ context.Context, snap model.PremiumSnapshot) error {
	s.ticks = append(s.ticks, snap)
	return nil
}
func (s *stubPremiumService) Hub() *service.Hub { return nil }

func TestPremiumFetcherCalculation(t *testing.T) {
	kr := service.NewMockExchangeClient("upbit", 92700000, 0)
	gl := service.NewMockExchangeClient("binance", 90000000, 0)

	krOB, _ := kr.GetOrderBook(context.Background(), "BTC/KRW")
	glOB, _ := gl.GetOrderBook(context.Background(), "BTC/KRW")

	premium := krOB.Bid.Sub(glOB.Ask).Div(glOB.Ask).Mul(decimal.NewFromInt(100))

	if premium.LessThan(decimal.NewFromFloat(-5)) || premium.GreaterThan(decimal.NewFromFloat(10)) {
		t.Errorf("premium = %s%%, want between -5%% and 10%%", premium)
	}
}

func TestPremiumFetcherName(t *testing.T) {
	f := &PremiumFetcher{}
	if f.Name() != "premium_fetcher" {
		t.Errorf("Name() = %q, want premium_fetcher", f.Name())
	}
}
