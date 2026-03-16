package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogging_PassesThrough(t *testing.T) {
	var called bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	handler := Logging(inner)
	req := httptest.NewRequest("POST", "/api/v1/bookings", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("inner handler should be called")
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestLogging_StatusWriter_DefaultsToOK(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't explicitly write status — defaults to 200
		w.Write([]byte("ok"))
	})

	handler := Logging(inner)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 default, got %d", rr.Code)
	}
}
