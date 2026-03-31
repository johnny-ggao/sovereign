package cobo

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

type signer struct {
	privKey ed25519.PrivateKey
	pubHex  string
}

func newSigner(privKeyHex, pubKeyHex string) (*signer, error) {
	seed, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid private key length: got %d, want %d", len(seed), ed25519.SeedSize)
	}

	return &signer{
		privKey: ed25519.NewKeyFromSeed(seed),
		pubHex:  pubKeyHex,
	}, nil
}

// sign 生成 Cobo WaaS 2.0 请求签名。
// str_to_sign = METHOD|PATH|TIMESTAMP|PARAMS|BODY
// 然后 double SHA-256，再 Ed25519 签名。
func (s *signer) sign(method, path, params, body string) (nonce string, signature string) {
	nonce = strconv.FormatInt(time.Now().UnixMilli(), 10)

	strToSign := method + "|" + path + "|" + nonce + "|" + params + "|" + body

	// Double SHA-256
	first := sha256.Sum256([]byte(strToSign))
	second := sha256.Sum256(first[:])

	sig := ed25519.Sign(s.privKey, second[:])

	return nonce, hex.EncodeToString(sig)
}
