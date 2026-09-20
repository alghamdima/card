package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"cards-api/internal/service"
)

func RequireAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				cookie, err := r.Cookie("admin_token")
				if err == nil {
					token = cookie.Value
				}
			}

			if token == "" || !authService.ValidateToken(token) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "Authentication required",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
