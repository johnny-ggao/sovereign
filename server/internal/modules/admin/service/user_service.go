package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	authmodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	investmodel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	settlemodel "github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
	"gorm.io/gorm"
)

type UserService interface {
	List(ctx context.Context, query dto.UserListQuery) ([]dto.UserListItem, int64, error)
	Detail(ctx context.Context, userID string) (*dto.UserDetail, error)
	Update(ctx context.Context, userID string, req dto.UpdateUserRequest) error
	Disable(ctx context.Context, userID string) error
	Enable(ctx context.Context, userID string) error
	ResetPassword(ctx context.Context, userID string) (string, error)
	AdjustBalance(ctx context.Context, userID string, req dto.AdjustBalanceRequest, adminID string) error
	ListInvestments(ctx context.Context, query dto.InvestmentListQuery) ([]dto.InvestmentListItem, int64, error)
	Reset2FA(ctx context.Context, userID string) error
}

type userService struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewUserService(db *gorm.DB, logger *slog.Logger) UserService {
	return &userService{db: db, logger: logger}
}

func (s *userService) List(ctx context.Context, query dto.UserListQuery) ([]dto.UserListItem, int64, error) {
	var users []authmodel.User
	var total int64

	q := s.db.WithContext(ctx).Model(&authmodel.User{})

	if query.Search != "" {
		pattern := "%" + query.Search + "%"
		q = q.Where("email ILIKE ? OR full_name ILIKE ?", pattern, pattern)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (query.Page - 1) * query.Limit
	if err := q.Order("created_at DESC").Offset(offset).Limit(query.Limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("find users: %w", err)
	}

	items := make([]dto.UserListItem, len(users))
	for i, u := range users {
		balance := s.calcUserBalance(ctx, u.ID)
		items[i] = dto.UserListItem{
			ID:        u.ID,
			Email:     u.Email,
			FullName:  u.FullName,
			Phone:     u.Phone,
			Language:  u.Language,
			IsActive:  true,
			Balance:   balance.StringFixed(2),
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		}
	}

	return items, total, nil
}

func (s *userService) Detail(ctx context.Context, userID string) (*dto.UserDetail, error) {
	var user authmodel.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	var wallets []walletmodel.Wallet
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("find wallets: %w", err)
	}

	var transactions []walletmodel.Transaction
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Limit(20).Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("find transactions: %w", err)
	}

	var investments []investmodel.Investment
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Find(&investments).Error; err != nil {
		return nil, fmt.Errorf("find investments: %w", err)
	}

	var settlements []settlemodel.Settlement
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("settled_at DESC").Limit(20).Find(&settlements).Error; err != nil {
		return nil, fmt.Errorf("find settlements: %w", err)
	}

	walletInfos := make([]dto.WalletInfo, len(wallets))
	for i, w := range wallets {
		walletInfos[i] = dto.WalletInfo{
			Currency:    w.Currency,
			Available:   w.Available.StringFixed(2),
			InOperation: w.InOperation.StringFixed(2),
			Frozen:      w.Frozen.StringFixed(2),
			Earnings:    w.Earnings.StringFixed(2),
			Total:       w.TotalBalance().StringFixed(2),
		}
	}

	txInfos := make([]dto.TransactionInfo, len(transactions))
	for i, t := range transactions {
		txInfos[i] = dto.TransactionInfo{
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

	investInfos := make([]dto.InvestmentInfo, len(investments))
	for i, inv := range investments {
		investInfos[i] = dto.InvestmentInfo{
			ID:        inv.ID,
			Amount:    inv.Amount.StringFixed(2),
			Currency:  inv.Currency,
			Status:    inv.Status,
			NetReturn: inv.NetReturn.StringFixed(2),
			StartDate: inv.StartDate.Format("2006-01-02"),
		}
	}

	settleInfos := make([]dto.SettlementInfo, len(settlements))
	for i, st := range settlements {
		settleInfos[i] = dto.SettlementInfo{
			ID:        st.ID,
			Period:    st.Period,
			NetReturn: st.NetReturn.StringFixed(2),
			FeeRate:   st.FeeRate.StringFixed(2),
			SettledAt: st.SettledAt.Format(time.RFC3339),
		}
	}

	return &dto.UserDetail{
		ID:           user.ID,
		Email:        user.Email,
		FullName:     user.FullName,
		Phone:        user.Phone,
		Language:     user.Language,
		IsActive:     true,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		Wallets:      walletInfos,
		Transactions: txInfos,
		Investments:  investInfos,
		Settlements:  settleInfos,
	}, nil
}

func (s *userService) Update(ctx context.Context, userID string, req dto.UpdateUserRequest) error {
	updates := map[string]interface{}{}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}

	if len(updates) == 0 {
		return nil
	}

	result := s.db.WithContext(ctx).Model(&authmodel.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	s.logger.Info("user updated by admin", slog.String("user_id", userID))
	return nil
}

func (s *userService) Disable(ctx context.Context, userID string) error {
	s.logger.Info("user disable requested (no-op, is_active field not present)", slog.String("user_id", userID))
	return nil
}

func (s *userService) Enable(ctx context.Context, userID string) error {
	s.logger.Info("user enable requested (no-op, is_active field not present)", slog.String("user_id", userID))
	return nil
}

func (s *userService) ResetPassword(ctx context.Context, userID string) (string, error) {
	tempPassword := "Temp1234!"

	hash, err := crypto.HashPassword(tempPassword)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	result := s.db.WithContext(ctx).Model(&authmodel.User{}).Where("id = ?", userID).
		Update("password_hash", hash)
	if result.Error != nil {
		return "", fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return "", fmt.Errorf("user not found")
	}

	s.logger.Info("user password reset by admin", slog.String("user_id", userID))
	return tempPassword, nil
}

func (s *userService) Reset2FA(ctx context.Context, userID string) error {
	err := s.db.WithContext(ctx).Model(&authmodel.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"two_fa_secret": "", "two_fa_enabled": false}).Error
	if err != nil {
		return fmt.Errorf("reset 2fa: %w", err)
	}
	s.logger.Info("user 2fa reset", slog.String("user_id", userID))
	return nil
}

