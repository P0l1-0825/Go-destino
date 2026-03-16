package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders_AllPresent(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	expected := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Cache-Control":             "no-store",
		"Pragma":                    "no-cache",
	}

	for header, want := range expected {
		got := rr.Header().Get(header)
		if got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestSecurityHeaders_CSPPresent(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	csp := rr.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("Content-Security-Policy header should be set")
	}
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	pp := rr.Header().Get("Permissions-Policy")
	if pp == "" {
		t.Error("Permissions-Policy header should be set")
	}
}

func TestSecurityHeaders_PassesThrough(t *testing.T) {
	var called bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	handler := SecurityHeaders(inner)
	req := httptest.NewRequest("POST", "/api/v1/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("inner handler should be called")
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}
