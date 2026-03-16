package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesNew(t *testing.T) {
	var gotID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should generate a UUID
	if gotID == "" {
		t.Error("request ID should be generated")
	}
	if len(gotID) != 36 { // UUID format: 8-4-4-4-12
		t.Errorf("request ID length = %d, want 36 (UUID)", len(gotID))
	}

	// Should also be in response header
	headerID := rr.Header().Get(HeaderRequestID)
	if headerID != gotID {
		t.Errorf("response header X-Request-ID = %q, context = %q", headerID, gotID)
	}
}

func TestRequestID_PreservesExisting(t *testing.T) {
	var gotID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderRequestID, "custom-req-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if gotID != "custom-req-123" {
		t.Errorf("should preserve existing ID, got %q", gotID)
	}
	if rr.Header().Get(HeaderRequestID) != "custom-req-123" {
		t.Errorf("response header should preserve existing ID")
	}
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	if got := GetRequestID(context.Background()); got != "" {
		t.Errorf("empty context should return empty string, got %q", got)
	}
}
