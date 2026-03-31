package dto

import "github.com/shopspring/decimal"

type WalletResponse struct {
	ID          string          `json:"id"`
	Currency    string          `json:"currency"`
	Available   decimal.Decimal `json:"available"`
	InOperation decimal.Decimal `json:"in_operation"`
	Frozen      decimal.Decimal `json:"frozen"`
	Total       decimal.Decimal `json:"total"`
}

type WalletOverview struct {
	Wallets    []WalletResponse `json:"wallets"`
	TotalUSDT  decimal.Decimal  `json:"total_usdt"`
}

type DepositAddressResponse struct {
	Currency string `json:"currency"`
	Network  string `json:"network"`
	Address  string `json:"address"`
}

type WithdrawResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type WhitelistAddressResponse struct {
	ID            string `json:"id"`
	Currency      string `json:"currency"`
	Network       string `json:"network"`
	Address       string `json:"address"`
	Label         string `json:"label"`
	CooldownUntil string `json:"cooldown_until"`
	IsActive      bool   `json:"is_active"`
}

type TransactionResponse struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Currency    string          `json:"currency"`
	Network     string          `json:"network"`
	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
	Address     string          `json:"address"`
	TxHash      string          `json:"tx_hash"`
	Status      string          `json:"status"`
	ConfirmedAt *string         `json:"confirmed_at"`
	CreatedAt   string          `json:"created_at"`
}
