package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_DefaultAllowAll(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS Allow-Origin = %q, want *", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("CORS Allow-Methods should be set")
	}
}

func TestCORS_PreflightReturns204(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called on preflight")
	})

	handler := CORS(inner)
	req := httptest.NewRequest("OPTIONS", "/api/v1/bookings", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("preflight expected 204, got %d", rr.Code)
	}
}

func TestCORSWithConfig_AllowedOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://godestino-admin.example.com"},
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSWithConfig(cfg)(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://godestino-admin.example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://godestino-admin.example.com" {
		t.Errorf("Allow-Origin = %q, want matching origin", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q, want true", got)
	}
}

func TestCORSWithConfig_DisallowedOrigin_Preflight(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://allowed.com"},
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach inner handler")
	})

	handler := CORSWithConfig(cfg)(inner)
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("disallowed origin preflight expected 403, got %d", rr.Code)
	}
}

func TestCORSWithConfig_DisallowedOrigin_Regular(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://allowed.com"},
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSWithConfig(cfg)(inner)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Regular request still served but without CORS headers
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed origin should not get Allow-Origin, got %q", got)
	}
}

func TestCORSHeaders_IncludeCustomHeaders(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	allowHeaders := rr.Header().Get("Access-Control-Allow-Headers")
	for _, expected := range []string{"Authorization", "X-Tenant-ID", "X-Kiosk-ID"} {
		if !contains(allowHeaders, expected) {
			t.Errorf("Allow-Headers should include %q, got %q", expected, allowHeaders)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
