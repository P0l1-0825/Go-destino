package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig allows environment-based CORS configuration.
type CORSConfig struct {
	AllowedOrigins []string // empty = allow all (dev only)
}

// CORS adds Cross-Origin Resource Sharing headers.
// In production, set CORS_ORIGINS to restrict allowed origins.
func CORS(next http.Handler) http.Handler {
	return CORSWithConfig(CORSConfig{})(next)
}

// CORSWithConfig returns CORS middleware with explicit origin whitelist.
func CORSWithConfig(cfg CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if len(cfg.AllowedOrigins) == 0 {
				// Development mode: allow all
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && isAllowedOrigin(origin, cfg.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else if origin != "" {
				// Origin not allowed — don't set CORS headers
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-Kiosk-ID, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" {
			return true
		}
		if strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}
