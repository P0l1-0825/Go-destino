package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data in response")
	}
	if data["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", data["status"])
	}
	if data["service"] != "godestino-api" {
		t.Errorf("service = %v", data["service"])
	}
}

func TestReadyCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/ready", nil)
	rr := httptest.NewRecorder()

	ReadyCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}