func (s *userService) AdjustBalance(ctx context.Context, userID string, req dto.AdjustBalanceRequest, adminID string) error {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	var wallet walletmodel.Wallet
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND currency = ?", userID, req.Currency).
		First(&wallet).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("find wallet: %w", err)
		}
		// 钱包不存在，自动创建
		wallet = walletmodel.Wallet{
			UserID:   userID,
			Currency: req.Currency,
		}
		if err := s.db.WithContext(ctx).Create(&wallet).Error; err != nil {
			return fmt.Errorf("create wallet: %w", err)
		}
	}

	newAvailable := wallet.Available.Add(amount)
	if newAvailable.IsNegative() {
		return fmt.Errorf("insufficient balance: available %s, adjustment %s", wallet.Available.StringFixed(2), amount.StringFixed(2))
	}

	if err := s.db.WithContext(ctx).Model(&walletmodel.Wallet{}).
		Where("id = ?", wallet.ID).
		Update("available", newAvailable).Error; err != nil {
		return fmt.Errorf("update wallet: %w", err)
	}

	s.logger.Info("balance adjusted by admin",
		slog.String("admin_id", adminID),
		slog.String("user_id", userID),
		slog.String("currency", req.Currency),
		slog.String("amount", amount.StringFixed(2)),
		slog.String("reason", req.Reason),
	)

	return nil
}

func (s *userService) ListInvestments(ctx context.Context, query dto.InvestmentListQuery) ([]dto.InvestmentListItem, int64, error) {
	type investmentRow struct {
		ID        string          `gorm:"column:id"`
		UserID    string          `gorm:"column:user_id"`
		UserEmail string          `gorm:"column:user_email"`
		Amount    decimal.Decimal `gorm:"column:amount"`
		Currency  string          `gorm:"column:currency"`
		Status    string          `gorm:"column:status"`
		NetReturn decimal.Decimal `gorm:"column:net_return"`
		StartDate time.Time       `gorm:"column:start_date"`
		CreatedAt time.Time       `gorm:"column:created_at"`
	}

	db := s.db.WithContext(ctx).Table("investments").
		Select("investments.id, investments.user_id, users.email as user_email, investments.amount, investments.currency, investments.status, investments.net_return, investments.start_date, investments.created_at").
		Joins("LEFT JOIN users ON users.id = investments.user_id")

	if query.Search != "" {
		pattern := "%" + query.Search + "%"
		db = db.Where("users.email ILIKE ? OR investments.user_id = ?", pattern, query.Search)
	}

	if query.Status != "" {
		db = db.Where("investments.status = ?", query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count investments: %w", err)
	}

	sortBy := "investments.created_at"
	if query.SortBy == "amount" {
		sortBy = "investments.amount"
	}
	sortOrder := "DESC"
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	offset := (query.Page - 1) * query.Limit
	var rows []investmentRow
	if err := db.Order(sortBy + " " + sortOrder).Offset(offset).Limit(query.Limit).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("find investments: %w", err)
	}

	items := make([]dto.InvestmentListItem, len(rows))
	for i, r := range rows {
		items[i] = dto.InvestmentListItem{
			ID:        r.ID,
			UserID:    r.UserID,
			UserEmail: r.UserEmail,
			Amount:    r.Amount.StringFixed(2),
			Currency:  r.Currency,
			Status:    r.Status,
			NetReturn: r.NetReturn.StringFixed(2),
			StartDate: r.StartDate.Format("2006-01-02"),
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		}
	}

	return items, total, nil
}

func (s *userService) calcUserBalance(ctx context.Context, userID string) decimal.Decimal {
	var wallets []walletmodel.Wallet
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets).Error; err != nil {
		return decimal.Zero
	}

	total := decimal.Zero
	for _, w := range wallets {
		total = total.Add(w.TotalBalance())
	}
	return total
}
