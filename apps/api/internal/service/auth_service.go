package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
	ErrUnauthorized       = errors.New("UNAUTHORIZED")
	ErrTokenExpired       = errors.New("TOKEN_EXPIRED")
)

type AuthService struct {
	adminPassword string
	sessionSecret string
}

func NewAuthService(adminPassword, sessionSecret string) *AuthService {
	return &AuthService{
		adminPassword: adminPassword,
		sessionSecret: sessionSecret,
	}
}

func (s *AuthService) Login(password string) (string, error) {
	if subtle.ConstantTimeCompare([]byte(password), []byte(s.adminPassword)) != 1 {
		time.Sleep(300 * time.Millisecond) // Thwart timing attacks
		return "", ErrInvalidCredentials
	}

	token := s.GenerateToken(24 * time.Hour)
	return token, nil
}

func (s *AuthService) GenerateToken(duration time.Duration) string {
	exp := time.Now().Add(duration).UnixMilli()
	expStr := strconv.FormatInt(exp, 10)

	h := hmac.New(sha256.New, []byte(s.sessionSecret))
	h.Write([]byte(expStr))
	sig := hex.EncodeToString(h.Sum(nil))

	return expStr + "." + sig
}

func (s *AuthService) ValidateToken(token string) bool {
	if token == "" {
		return false
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().UnixMilli() > exp {
		return false
	}

	h := hmac.New(sha256.New, []byte(s.sessionSecret))
	h.Write([]byte(parts[0]))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedSig)) == 1
}
