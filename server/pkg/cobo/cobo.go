package cobo

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/shopspring/decimal"

	coboWaas2 "github.com/CoboGlobal/cobo-waas2-go-sdk/cobo_waas2"
	coboWaas2Crypto "github.com/CoboGlobal/cobo-waas2-go-sdk/cobo_waas2/crypto"
)

// CoboProvider 基于 Cobo 官方 Go SDK 的钱包服务实现
type CoboProvider struct {
	client            *coboWaas2.APIClient
	ctx               context.Context
	walletID          string
	webhookPubKey     string
	withdrawAddresses map[string]string
}

type Options struct {
	BaseURL           string
	APISecret         string
	APIPubKey         string
	WalletID          string
	WebhookPubKey     string
	WithdrawAddresses map[string]string // network -> source address
}

func NewCoboProvider(opts Options) (WalletProvider, error) {
	// 配置为空时从 ~/.cobo/config.toml 读取
	if opts.APISecret == "" || opts.WalletID == "" {
		fileOpts, err := loadCoboConfigFile()
		if err == nil {
			if opts.APISecret == "" {
				opts.APISecret = fileOpts.APISecret
			}
			if opts.APIPubKey == "" {
				opts.APIPubKey = fileOpts.APIPubKey
			}
			if opts.WalletID == "" {
				opts.WalletID = fileOpts.WalletID
			}
			if opts.BaseURL == "" {
				opts.BaseURL = fileOpts.BaseURL
			}
		}
	}

	if opts.APISecret == "" {
		return nil, fmt.Errorf("cobo api_secret is required (set in config or ~/.cobo/config.toml)")
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
		client:            client,
		ctx:               ctx,
		walletID:          opts.WalletID,
		webhookPubKey:     opts.WebhookPubKey,
		withdrawAddresses: opts.WithdrawAddresses,
	}, nil
}

// loadCoboConfigFile 从 ~/.cobo/config.toml 读取密钥配置
func loadCoboConfigFile() (*Options, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".cobo", "config.toml")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	opts := &Options{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "api_secret", "secret", "private_key":
			opts.APISecret = value
		case "api_key", "public_key":
			opts.APIPubKey = value
		case "wallet_id":
			opts.WalletID = value
		case "env", "environment":
			if value == "production" || value == "prod" {
				opts.BaseURL = "https://api.cobo.com"
			} else {
				opts.BaseURL = "https://api.dev.cobo.com"
			}
		}
	}

	return opts, scanner.Err()
}

// --- 币种/网络映射 ---

// chainIDMap 用于生成充值地址时的 chain_id
var chainIDMap = map[string]string{
	"ERC20": "ETH",
	"TRC20": "TRON",
	"BEP20": "BSC_BNB",
}

// tokenIDMap 用于提现/余额查询时的 token_id（MPC 钱包格式）
var tokenIDMap = map[string]map[string]string{
	"USDT": {
		"ERC20": "ETH_USDT",
		"TRC20": "TRON_USDT",
		"BEP20": "BSC_USDT",
	},
}

func coinID(currency, network string) string {
	if tokens, ok := tokenIDMap[currency]; ok {
		if tokenID, ok := tokens[network]; ok {
			return tokenID
		}
	}
	// fallback
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

	mpcSource := coboWaas2.NewMpcTransferSource(
		coboWaas2.WALLETSUBTYPE_ORG_CONTROLLED,
		p.walletID,
	)
	if addr, ok := p.withdrawAddresses[req.Network]; ok && addr != "" {
		mpcSource.SetAddress(addr)
	} else {
		return nil, fmt.Errorf("no withdraw source address configured for network %s (configured: %v)", req.Network, p.withdrawAddresses)
	}
	source := coboWaas2.TransferSource{
		MpcTransferSource: mpcSource,
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
