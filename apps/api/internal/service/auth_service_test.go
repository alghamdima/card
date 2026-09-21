package service_test

import (
	"strings"
	"testing"
	"time"

	"cards-api/internal/service"
)

func TestAuthTokenLifecycle(t *testing.T) {
	authService := service.NewAuthService("test-admin-pw", "test-secret-key-123")

	// 1. Correct password
	token, err := authService.Login("test-admin-pw")
	if err != nil {
		t.Fatalf("expected successful login, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// 2. Validate token
	if !authService.ValidateToken(token) {
		t.Fatal("expected token to be valid")
	}

	// 3. Invalid password
	_, err = authService.Login("wrong-password")
	if err != service.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}

	// 4. Invalid token
	if authService.ValidateToken("invalid.token.structure") {
		t.Fatal("expected invalid token to fail validation")
	}
	if authService.ValidateToken("") {
		t.Fatal("expected empty token to fail validation")
	}
}

func TestAuthRejectsExpiredAndTamperedTokens(t *testing.T) {
	authService := service.NewAuthService("pw", "secret-one")

	if authService.ValidateToken(authService.GenerateToken(-time.Minute)) {
		t.Fatal("expired token must be rejected")
	}

	token := authService.GenerateToken(time.Hour)
	expiry, signature, _ := strings.Cut(token, ".")

	// Extending the expiry without the secret must invalidate the signature.
	if authService.ValidateToken("99999999999999." + signature) {
		t.Fatal("token with forged expiry must be rejected")
	}
	if authService.ValidateToken(expiry + ".nothex") {
		t.Fatal("token with malformed signature must be rejected")
	}

	// A token signed with another secret must not validate.
	other := service.NewAuthService("pw", "secret-two")
	if authService.ValidateToken(other.GenerateToken(time.Hour)) {
		t.Fatal("token signed with a different secret must be rejected")
	}
}

func TestLoginRejectsPasswordsOfDifferentLength(t *testing.T) {
	authService := service.NewAuthService("correct-horse", "secret")
	for _, pw := range []string{"", "c", "correct-horse-battery-staple"} {
		if _, err := authService.Login(pw); err != service.ErrInvalidCredentials {
			t.Fatalf("password %q should be rejected, got %v", pw, err)
		}
	}
}
