package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	investRepo "github.com/sovereign-fund/sovereign/internal/modules/investment/repository"
	settlModel "github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	settlRepo "github.com/sovereign-fund/sovereign/internal/modules/settlement/repository"
	tradeRepo "github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	walletRepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"gorm.io/gorm"
)

type SettlementJob struct {
	invRepo    investRepo.InvestmentRepository
	tradeRepo  tradeRepo.TradeRepository
	settlRepo  settlRepo.SettlementRepository
	walletRepo walletRepo.WalletRepository
	eventBus   events.Bus
	logger     *slog.Logger
	feeRate    decimal.Decimal
}

func NewSettlementJob(
	ir investRepo.InvestmentRepository,
	tr tradeRepo.TradeRepository,
	sr settlRepo.SettlementRepository,
	wr walletRepo.WalletRepository,
	bus events.Bus,
	logger *slog.Logger,
) *SettlementJob {
	return &SettlementJob{
		invRepo:    ir,
		tradeRepo:  tr,
		settlRepo:  sr,
		walletRepo: wr,
		eventBus:   bus,
		logger:     logger,
		feeRate:    decimal.NewFromFloat(0.5),
	}
}

// NewSettlementJobFromDB 便捷构造函数，自动创建所有 repo
func NewSettlementJobFromDB(db *gorm.DB, bus events.Bus, logger *slog.Logger) *SettlementJob {
	return NewSettlementJob(
		investRepo.NewInvestmentRepository(db),
		tradeRepo.NewTradeRepository(db),
		settlRepo.NewSettlementRepository(db),
		walletRepo.NewWalletRepository(db),
		bus,
		logger,
	)
}

func (j *SettlementJob) Name() string {
	return "settlement_job"
}

// Run 定时任务调用，结算昨天的交易
func (j *SettlementJob) Run(ctx context.Context) error {
	return j.RunForDate(ctx, time.Now().AddDate(0, 0, -1))
}

// RunToday 手动触发，结算今天的交易
func (j *SettlementJob) RunToday(ctx context.Context) error {
	return j.RunForDate(ctx, time.Now())
}

// RunForDate 结算指定日期的交易盈利
func (j *SettlementJob) RunForDate(ctx context.Context, date time.Time) error {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)
	period := date.Format("2006-01-02")

	j.logger.Info("settlement started", slog.String("period", period))

	// 1. 获取所有 active 投资
	activeInvs, err := j.invRepo.FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("find active investments: %w", err)
	}

	if len(activeInvs) == 0 {
		j.logger.Info("no active investments, skip settlement")
		return nil
	}

	// 2. 统计当天所有基金级交易的总盈利
	summary, err := j.tradeRepo.SummarizeByPeriod(ctx, dayStart, dayEnd)
	if err != nil {
		return fmt.Errorf("summarize trades: %w", err)
	}

	totalPnL := decimal.NewFromFloat(summary.TotalPnL)
	totalTradeCount := summary.TotalTrades
	avgPremium := decimal.NewFromFloat(summary.AvgPremium)

	j.logger.Info("daily pnl calculated",
		slog.String("period", period),
		slog.String("total_pnl", totalPnL.String()),
		slog.Int64("total_trades", totalTradeCount),
	)

	// 如果当天没有盈利，跳过分配
	if totalPnL.LessThanOrEqual(decimal.Zero) {
		j.logger.Info("no profit to distribute", slog.String("period", period))
		return nil
	}

	// 3. 盈利的 50% 分给用户
	userShare := totalPnL.Mul(j.feeRate) // 50% 给用户
	platformFee := totalPnL.Sub(userShare)

	// 4. 计算总投资额（用于按比例分配）
	totalInvested := decimal.Zero
	for _, inv := range activeInvs {
		totalInvested = totalInvested.Add(inv.Amount)
	}

	if totalInvested.IsZero() {
		return nil
	}

	// 5. 按投资金额比例分配给每个投资
	for _, inv := range activeInvs {
		ratio := inv.Amount.Div(totalInvested)
		invShare := userShare.Mul(ratio).Round(18)
		invFee := platformFee.Mul(ratio).Round(18)
		invGross := totalPnL.Mul(ratio).Round(18)

		tradeCount := totalTradeCount

		// 检查是否已结算
		existing, _ := j.settlRepo.FindByInvestmentAndPeriod(ctx, inv.ID, period)
		if existing != nil {
			continue
		}

		// 创建结算记录
		settlement := &settlModel.Settlement{
			InvestmentID:   inv.ID,
			UserID:         inv.UserID,
			Period:         period,
			GrossReturn:    invGross,
			PerformanceFee: invFee,
			FeeRate:        j.feeRate,
			NetReturn:      invShare,
			TradeCount:     int(tradeCount),
			AvgPremiumPct:  avgPremium,
			SettledAt:      time.Now(),
		}

		if err := j.settlRepo.Create(ctx, settlement); err != nil {
			j.logger.Error("create settlement failed",
				slog.String("investment_id", inv.ID),
				slog.String("error", err.Error()),
			)
			continue
		}

		// 更新投资收益
		updated := inv
		updated.TotalReturn = updated.TotalReturn.Add(invGross)
		updated.PerformanceFee = updated.PerformanceFee.Add(invFee)
		updated.NetReturn = updated.NetReturn.Add(invShare)
		if err := j.invRepo.Update(ctx, &updated); err != nil {
			j.logger.Error("update investment failed", slog.String("error", err.Error()))
			continue
		}

		// 更新钱包余额：将净收益加到 Available
		wallet, err := j.walletRepo.FindByUserIDAndCurrency(ctx, inv.UserID, inv.Currency)
		if err != nil {
			j.logger.Error("find wallet failed", slog.String("error", err.Error()))
			continue
		}
		newAvailable := wallet.Available.Add(invShare)
		if err := j.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, wallet.Frozen); err != nil {
			j.logger.Error("update wallet failed", slog.String("error", err.Error()))
			continue
		}

		j.eventBus.Publish(ctx, events.Event{
			Type: events.SettlementCreated,
			Payload: map[string]string{
				"user_id":       inv.UserID,
				"settlement_id": settlement.ID,
				"period":        period,
				"net_return":    invShare.String(),
			},
		})

		j.logger.Info("settlement distributed",
			slog.String("investment_id", inv.ID),
			slog.String("user_id", inv.UserID),
			slog.String("gross", invGross.String()),
			slog.String("fee", invFee.String()),
			slog.String("net_to_user", invShare.String()),
			slog.String("ratio", ratio.String()),
		)
	}

	j.logger.Info("daily settlement completed",
		slog.String("period", period),
		slog.String("total_pnl", totalPnL.String()),
		slog.String("user_share", userShare.String()),
		slog.String("platform_fee", platformFee.String()),
	)

	return nil
}
