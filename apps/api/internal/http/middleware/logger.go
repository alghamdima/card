package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer (Flush, deadlines).
func (rw *responseWriter) Unwrap() http.ResponseWriter { return rw.ResponseWriter }

// Logger logs one line per request. The query string is left out on purpose:
// it can carry search terms containing personal data.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		log.Printf("[%s] %s %s %d %dB in %v", RequestIDFrom(r.Context()), r.Method, r.URL.Path, rw.status, rw.bytesWritten, time.Since(start))
	})
}
