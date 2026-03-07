package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	resetAt  time.Time
}

// RateLimit limits requests per IP within a time window.
// In production, replace with Redis INCR for distributed rate limiting.
func RateLimit(requestsPerWindow int, window time.Duration) func(http.Handler) http.Handler {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerWindow,
		window:   window,
	}

	// Cleanup stale entries every window duration
	go func() {
		for {
			time.Sleep(window)
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.After(v.resetAt) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			// Prefer user_id if authenticated
			if uid := GetUserID(r.Context()); uid != "" {
				ip = "user:" + uid
			}

			rl.mu.Lock()
			v, exists := rl.visitors[ip]
			now := time.Now()
			if !exists || now.After(v.resetAt) {
				rl.visitors[ip] = &visitor{count: 1, resetAt: now.Add(window)}
				rl.mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if v.count >= rl.rate {
				rl.mu.Unlock()
				w.Header().Set("Retry-After", v.resetAt.Format(time.RFC1123))
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			v.count++
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
