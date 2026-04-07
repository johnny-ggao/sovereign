package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	authmodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	investmodel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"gorm.io/gorm"
)

type DashboardService interface {
	Stats(ctx context.Context) (*dto.DashboardStats, error)
}

type dashboardService struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewDashboardService(db *gorm.DB, logger *slog.Logger) DashboardService {
	return &dashboardService{db: db, logger: logger}
}

func (s *dashboardService) Stats(ctx context.Context) (*dto.DashboardStats, error) {
	var totalUsers int64
	if err := s.db.WithContext(ctx).Model(&authmodel.User{}).Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	todayMidnight := time.Now().Truncate(24 * time.Hour)
	var newUsersToday int64
	if err := s.db.WithContext(ctx).Model(&authmodel.User{}).
		Where("created_at >= ?", todayMidnight).Count(&newUsersToday).Error; err != nil {
		return nil, fmt.Errorf("count new users: %w", err)
	}

	var totalInvested decimal.Decimal
	if err := s.db.WithContext(ctx).Model(&investmodel.Investment{}).
		Where("status = ?", investmodel.InvestStatusActive).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalInvested).Error; err != nil {
		return nil, fmt.Errorf("sum investments: %w", err)
	}

	var totalDeposits decimal.Decimal
	if err := s.db.WithContext(ctx).Model(&walletmodel.Transaction{}).
		Where("type = ? AND status = ?", walletmodel.TxTypeDeposit, walletmodel.TxStatusConfirmed).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalDeposits).Error; err != nil {
		return nil, fmt.Errorf("sum deposits: %w", err)
	}

	var totalWithdrawals decimal.Decimal
	if err := s.db.WithContext(ctx).Model(&walletmodel.Transaction{}).
		Where("type = ? AND status = ?", walletmodel.TxTypeWithdraw, walletmodel.TxStatusConfirmed).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalWithdrawals).Error; err != nil {
		return nil, fmt.Errorf("sum withdrawals: %w", err)
	}

	var activeInvestments int64
	if err := s.db.WithContext(ctx).Model(&investmodel.Investment{}).
		Where("status = ?", investmodel.InvestStatusActive).Count(&activeInvestments).Error; err != nil {
		return nil, fmt.Errorf("count investments: %w", err)
	}

	userTrend, err := s.userTrend(ctx)
	if err != nil {
		return nil, fmt.Errorf("user trend: %w", err)
	}

	recentTxs, err := s.recentTransactions(ctx)
	if err != nil {
		return nil, fmt.Errorf("recent transactions: %w", err)
	}

	return &dto.DashboardStats{
		TotalUsers:         totalUsers,
		NewUsersToday:      newUsersToday,
		TotalInvested:      totalInvested.StringFixed(2),
		TotalDeposits:      totalDeposits.StringFixed(2),
		TotalWithdrawals:   totalWithdrawals.StringFixed(2),
		ActiveInvestments:  activeInvestments,
		UserTrend:          userTrend,
		RecentTransactions: recentTxs,
	}, nil
}

func (s *dashboardService) userTrend(ctx context.Context) ([]dto.UserTrendItem, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Truncate(24 * time.Hour)

	type dayCount struct {
		Date  time.Time `gorm:"column:date"`
		Count int64     `gorm:"column:count"`
	}

	var rows []dayCount
	if err := s.db.WithContext(ctx).Model(&authmodel.User{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("query user trend: %w", err)
	}

	countMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		countMap[r.Date.Format("2006-01-02")] = r.Count
	}

	trend := make([]dto.UserTrendItem, 0, 30)
	for i := 29; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		trend = append(trend, dto.UserTrendItem{
			Date:  day,
			Count: countMap[day],
		})
	}

	return trend, nil
}

func (s *dashboardService) recentTransactions(ctx context.Context) ([]dto.TransactionInfo, error) {
	var transactions []walletmodel.Transaction
	if err := s.db.WithContext(ctx).Model(&walletmodel.Transaction{}).
		Order("created_at DESC").Limit(10).Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find recent transactions: %w", err)
	}

	infos := make([]dto.TransactionInfo, len(transactions))
	for i, t := range transactions {
		infos[i] = dto.TransactionInfo{
			ID:        t.ID,
			Type:      t.Type,
			Currency:  t.Currency,
			Network:   t.Network,
			Amount:    t.Amount.StringFixed(2),
			Status:    t.Status,
			TxHash:    t.TxHash,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		}
	}

	return infos, nil
}
