package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

// A client-supplied ID is only trusted when it is short and printable, so it cannot forge log lines.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if !validRequestID.MatchString(reqID) {
			buf := make([]byte, 16)
			_, _ = rand.Read(buf)
			reqID = hex.EncodeToString(buf)
		}

		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom returns the request ID stored by the RequestID middleware.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}
