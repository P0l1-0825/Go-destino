package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	handler := RateLimit(5, 1*time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	handler := RateLimit(3, 1*time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use unique IP to avoid interference from other tests
	ip := "192.168.99.99:1234"

	// First 3 should pass
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// 4th should be blocked
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("4th request expected 429, got %d", rr.Code)
	}

	retryAfter := rr.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("Retry-After header should be set")
	}
}

func TestRateLimit_DifferentIPsIndependent(t *testing.T) {
	handler := RateLimit(2, 1*time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// IP A uses 2 requests (maxed out)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.10.10.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// IP B should still work
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.10.10.2:5678"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("different IP should not be rate limited, got %d", rr.Code)
	}
}

func TestRateLimit_WindowResets(t *testing.T) {
	handler := RateLimit(2, 10*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "172.16.0.1:1234"

	// Use up the limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Wait for window to expire
	time.Sleep(20 * time.Millisecond)

	// Should work again
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("after window reset, expected 200, got %d", rr.Code)
	}
}
