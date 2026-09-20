package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				reqID, _ := r.Context().Value(RequestIDKey).(string)
				log.Printf("[%s] PANIC: %v\n%s", reqID, rvr, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]string{
						"code":    "INTERNAL_SERVER_ERROR",
						"message": "An unexpected error occurred",
					},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
