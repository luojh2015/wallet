package pwd

import (
	"encoding/hex"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	// should be valid hex
	if _, err := hex.DecodeString(hash); err != nil {
		t.Errorf("HashPassword() returned invalid hex: %v", err)
	}
}

func TestHashPassword_DifferentSalts(t *testing.T) {
	hash1, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("first HashPassword() error = %v", err)
	}
	hash2, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("second HashPassword() error = %v", err)
	}
	if hash1 == hash2 {
		t.Error("same password should produce different hashes due to bcrypt salt")
	}
}

func TestHashPassword_EmptyString(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword('') error = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword('') returned empty hash")
	}
	if !CheckPasswordHash("", hash) {
		t.Error("empty password should verify against its own hash")
	}
}

func TestCheckPasswordHash_Matching(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !CheckPasswordHash("mypassword", hash) {
		t.Error("CheckPasswordHash() should return true for correct password")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if CheckPasswordHash("wrongpassword", hash) {
		t.Error("CheckPasswordHash() should return false for wrong password")
	}
}

func TestCheckPasswordHash_InvalidHex(t *testing.T) {
	if CheckPasswordHash("password", "not-valid-hex-!@#$") {
		t.Error("CheckPasswordHash() should return false for invalid hex")
	}
}

func TestCheckPasswordHash_CorruptedHash(t *testing.T) {
	// valid hex but not a valid bcrypt hash
	fakeHash := hex.EncodeToString([]byte("this is not bcrypt"))
	if CheckPasswordHash("password", fakeHash) {
		t.Error("CheckPasswordHash() should return false for corrupted bcrypt hash")
	}
}

func TestCheckPasswordHash_EmptyHash(t *testing.T) {
	if CheckPasswordHash("password", "") {
		t.Error("CheckPasswordHash() should return false for empty hash")
	}
}
