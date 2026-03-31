package cobo

import (
	"context"

	"github.com/shopspring/decimal"
)

// WalletProvider 抽象钱包服务商接口，当前由 Mock 实现，
// Cobo 审核通过后替换为真实实现。
type WalletProvider interface {
	// GenerateAddress 为用户生成指定币种+网络的充值地址
	GenerateAddress(ctx context.Context, req GenerateAddressReq) (*GenerateAddressResp, error)

	// Withdraw 发起提款请求
	Withdraw(ctx context.Context, req WithdrawReq) (*WithdrawResp, error)

	// GetBalance 查询托管钱包某币种余额
	GetBalance(ctx context.Context, currency string) (*BalanceResp, error)

	// GetTransaction 根据外部 ID 查询交易状态
	GetTransaction(ctx context.Context, externalID string) (*TransactionResp, error)

	// VerifyWebhook 验证 Webhook 回调签名
	VerifyWebhook(signature string, payload []byte) (bool, error)
}

type GenerateAddressReq struct {
	Currency string
	Network  string
	Label    string
}

type GenerateAddressResp struct {
	Address    string
	ExternalID string
}

type WithdrawReq struct {
	Currency  string
	Network   string
	Address   string
	Amount    decimal.Decimal
	RequestID string
}

type WithdrawResp struct {
	ExternalID string
	Status     string
}

type BalanceResp struct {
	Currency  string
	Available decimal.Decimal
	Frozen    decimal.Decimal
}

type TransactionResp struct {
	ExternalID  string
	TxHash      string
	Status      string // pending, processing, confirmed, failed
	Amount      decimal.Decimal
	Fee         decimal.Decimal
	ConfirmedAt *int64
}

// WebhookPayload Cobo Webhook 回调数据结构
type WebhookPayload struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"` // deposit, withdraw
	Status      string          `json:"status"`
	Currency    string          `json:"currency"`
	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
	TxHash      string          `json:"tx_hash"`
	Address     string          `json:"address"`
	RequestID   string          `json:"request_id"`
	ConfirmedAt *int64          `json:"confirmed_at"`
}
