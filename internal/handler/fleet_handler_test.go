package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFleetRegisterDriver_InvalidJSON(t *testing.T) {
	h := &FleetHandler{}

	req := httptest.NewRequest("POST", "/api/v1/fleet/drivers", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.RegisterDriver(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestFleetRegisterVehicle_InvalidJSON(t *testing.T) {
	h := &FleetHandler{}

	req := httptest.NewRequest("POST", "/api/v1/fleet/vehicles", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	h.RegisterVehicle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestFleetUpdateLocation_InvalidJSON(t *testing.T) {
	h := &FleetHandler{}

	req := httptest.NewRequest("PUT", "/", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.UpdateLocation(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestFleetUpdateStatus_InvalidJSON(t *testing.T) {
	h := &FleetHandler{}

	req := httptest.NewRequest("PUT", "/", strings.NewReader("{!!}"))
	rr := httptest.NewRecorder()

	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
