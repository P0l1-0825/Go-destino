package middleware

import (
	"context"
	"net/http"

	"github.com/P0l1-0825/Go-destino/pkg/response"
)

// TenantFromHeader extracts tenant ID from X-Tenant-ID header for multi-tenant isolation.
func TenantFromHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			// Try from JWT claims if already authenticated
			if t := r.Context().Value(ContextTenantID); t != nil {
				next.ServeHTTP(w, r)
				return
			}
			response.Error(w, http.StatusBadRequest, "X-Tenant-ID header is required")
			return
		}

		ctx := context.WithValue(r.Context(), ContextTenantID, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
