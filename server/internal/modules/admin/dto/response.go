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

type InvestmentListItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	NetReturn string `json:"net_return"`
	StartDate string `json:"start_date"`
	CreatedAt string `json:"created_at"`
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

type TradeListItem struct {
	ID           string `json:"id"`
	Pair         string `json:"pair"`
	BuyExchange  string `json:"buy_exchange"`
	SellExchange string `json:"sell_exchange"`
	BuyPrice     string `json:"buy_price"`
	SellPrice    string `json:"sell_price"`
	Amount       string `json:"amount"`
	PremiumPct   string `json:"premium_pct"`
	PnL          string `json:"pnl"`
	Fee          string `json:"fee"`
	ExecutedAt   string `json:"executed_at"`
}

type TradeStats struct {
	PnL1D         string `json:"pnl_1d"`
	PnL7D         string `json:"pnl_7d"`
	PnL30D        string `json:"pnl_30d"`
	UserProfit1D  string `json:"user_profit_1d"`
	UserProfit7D  string `json:"user_profit_7d"`
	UserProfit30D string `json:"user_profit_30d"`
	TradeCount1D  int64  `json:"trade_count_1d"`
	TradeCount7D  int64  `json:"trade_count_7d"`
	TradeCount30D int64  `json:"trade_count_30d"`
}

type TransactionListItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
	Type      string `json:"type"`
	Currency  string `json:"currency"`
	Network   string `json:"network"`
	Amount    string `json:"amount"`
	Fee       string `json:"fee"`
	Address   string `json:"address"`
	TxHash    string `json:"tx_hash"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type TransactionStats struct {
	Deposit1D        string `json:"deposit_1d"`
	Deposit7D        string `json:"deposit_7d"`
	Deposit30D       string `json:"deposit_30d"`
	DepositCount1D   int64  `json:"deposit_count_1d"`
	DepositCount7D   int64  `json:"deposit_count_7d"`
	DepositCount30D  int64  `json:"deposit_count_30d"`
	Withdraw1D       string `json:"withdraw_1d"`
	Withdraw7D       string `json:"withdraw_7d"`
	Withdraw30D      string `json:"withdraw_30d"`
	WithdrawCount1D  int64  `json:"withdraw_count_1d"`
	WithdrawCount7D  int64  `json:"withdraw_count_7d"`
	WithdrawCount30D int64  `json:"withdraw_count_30d"`
}
