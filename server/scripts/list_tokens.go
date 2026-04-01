// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	coboWaas2 "github.com/CoboGlobal/cobo-waas2-go-sdk/cobo_waas2"
	coboWaas2Crypto "github.com/CoboGlobal/cobo-waas2-go-sdk/cobo_waas2/crypto"
	"github.com/sovereign-fund/sovereign/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, coboWaas2.ContextEnv, coboWaas2.ProdEnv)
	ctx = context.WithValue(ctx, coboWaas2.ContextPortalSigner, coboWaas2Crypto.Ed25519Signer{
		Secret: cfg.Cobo.APISecret,
	})

	client := coboWaas2.NewAPIClient(coboWaas2.NewConfiguration())

	// 查询钱包支持的 token
	resp, _, err := client.WalletsAPI.ListTokenBalancesForWallet(ctx, cfg.Cobo.WalletID).Execute()
	if err != nil {
		fmt.Printf("ListTokenBalances error: %v\n", err)
	} else if resp.HasData() {
		fmt.Println("=== Wallet Token Balances ===")
		for _, t := range resp.GetData() {
			fmt.Printf("  token_id=%-20s balance=%s\n", t.GetTokenId(), t.GetBalance().GetAvailable())
		}
	}

	// 查询支持的 token 列表
	fmt.Println("\n=== Enabled Tokens (USDT related) ===")
	tokResp, _, err := client.WalletsAPI.ListEnabledTokens(ctx).WalletId(cfg.Cobo.WalletID).Execute()
	if err != nil {
		fmt.Printf("ListEnabledTokens error: %v\n", err)
	} else if tokResp.HasData() {
		for _, t := range tokResp.GetData() {
			id := t.GetTokenId()
			if len(id) > 3 && (id[len(id)-4:] == "USDT" || id[:4] == "USDT") {
				fmt.Printf("  token_id=%s\n", id)
			}
		}
	}
}
