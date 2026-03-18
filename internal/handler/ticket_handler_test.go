package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTicketHandler_Purchase_InvalidJSON(t *testing.T) {
	h := &TicketHandler{} // nil svc — we never reach it
	req := httptest.NewRequest("POST", "/api/v1/tickets/purchase", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.Purchase(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestTicketHandler_Validate_InvalidJSON(t *testing.T) {
	h := &TicketHandler{}
	req := httptest.NewRequest("POST", "/api/v1/tickets/validate", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.Validate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestTicketHandler_Cancel_MissingID(t *testing.T) {
	h := &TicketHandler{}
	req := httptest.NewRequest("DELETE", "/api/v1/tickets/", nil)
	rr := httptest.NewRecorder()

	h.Cancel(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}
