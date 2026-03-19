package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

func TestFlightHandler_ReportIROPS_InvalidJSON(t *testing.T) {
	h := &FlightHandler{}

	req := httptest.NewRequest("POST", "/api/v1/flights/irops", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.ReportIROPS(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestFlightHandler_GetFlightStatus_WithService(t *testing.T) {
	svc := service.NewFlightService(nil, nil)
	h := NewFlightHandler(svc)

	req := httptest.NewRequest("GET", "/api/v1/flights/AM2341", nil)
	req.SetPathValue("number", "AM2341")
	rr := httptest.NewRecorder()

	h.GetFlightStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var env response.APIResponse
	json.NewDecoder(rr.Body).Decode(&env)
	if !env.Success {
		t.Error("expected success=true")
	}
}

func TestFlightHandler_ListArrivals_WithService(t *testing.T) {
	svc := service.NewFlightService(nil, nil)
	h := NewFlightHandler(svc)

	req := httptest.NewRequest("GET", "/api/v1/flights/arrivals/CUN", nil)
	req.SetPathValue("code", "CUN")
	rr := httptest.NewRecorder()

	h.ListArrivals(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
