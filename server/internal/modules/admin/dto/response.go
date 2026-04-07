package dto

import "time"

type LoginResponse struct {
	Token     string        `json:"token"`
	ExpiresAt int64         `json:"expires_at"`
	Admin     AdminResponse `json:"admin"`
}

type AdminResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
}

type UserListItem struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Language  string `json:"language"`
	IsActive  bool   `json:"is_active"`
	Balance   string `json:"balance"`
	CreatedAt string `json:"created_at"`
}

type UserDetail struct {
	ID           string            `json:"id"`
	Email        string            `json:"email"`
	FullName     string            `json:"full_name"`
	Phone        string            `json:"phone"`
	Language     string            `json:"language"`
	IsActive     bool              `json:"is_active"`
	CreatedAt    string            `json:"created_at"`
	Wallets      []WalletInfo      `json:"wallets"`
	Transactions []TransactionInfo `json:"recent_transactions"`
	Investments  []InvestmentInfo  `json:"investments"`
	Settlements  []SettlementInfo  `json:"recent_settlements"`
}

type WalletInfo struct {
	Currency    string `json:"currency"`
	Available   string `json:"available"`
	InOperation string `json:"in_operation"`
	Frozen      string `json:"frozen"`
	Earnings    string `json:"earnings"`
	Total       string `json:"total"`
}

type TransactionInfo struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Currency  string `json:"currency"`
	Network   string `json:"network"`
	Amount    string `json:"amount"`
	Status    string `json:"status"`
	TxHash    string `json:"tx_hash"`
	CreatedAt string `json:"created_at"`
}

type InvestmentInfo struct {
	ID        string `json:"id"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	NetReturn string `json:"net_return"`
	StartDate string `json:"start_date"`
}

type SettlementInfo struct {
	ID        string `json:"id"`
	Period    string `json:"period"`
	NetReturn string `json:"net_return"`
	FeeRate   string `json:"fee_rate"`
	SettledAt string `json:"settled_at"`
}

type DashboardStats struct {
	TotalUsers         int64             `json:"total_users"`
	NewUsersToday      int64             `json:"new_users_today"`
	TotalInvested      string            `json:"total_invested"`
	TotalDeposits      string            `json:"total_deposits"`
	TotalWithdrawals   string            `json:"total_withdrawals"`
	ActiveInvestments  int64             `json:"active_investments"`
	UserTrend          []UserTrendItem   `json:"user_trend"`
	RecentTransactions []TransactionInfo `json:"recent_transactions"`
}

type UserTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
