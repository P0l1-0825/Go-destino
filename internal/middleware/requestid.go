package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	ContextRequestID contextKey = "request_id"
	HeaderRequestID             = "X-Request-ID"
)

// RequestID injects a unique request ID into the context and response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set(HeaderRequestID, id)
		ctx := context.WithValue(r.Context(), ContextRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts request ID from context.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(ContextRequestID).(string); ok {
		return v
	}
	return ""
}
