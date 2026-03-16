package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/P0l1-0825/Go-destino/pkg/response"
)

func TestRecovery_NoPanic(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Recovery(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRecovery_CatchesPanic(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went terribly wrong")
	})

	handler := Recovery(inner)
	req := httptest.NewRequest("GET", "/api/v1/crash", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}

	var resp response.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Success {
		t.Error("panic response should have success=false")
	}
	if resp.Error != "internal server error" {
		t.Errorf("error = %q, want 'internal server error'", resp.Error)
	}
}

func TestRecovery_CatchesNilPanic(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var s *string
		_ = *s // nil pointer dereference
	})

	handler := Recovery(inner)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on nil panic, got %d", rr.Code)
	}
}
