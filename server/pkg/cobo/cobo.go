package cobo

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/shopspring/decimal"
)

// CoboProvider 真实 Cobo WaaS 2.0 集成
type CoboProvider struct {
	client        *httpClient
	walletID      string
	webhookPubKey string
}

type Options struct {
	BaseURL       string
	APISecret     string
	APIPubKey     string
	WalletID      string
	WebhookPubKey string
}

func NewCoboProvider(opts Options) (WalletProvider, error) {
	s, err := newSigner(opts.APISecret, opts.APIPubKey)
	if err != nil {
		return nil, fmt.Errorf("init cobo signer: %w", err)
	}

	return &CoboProvider{
		client:        newHTTPClient(opts.BaseURL, s),
		walletID:      opts.WalletID,
		webhookPubKey: opts.WebhookPubKey,
	}, nil
}

// --- 币种/网络映射 ---

var chainIDMap = map[string]string{
	"ERC20": "ETH",
	"TRC20": "TRON",
	"BEP20": "BSC_BNB",
}

func coinID(currency, network string) string {
	chain, ok := chainIDMap[network]
	if !ok {
		chain = currency
	}
	return chain + "_" + currency
}

// --- GenerateAddress ---

type generateAddressReq struct {
	ChainID string `json:"chain_id"`
	Count   int    `json:"count"`
}

type addressItem struct {
	Address string `json:"address"`
	ChainID string `json:"chain_id"`
	Memo    string `json:"memo"`
	Path    string `json:"path"`
}

func (p *CoboProvider) GenerateAddress(ctx context.Context, req GenerateAddressReq) (*GenerateAddressResp, error) {
	chain, ok := chainIDMap[req.Network]
	if !ok {
		return nil, fmt.Errorf("unsupported network: %s", req.Network)
	}

	path := fmt.Sprintf("/v2/wallets/%s/addresses", p.walletID)
	body := generateAddressReq{ChainID: chain, Count: 1}

	data, err := p.client.post(ctx, path, body)
	if err != nil {
		return nil, fmt.Errorf("generate address: %w", err)
	}

	var items []addressItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parse address response: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no address returned from cobo")
	}

	return &GenerateAddressResp{
		Address:    items[0].Address,
		ExternalID: chain + ":" + items[0].Address,
	}, nil
}

// --- Withdraw ---

type withdrawAPIReq struct {
	CoinID    string `json:"coin_id"`
	Address   string `json:"address"`
	Amount    string `json:"amount"`
	RequestID string `json:"request_id"`
}

type withdrawAPIResp struct {
	CoboID string `json:"cobo_id"`
	Status string `json:"status"`
}

func (p *CoboProvider) Withdraw(ctx context.Context, req WithdrawReq) (*WithdrawResp, error) {
	body := withdrawAPIReq{
		CoinID:    coinID(req.Currency, req.Network),
		Address:   req.Address,
		Amount:    req.Amount.String(),
		RequestID: req.RequestID,
	}

	data, err := p.client.post(ctx, "/v2/transactions/withdraw", body)
	if err != nil {
		return nil, fmt.Errorf("withdraw: %w", err)
	}

	var resp withdrawAPIResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse withdraw response: %w", err)
	}

	return &WithdrawResp{
		ExternalID: resp.CoboID,
		Status:     mapStatus(resp.Status),
	}, nil
}

// --- GetBalance ---

type balanceAPIResp struct {
	TokenID string `json:"token_id"`
	Balance struct {
		Available string `json:"available"`
		Frozen    string `json:"frozen"`
	} `json:"balance"`
}

func (p *CoboProvider) GetBalance(ctx context.Context, currency string) (*BalanceResp, error) {
	tokenID := coinID(currency, "ERC20")
	path := fmt.Sprintf("/v2/wallets/%s/tokens/%s", p.walletID, tokenID)

	data, err := p.client.get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	var resp balanceAPIResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse balance response: %w", err)
	}

	available, _ := decimal.NewFromString(resp.Balance.Available)
	frozen, _ := decimal.NewFromString(resp.Balance.Frozen)

	return &BalanceResp{
		Currency:  currency,
		Available: available,
		Frozen:    frozen,
	}, nil
}

// --- GetTransaction ---

type transactionAPIResp struct {
	CoboID      string `json:"cobo_id"`
	RequestID   string `json:"request_id"`
	TxHash      string `json:"tx_hash"`
	Status      string `json:"status"`
	Amount      string `json:"amount"`
	Fee         string `json:"fee"`
	ConfirmedAt int64  `json:"confirmed_at"`
}

func (p *CoboProvider) GetTransaction(ctx context.Context, externalID string) (*TransactionResp, error) {
	params := url.Values{"request_id": {externalID}}

	data, err := p.client.get(ctx, "/v2/transactions", params)
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}

	var resp transactionAPIResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse transaction response: %w", err)
	}

	amount, _ := decimal.NewFromString(resp.Amount)
	fee, _ := decimal.NewFromString(resp.Fee)

	var confirmedAt *int64
	if resp.ConfirmedAt > 0 {
		confirmedAt = &resp.ConfirmedAt
	}

	return &TransactionResp{
		ExternalID:  resp.CoboID,
		TxHash:      resp.TxHash,
		Status:      mapStatus(resp.Status),
		Amount:      amount,
		Fee:         fee,
		ConfirmedAt: confirmedAt,
	}, nil
}

// --- VerifyWebhook ---

func (p *CoboProvider) VerifyWebhook(signature string, payload []byte) (bool, error) {
	if p.webhookPubKey == "" {
		// 未配置 webhook 公钥时跳过验证（开发环境）
		return true, nil
	}

	pubBytes, err := hex.DecodeString(p.webhookPubKey)
	if err != nil {
		return false, fmt.Errorf("decode webhook public key: %w", err)
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("decode webhook signature: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid webhook public key size")
	}

	return ed25519.Verify(ed25519.PublicKey(pubBytes), payload, sigBytes), nil
}

// --- 状态映射 ---

// RawGet 暴露底层 GET 请求（用于调试/查询）
func RawGet(p WalletProvider, ctx context.Context, path string, params url.Values) ([]byte, error) {
	cp, ok := p.(*CoboProvider)
	if !ok {
		return nil, fmt.Errorf("not a CoboProvider")
	}
	return cp.client.get(ctx, path, params)
}

func mapStatus(coboStatus string) string {
	switch coboStatus {
	case "submitted", "pending_approval":
		return "pending"
	case "queued", "pending_signature", "broadcasting":
		return "processing"
	case "confirmed", "completed":
		return "confirmed"
	case "failed", "rejected":
		return "failed"
	default:
		return "pending"
	}
}
