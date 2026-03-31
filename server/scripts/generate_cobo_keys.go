// +build ignore

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	privHex := hex.EncodeToString(priv.Seed())
	pubHex := hex.EncodeToString(pub)

	fmt.Println("=== Cobo WaaS 2.0 Ed25519 Key Pair ===")
	fmt.Println()
	fmt.Println("Private Key (保存到 .env，不要泄露):")
	fmt.Println(privHex)
	fmt.Println()
	fmt.Println("Public Key (粘贴到 Cobo 控制台绑定 API Key):")
	fmt.Println(pubHex)
	fmt.Println()
	fmt.Println("配置方式:")
	fmt.Printf("  SOVEREIGN_COBO_API__SECRET=%s\n", privHex)
}
