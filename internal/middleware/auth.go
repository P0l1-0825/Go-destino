package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextTenantID contextKey = "tenant_id"
	ContextRole     contextKey = "role"
	ContextClaims   contextKey = "claims"
)

// Auth validates JWT tokens and injects claims into context.
func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			claims, err := authSvc.ValidateToken(parts[1])
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.Subject)
			ctx = context.WithValue(ctx, ContextTenantID, claims.TenantID)
			ctx = context.WithValue(ctx, ContextRole, claims.Role)
			ctx = context.WithValue(ctx, ContextClaims, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission checks that the authenticated user has a specific permission.
func RequirePermission(perm domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(ContextRole).(domain.UserRole)
			if !ok {
				response.Error(w, http.StatusForbidden, "no role in context")
				return
			}

			if !domain.HasPermission(role, perm) {
				response.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextUserID).(string); ok {
		return v
	}
	return ""
}

// GetTenantID extracts tenant ID from context.
func GetTenantID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextTenantID).(string); ok {
		return v
	}
	return ""
}
