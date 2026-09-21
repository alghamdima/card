package middleware

import (
	"net/http"
	"strings"

	"cards-api/internal/http/response"
	"cards-api/internal/service"
)

const sessionCookieName = "admin_token"

func RequireAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authService.ValidateToken(extractToken(r)) {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractToken prefers the Authorization header and falls back to the session cookie.
func extractToken(r *http.Request) string {
	if scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " "); ok && strings.EqualFold(scheme, "Bearer") {
		return strings.TrimSpace(token)
	}
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		return cookie.Value
	}
	return ""
}
