package service_test

import (
	"testing"

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
