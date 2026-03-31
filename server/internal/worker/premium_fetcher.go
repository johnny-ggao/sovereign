package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/model"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/service"
)

type PremiumFetcher struct {
	premiumSvc   service.PremiumService
	krClients    []service.ExchangeClient
	glClients    []service.ExchangeClient
	rateProvider *service.ScolkgRateProvider
	pairs        []string
	logger       *slog.Logger
}

func NewPremiumFetcher(
	premiumSvc service.PremiumService,
	krClients []service.ExchangeClient,
	glClients []service.ExchangeClient,
	rateProvider *service.ScolkgRateProvider,
	logger *slog.Logger,
) *PremiumFetcher {
	return &PremiumFetcher{
		premiumSvc:   premiumSvc,
		krClients:    krClients,
		glClients:    glClients,
		rateProvider: rateProvider,
		pairs:        []string{model.PairBTCKRW, model.PairETHKRW, model.PairSOLKRW, model.PairXRPKRW},
		logger:       logger,
	}
}

func (f *PremiumFetcher) Name() string {
	return "premium_fetcher"
}

func (f *PremiumFetcher) Run(ctx context.Context) error {
	usdtKRW, err := f.getUSDTKRWRate(ctx)
	if err != nil {
		f.logger.Error("failed to fetch USDT/KRW rate", slog.String("error", err.Error()))
		return nil
	}

	for _, pair := range f.pairs {
		if err := f.fetchPair(ctx, pair, usdtKRW); err != nil {
			f.logger.Error("fetch pair failed",
				slog.String("pair", pair),
				slog.String("error", err.Error()),
			)
		}
	}
	return nil
}

// AggregatedOB 聚合多交易所的最优价格
type AggregatedOB struct {
	BestBid   decimal.Decimal
	BestAsk   decimal.Decimal
	BidSource string
	AskSource string
}

func (f *PremiumFetcher) aggregateOrderBooks(ctx context.Context, clients []service.ExchangeClient, pair string) (*AggregatedOB, error) {
	var agg AggregatedOB
	found := false

	for _, client := range clients {
		ob, err := client.GetOrderBook(ctx, pair)
		if err != nil {
			continue
		}

		if !found || ob.Bid.GreaterThan(agg.BestBid) {
			agg.BestBid = ob.Bid
			agg.BidSource = client.Name()
		}
		if !found || ob.Ask.LessThan(agg.BestAsk) {
			agg.BestAsk = ob.Ask
			agg.AskSource = client.Name()
		}
		found = true
	}

	if !found {
		return nil, errAllClientsFailed
	}

	return &agg, nil
}

var errAllClientsFailed = &clientError{msg: "all exchange clients failed for this pair"}

type clientError struct{ msg string }

func (e *clientError) Error() string { return e.msg }

func (f *PremiumFetcher) getUSDTKRWRate(_ context.Context) (decimal.Decimal, error) {
	rate := f.rateProvider.Rate()
	if rate.IsZero() {
		return decimal.Zero, fmt.Errorf("scolkg rate not available")
	}
	return rate, nil
}

func (f *PremiumFetcher) fetchPair(ctx context.Context, pair string, usdtKRW decimal.Decimal) error {
	krAgg, err := f.aggregateOrderBooks(ctx, f.krClients, pair)
	if err != nil {
		return err
	}

	glAgg, err := f.aggregateOrderBooks(ctx, f.glClients, pair)
	if err != nil {
		return err
	}

	hundred := decimal.NewFromInt(100)

	// 正向溢价：全球买入(gl best ask) → 韩国卖出(kr best bid)
	glAskKRW := glAgg.BestAsk.Mul(usdtKRW)
	forwardPct := decimal.Zero
	if glAskKRW.GreaterThan(decimal.Zero) {
		forwardPct = krAgg.BestBid.Sub(glAskKRW).Div(glAskKRW).Mul(hundred)
	}

	// 反向溢价：韩国买入(kr best ask) → 全球卖出(gl best bid)
	glBidKRW := glAgg.BestBid.Mul(usdtKRW)
	reversePct := decimal.Zero
	if krAgg.BestAsk.GreaterThan(decimal.Zero) {
		reversePct = glBidKRW.Sub(krAgg.BestAsk).Div(krAgg.BestAsk).Mul(hundred)
	}

	// 收集各交易所延迟
	latencies := make(map[string]int64)
	for _, c := range f.krClients {
		if l := c.Latency(); l > 0 {
			latencies[c.Name()] = l.Milliseconds()
		}
	}
	for _, c := range f.glClients {
		if l := c.Latency(); l > 0 {
			latencies[c.Name()] = l.Milliseconds()
		}
	}

	snapshot := model.PremiumSnapshot{
		Pair:              pair,
		KoreanPrice:       krAgg.BestBid,
		GlobalPrice:       glAskKRW,
		PremiumPct:        forwardPct.Round(4),
		ReversePremiumPct: reversePct.Round(4),
		SourceKR:          krAgg.BidSource,
		SourceGL:          glAgg.AskSource,
		Latencies:         latencies,
		Timestamp:         time.Now(),
	}

	return f.premiumSvc.SaveTick(ctx, snapshot)
}
