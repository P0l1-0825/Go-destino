package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaymentGetByID_MissingID(t *testing.T) {
	h := &PaymentHandler{}

	// PathValue("id") returns "" when no path param matched
	req := httptest.NewRequest("GET", "/api/v1/payments/", nil)
	rr := httptest.NewRecorder()

	h.GetByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "id")
}

func TestPaymentGetByBooking_MissingBookingID(t *testing.T) {
	h := &PaymentHandler{}

	req := httptest.NewRequest("GET", "/api/v1/payments/booking/", nil)
	rr := httptest.NewRecorder()

	h.GetByBooking(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "bookingId")
}

func TestPaymentRefund_MissingID(t *testing.T) {
	h := &PaymentHandler{}

	req := httptest.NewRequest("POST", "/api/v1/payments//refund", nil)
	rr := httptest.NewRecorder()

	h.Refund(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "id")
}
