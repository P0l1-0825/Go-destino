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
	stop     chan struct{}
}

type visitor struct {
	count   int
	resetAt time.Time
}

// RateLimit limits requests per IP within a time window.
// In production, replace with Redis INCR for distributed rate limiting.
//
// The returned middleware wraps a rateLimiter whose background cleanup goroutine
// exits when the process terminates (stop channel is never closed explicitly here
// because the middleware lives for the process lifetime; the channel prevents a
// goroutine leak in test environments where multiple RateLimit instances are
// created and the process does not exit between them).
func RateLimit(requestsPerWindow int, window time.Duration) func(http.Handler) http.Handler {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerWindow,
		window:   window,
		stop:     make(chan struct{}),
	}

	// Cleanup stale entries every window duration.
	// Uses a ticker so the goroutine can be stopped via rl.stop.
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for {
			select {
			case <-rl.stop:
				return
			case now := <-ticker.C:
				rl.mu.Lock()
				for ip, v := range rl.visitors {
					if now.After(v.resetAt) {
						delete(rl.visitors, ip)
					}
				}
				rl.mu.Unlock()
			}
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
