package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKioskHandler_Register_InvalidJSON(t *testing.T) {
	h := &KioskHandler{}
	req := httptest.NewRequest("POST", "/api/v1/kiosks", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}


func TestKioskHandler_UpdateStatus_InvalidJSON(t *testing.T) {
	h := &KioskHandler{}

	req := httptest.NewRequest("PUT", "/api/v1/kiosks/k1/status", strings.NewReader("{bad"))
	req.SetPathValue("id", "k1")
	rr := httptest.NewRecorder()

	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}
