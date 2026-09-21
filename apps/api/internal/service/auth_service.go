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

var ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")

const (
	tokenLifetime      = 24 * time.Hour
	failedLoginDelay   = 300 * time.Millisecond
	tokenPartsExpected = 2
)

type AuthService struct {
	passwordDigest [sha256.Size]byte
	sessionSecret  []byte
	now            func() time.Time
}

func NewAuthService(adminPassword, sessionSecret string) *AuthService {
	return &AuthService{
		// Comparing fixed-length digests keeps the check constant-time even
		// when the attacker-supplied password has a different length.
		passwordDigest: sha256.Sum256([]byte(adminPassword)),
		sessionSecret:  []byte(sessionSecret),
		now:            time.Now,
	}
}

func (s *AuthService) Login(password string) (string, error) {
	digest := sha256.Sum256([]byte(password))
	if subtle.ConstantTimeCompare(digest[:], s.passwordDigest[:]) != 1 {
		time.Sleep(failedLoginDelay) // slow down online guessing
		return "", ErrInvalidCredentials
	}
	return s.GenerateToken(tokenLifetime), nil
}

// GenerateToken returns "<expiryUnixMilli>.<hex hmac-sha256 of expiry>".
func (s *AuthService) GenerateToken(duration time.Duration) string {
	expiry := strconv.FormatInt(s.now().Add(duration).UnixMilli(), 10)
	return expiry + "." + s.sign(expiry)
}

func (s *AuthService) ValidateToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != tokenPartsExpected {
		return false
	}

	expiry, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || s.now().UnixMilli() > expiry {
		return false
	}

	signature, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	expected, _ := hex.DecodeString(s.sign(parts[0]))
	return hmac.Equal(signature, expected)
}

func (s *AuthService) sign(payload string) string {
	h := hmac.New(sha256.New, s.sessionSecret)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}
