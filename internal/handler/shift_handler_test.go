package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShiftHandler_Open_InvalidJSON(t *testing.T) {
	h := &ShiftHandler{}
	req := httptest.NewRequest("POST", "/api/v1/shifts", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()
	h.Open(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
	assertErrorContains(t, rr, "invalid request body")
}

func TestShiftHandler_Open_MissingAirportID(t *testing.T) {
	h := &ShiftHandler{}
	body := `{"airport_id":"","terminal_id":"T1","kiosk_id":"k1"}`
	req := httptest.NewRequest("POST", "/api/v1/shifts", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.Open(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
	assertErrorContains(t, rr, "airport_id")
}

func TestShiftHandler_Open_MissingKioskID(t *testing.T) {
	h := &ShiftHandler{}
	body := `{"airport_id":"MEX","terminal_id":"T1","kiosk_id":""}`
	req := httptest.NewRequest("POST", "/api/v1/shifts", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.Open(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
	assertErrorContains(t, rr, "kiosk_id")
}

func TestShiftHandler_Close_InvalidJSON(t *testing.T) {
	h := &ShiftHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/shifts/{id}/close", h.Close)
	req := httptest.NewRequest("PUT", "/api/v1/shifts/s1/close", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}

func TestShiftHandler_Close_NegativeSales(t *testing.T) {
	h := &ShiftHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/shifts/{id}/close", h.Close)
	body := `{"total_sales_cents":-100,"cash_collected_cents":0,"card_collected_cents":0}`
	req := httptest.NewRequest("PUT", "/api/v1/shifts/s1/close", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
	assertErrorContains(t, rr, "negative")
}

func TestShiftHandler_GetByID_Missing(t *testing.T) {
	h := &ShiftHandler{}
	req := httptest.NewRequest("GET", "/api/v1/shifts/", nil)
	rr := httptest.NewRecorder()
	h.GetByID(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}

func TestShiftHandler_ListByKiosk_Missing(t *testing.T) {
	h := &ShiftHandler{}
	req := httptest.NewRequest("GET", "/api/v1/shifts/kiosk/", nil)
	rr := httptest.NewRecorder()
	h.ListByKiosk(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}
