package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters = make(map[string]*ipLimiter)
	mu       sync.Mutex
)

func init() {
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			mu.Lock()
			for ip, il := range limiters {
				if time.Since(il.lastSeen) > 5*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func getLimiter(ip string, r rate.Limit, b int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	il, exists := limiters[ip]
	if !exists {
		limiter := rate.NewLimiter(r, b)
		limiters[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	il.lastSeen = time.Now()
	return il.limiter
}

// RateLimit limits requests per IP.
// r: requests per second, b: burst allowance
func RateLimit(r rate.Limit, b int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rHttp *http.Request) {
			ip := rHttp.RemoteAddr

			limiter := getLimiter(ip, r, b)
			if !limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]string{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "Too many requests, please try again later",
					},
				})
				return
			}

			next.ServeHTTP(w, rHttp)
		})
	}
}
