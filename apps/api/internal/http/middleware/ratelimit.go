package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"cards-api/internal/http/response"
)

const (
	limiterIdleTTL       = 5 * time.Minute
	limiterSweepInterval = time.Minute
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is a per-client token bucket. Each instance keeps its own state,
// so different routes can have different budgets.
type RateLimiter struct {
	limit      rate.Limit
	burst      int
	trustProxy bool

	mu        sync.Mutex
	visitors  map[string]*visitor
	lastSweep time.Time
	now       func() time.Time
}

// NewRateLimiter allows `burst` requests at once per client, refilled at `limit` tokens per second.
// Set trustProxy only when the API sits behind a proxy that overwrites X-Real-IP.
func NewRateLimiter(limit rate.Limit, burst int, trustProxy bool) *RateLimiter {
	return &RateLimiter{
		limit:      limit,
		burst:      burst,
		trustProxy: trustProxy,
		visitors:   make(map[string]*visitor),
		now:        time.Now,
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	if now.Sub(rl.lastSweep) > limiterSweepInterval {
		for k, v := range rl.visitors {
			if now.Sub(v.lastSeen) > limiterIdleTTL {
				delete(rl.visitors, k)
			}
		}
		rl.lastSweep = now
	}

	v, ok := rl.visitors[key]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.visitors[key] = v
	}
	v.lastSeen = now
	return v.limiter.AllowN(now, 1)
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(ClientIP(r, rl.trustProxy)) {
			w.Header().Set("Retry-After", "60")
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please try again later")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClientIP returns the caller's IP without the ephemeral port. RemoteAddr
// includes the port, which differs per connection: keying limits on it would
// give every new connection a fresh budget.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(ip) != nil {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
