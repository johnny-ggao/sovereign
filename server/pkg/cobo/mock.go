package cobo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MockProvider 模拟钱包服务商，用于开发/测试阶段。
// Cobo 审核通过后替换为 CoboProvider。
type MockProvider struct{}

func NewMockProvider() WalletProvider {
	return &MockProvider{}
}

func (m *MockProvider) GenerateAddress(_ context.Context, req GenerateAddressReq) (*GenerateAddressResp, error) {
	addr := fmt.Sprintf("0xMOCK_%s_%s_%s", req.Currency, req.Network, uuid.New().String()[:8])
	return &GenerateAddressResp{
		Address:    addr,
		ExternalID: uuid.New().String(),
	}, nil
}

func (m *MockProvider) Withdraw(_ context.Context, req WithdrawReq) (*WithdrawResp, error) {
	return &WithdrawResp{
		ExternalID: uuid.New().String(),
		Status:     "pending",
	}, nil
}

func (m *MockProvider) GetBalance(_ context.Context, currency string) (*BalanceResp, error) {
	return &BalanceResp{
		Currency:  currency,
		Available: decimal.NewFromFloat(100000),
		Frozen:    decimal.Zero,
	}, nil
}

func (m *MockProvider) GetTransaction(_ context.Context, externalID string) (*TransactionResp, error) {
	return &TransactionResp{
		ExternalID: externalID,
		TxHash:     "0xMOCK_TX_" + externalID[:8],
		Status:     "confirmed",
		Amount:     decimal.NewFromFloat(1000),
		Fee:        decimal.NewFromFloat(1),
	}, nil
}

func (m *MockProvider) VerifyWebhook(_ string, _ []byte) (bool, error) {
	return true, nil
}
