package security

import (
	"strings"
	"testing"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("safe-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "safe-password" {
		t.Fatal("password was stored as plaintext")
	}
	if !VerifyPassword(hash, "safe-password") {
		t.Fatal("valid password was rejected")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("invalid password was accepted")
	}
	if NeedsUpgrade(hash) {
		t.Fatal("new hash should not need an upgrade")
	}
}

func TestLegacyPlaintextPassword(t *testing.T) {
	if !VerifyPassword("123456", "123456") {
		t.Fatal("legacy plaintext password was rejected")
	}
	if VerifyPassword("123456", "654321") {
		t.Fatal("invalid legacy plaintext password was accepted")
	}
	if !NeedsUpgrade("123456") {
		t.Fatal("legacy plaintext password should need an upgrade")
	}
}

func TestLongLegacyPasswordRemainsVerifiable(t *testing.T) {
	password := strings.Repeat("密", 25)
	if !VerifyPassword(password, password) {
		t.Fatal("long legacy plaintext password was rejected")
	}
	if CanHashPassword(password) {
		t.Fatal("password over bcrypt's 72-byte limit should not be upgraded")
	}
}
