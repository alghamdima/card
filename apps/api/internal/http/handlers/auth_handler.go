package handlers

import (
	"encoding/json"
	"net/http"

	"cards-api/internal/service"
)

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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid JSON payload")
		return
	}

	token, err := h.authService.Login(req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			JSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid admin password")
			return
		}
		JSONError(w, http.StatusInternalServerError, "LOGIN_FAILED", "Authentication failed")
		return
	}

	// Set httpOnly cookie as well for flexibility
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	JSONOK(w, map[string]interface{}{
		"token": token,
		"user": map[string]string{
			"role": "admin",
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	JSONOK(w, map[string]bool{"loggedOut": true})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	JSONOK(w, map[string]interface{}{
		"role":          "admin",
		"authenticated": true,
	})
}
