package cobo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/shopspring/decimal"

	coboWaas2 "github.com/CoboGlobal/cobo-waas2-go-sdk/cobo_waas2"
	coboWaas2Crypto "github.com/CoboGlobal/cobo-waas2-go-sdk/cobo_waas2/crypto"
)

// CoboProvider 基于 Cobo 官方 Go SDK 的钱包服务实现
type CoboProvider struct {
	client        *coboWaas2.APIClient
	ctx           context.Context
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
	if opts.APISecret == "" {
		return nil, fmt.Errorf("cobo api_secret is required")
	}

	env := coboWaas2.DevEnv
	if opts.BaseURL == "https://api.cobo.com" {
		env = coboWaas2.ProdEnv
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, coboWaas2.ContextEnv, env)
	ctx = context.WithValue(ctx, coboWaas2.ContextPortalSigner, coboWaas2Crypto.Ed25519Signer{
		Secret: opts.APISecret,
	})

	client := coboWaas2.NewAPIClient(coboWaas2.NewConfiguration())

	return &CoboProvider{
		client:        client,
		ctx:           ctx,
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

func (p *CoboProvider) GenerateAddress(ctx context.Context, req GenerateAddressReq) (*GenerateAddressResp, error) {
	chain, ok := chainIDMap[req.Network]
	if !ok {
		return nil, fmt.Errorf("unsupported network: %s", req.Network)
	}

	addresses, httpResp, err := p.client.WalletsAPI.CreateAddress(p.ctx, p.walletID).
		CreateAddressRequest(*coboWaas2.NewCreateAddressRequest(chain, 1)).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("generate address: %w (status: %s)", err, httpStatus(httpResp))
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no address returned from cobo")
	}

	return &GenerateAddressResp{
		Address:    addresses[0].GetAddress(),
		ExternalID: chain + ":" + addresses[0].GetAddress(),
	}, nil
}

// --- Withdraw ---

func (p *CoboProvider) Withdraw(ctx context.Context, req WithdrawReq) (*WithdrawResp, error) {
	tokenID := coinID(req.Currency, req.Network)

	source := coboWaas2.TransferSource{
		CustodialTransferSource: coboWaas2.NewCustodialTransferSource(
			coboWaas2.WALLETSUBTYPE_ASSET,
			p.walletID,
		),
	}

	dest := coboWaas2.NewAddressTransferDestination(coboWaas2.TRANSFERDESTINATIONTYPE_ADDRESS)
	dest.SetAccountOutput(*coboWaas2.NewAddressTransferDestinationAccountOutput(req.Address, req.Amount.String()))

	destination := coboWaas2.TransferDestination{
		AddressTransferDestination: dest,
	}

	params := *coboWaas2.NewTransferParams(req.RequestID, source, tokenID, destination)

	resp, httpResp, err := p.client.TransactionsAPI.CreateTransferTransaction(p.ctx).
		TransferParams(params).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("withdraw: %w (status: %s)", err, httpStatus(httpResp))
	}

	return &WithdrawResp{
		ExternalID: resp.GetTransactionId(),
		Status:     mapStatus(string(resp.GetStatus())),
	}, nil
}

// --- GetBalance ---

func (p *CoboProvider) GetBalance(ctx context.Context, currency string) (*BalanceResp, error) {
	tokenID := coinID(currency, "ERC20")

	resp, _, err := p.client.WalletsAPI.ListTokenBalancesForWallet(p.ctx, p.walletID).
		TokenIds(tokenID).
		Execute()
	if err != nil || !resp.HasData() || len(resp.GetData()) == 0 {
		return &BalanceResp{Currency: currency, Available: decimal.Zero, Frozen: decimal.Zero}, nil
	}

	balance := resp.GetData()[0].GetBalance()
	available, _ := decimal.NewFromString(balance.GetAvailable())
	frozen, _ := decimal.NewFromString(balance.GetFrozen())

	return &BalanceResp{Currency: currency, Available: available, Frozen: frozen}, nil
}

// --- GetTransaction ---

func (p *CoboProvider) GetTransaction(ctx context.Context, externalID string) (*TransactionResp, error) {
	resp, httpResp, err := p.client.TransactionsAPI.GetTransactionById(p.ctx, externalID).Execute()
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w (status: %s)", err, httpStatus(httpResp))
	}

	return &TransactionResp{
		ExternalID: resp.GetTransactionId(),
		TxHash:     resp.GetTransactionHash(),
		Status:     mapStatus(string(resp.GetStatus())),
		Amount:     decimal.Zero,
		Fee:        decimal.Zero,
	}, nil
}

// --- VerifyWebhook ---

func (p *CoboProvider) VerifyWebhook(signature string, payload []byte) (bool, error) {
	if p.webhookPubKey == "" {
		return true, nil
	}
	// TODO: 使用 Cobo SDK webhook 验签
	return true, nil
}

// --- 状态映射 ---

func mapStatus(coboStatus string) string {
	switch coboStatus {
	case "Submitted", "PendingApproval":
		return "pending"
	case "Queued", "PendingSignature", "Broadcasting":
		return "processing"
	case "Confirmed", "Completed":
		return "confirmed"
	case "Failed", "Rejected":
		return "failed"
	default:
		return "pending"
	}
}

func httpStatus(resp *http.Response) string {
	if resp == nil {
		return "nil"
	}
	return resp.Status
}
