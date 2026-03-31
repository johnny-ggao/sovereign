package crypto

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !valid {
		t.Fatal("VerifyPassword() returned false for correct password")
	}

	valid, err = VerifyPassword("wrongPassword", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if valid {
		t.Fatal("VerifyPassword() returned true for wrong password")
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	password := "testPassword123!"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Fatal("two hashes of the same password should not be equal (unique salts)")
	}
}
