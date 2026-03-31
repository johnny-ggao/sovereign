package crypto

import (
	"encoding/hex"
	"testing"
)

func TestAESEncryptDecrypt(t *testing.T) {
	key := hex.EncodeToString([]byte("12345678901234567890123456789012"))

	enc, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("NewAESEncryptor() error = %v", err)
	}

	plaintext := "sensitive wallet address"

	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if ciphertext == plaintext {
		t.Fatal("ciphertext should not equal plaintext")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestAESInvalidKey(t *testing.T) {
	_, err := NewAESEncryptor("tooshort")
	if err == nil {
		t.Fatal("expected error for short key")
	}
}
