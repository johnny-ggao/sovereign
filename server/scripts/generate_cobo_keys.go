// +build ignore

package main

import (
	"fmt"
	"os"

	coboWaas2Crypto "github.com/CoboGlobal/cobo-waas2-go-sdk/cobo_waas2/crypto"
)

func main() {
	apiKey, apiSecret, err := coboWaas2Crypto.GenerateApiKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Cobo WaaS 2.0 Ed25519 Key Pair ===")
	fmt.Println()
	fmt.Println("API Secret (保存到 .env 的 SOVEREIGN_COBO_API__SECRET):")
	fmt.Println(apiSecret)
	fmt.Println()
	fmt.Println("API Key / Public Key (粘贴到 Cobo Portal 绑定):")
	fmt.Println(apiKey)
}
