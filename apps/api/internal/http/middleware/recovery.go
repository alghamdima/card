package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"cards-api/internal/http/response"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}
				log.Printf("[%s] PANIC: %v\n%s", RequestIDFrom(r.Context()), rvr, debug.Stack())
				response.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
