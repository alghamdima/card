package handlers

import (
	"errors"
	"net/http"
	"strings"

	"cards-api/internal/http/response"
	"cards-api/internal/service"
)

const sessionCookieName = "admin_token"

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !decodeJSON(w, r, &req, maxSmallBody) {
		return
	}

	token, err := h.authService.Login(req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid admin password")
			return
		}
		writeServiceError(w, r, err, "LOGIN_FAILED", "Authentication failed")
		return
	}

	// The SPA authenticates with the bearer token; the cookie keeps plain
	// browser navigation and curl usable.
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})

	response.OK(w, map[string]any{
		"token": token,
		"user":  map[string]string{"role": "admin"},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})

	response.OK(w, map[string]bool{"loggedOut": true})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]any{
		"role":          "admin",
		"authenticated": true,
	})
}

// isHTTPS detects TLS both on direct connections and behind a terminating proxy.
func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
