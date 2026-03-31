package cobo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// CoboV2WebhookEvent Cobo WaaS 2.0 webhook 事件的完整结构
type CoboV2WebhookEvent struct {
	TransactionID string `json:"transaction_id"`
	WalletID      string `json:"wallet_id"`
	Type          string `json:"type"`   // "Deposit", "Withdrawal"
	Status        string `json:"status"` // "Confirming", "Completed", "Failed"
	CoboID        string `json:"cobo_id"`
	RequestID     *string `json:"request_id"`
	ChainID       string `json:"chain_id"` // "BSC_BNB", "TRON"
	TokenID       string `json:"token_id"` // "BSC_USDT", "TRON_USDT"
	AssetID       string `json:"asset_id"` // "USDT"
	TransactionHash string `json:"transaction_hash"`
	Source struct {
		SourceType string   `json:"source_type"`
		Addresses  []string `json:"addresses"`
	} `json:"source"`
	Destination struct {
		DestinationType string `json:"destination_type"`
		Amount          string `json:"amount"`
		Address         string `json:"address"`
	} `json:"destination"`
	Fee *struct {
		FeeAmount string `json:"fee_amount"`
	} `json:"fee"`
	CreatedTimestamp int64  `json:"created_timestamp"`
	UpdatedTimestamp int64  `json:"updated_timestamp"`
	FailedReason    *string `json:"failed_reason"`
	DataType        string `json:"data_type"`
}

// ParseWebhookPayload 将 Cobo v2 webhook JSON 转为内部 WebhookPayload
func ParseWebhookPayload(body []byte) (*WebhookPayload, error) {
	var event CoboV2WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("parse cobo webhook: %w", err)
	}

	if event.DataType != "Transaction" && event.TransactionID == "" {
		return nil, fmt.Errorf("not a transaction webhook event")
	}

	// 类型映射: Deposit -> deposit, Withdrawal -> withdraw
	txType := strings.ToLower(event.Type)
	if txType == "withdrawal" {
		txType = "withdraw"
	}

	// 状态映射
	status := mapWebhookStatus(event.Status)

	// 金额
	amount, _ := decimal.NewFromString(event.Destination.Amount)

	// 手续费
	fee := decimal.Zero
	if event.Fee != nil && event.Fee.FeeAmount != "" {
		fee, _ = decimal.NewFromString(event.Fee.FeeAmount)
	}

	// request_id
	requestID := ""
	if event.RequestID != nil {
		requestID = *event.RequestID
	}

	// 网络映射: chain_id -> network
	network := chainIDToNetwork(event.ChainID)

	return &WebhookPayload{
		ID:        event.CoboID,
		Type:      txType,
		Status:    status,
		Currency:  event.AssetID,
		Network:   network,
		Amount:    amount,
		Fee:       fee,
		TxHash:    event.TransactionHash,
		Address:   event.Destination.Address,
		RequestID: requestID,
	}, nil
}

func mapWebhookStatus(coboStatus string) string {
	switch coboStatus {
	case "Submitted", "PendingScreening", "PendingApproval":
		return "pending"
	case "Confirming", "Broadcasting", "InPool", "Signing":
		return "processing"
	case "Completed":
		return "confirmed"
	case "Failed", "Rejected":
		return "failed"
	default:
		return "pending"
	}
}

func chainIDToNetwork(chainID string) string {
	switch chainID {
	case "ETH":
		return "ERC20"
	case "TRON":
		return "TRC20"
	case "BSC_BNB":
		return "BEP20"
	default:
		return chainID
	}
}
