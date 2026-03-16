package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBookingCreate_InvalidJSON(t *testing.T) {
	h := &BookingHandler{}

	req := httptest.NewRequest("POST", "/api/v1/bookings", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestBookingCreate_MissingPickup(t *testing.T) {
	h := &BookingHandler{}

	body := `{"pickup_address":"","dropoff_address":"Hotel Zona","service_type":"taxi","passenger_count":2}`
	req := httptest.NewRequest("POST", "/api/v1/bookings", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "pickup_address")
}

func TestBookingCreate_MissingDropoff(t *testing.T) {
	h := &BookingHandler{}

	body := `{"pickup_address":"Airport T1","dropoff_address":"","service_type":"taxi","passenger_count":2}`
	req := httptest.NewRequest("POST", "/api/v1/bookings", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "dropoff_address")
}

func TestBookingCreate_InvalidServiceType(t *testing.T) {
	h := &BookingHandler{}

	body := `{"pickup_address":"Airport","dropoff_address":"Hotel","service_type":"helicopter","passenger_count":2}`
	req := httptest.NewRequest("POST", "/api/v1/bookings", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "service_type")
}

func TestBookingCreate_InvalidPassengerCount(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"zero passengers", `{"pickup_address":"A","dropoff_address":"B","service_type":"taxi","passenger_count":0}`},
		{"too many", `{"pickup_address":"A","dropoff_address":"B","service_type":"taxi","passenger_count":51}`},
		{"negative", `{"pickup_address":"A","dropoff_address":"B","service_type":"taxi","passenger_count":-1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &BookingHandler{}
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.Create(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
			assertErrorContains(t, rr, "passenger_count")
		})
	}
}

func TestBookingAssignDriver_InvalidJSON(t *testing.T) {
	h := &BookingHandler{}

	req := httptest.NewRequest("POST", "/api/v1/bookings/123/assign", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.AssignDriver(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestBookingAssignDriver_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing driver_id", `{"driver_id":"","vehicle_id":"v-1"}`},
		{"missing vehicle_id", `{"driver_id":"d-1","vehicle_id":""}`},
		{"both missing", `{"driver_id":"","vehicle_id":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &BookingHandler{}
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.AssignDriver(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestBookingUpdateStatus_InvalidJSON(t *testing.T) {
	h := &BookingHandler{}

	req := httptest.NewRequest("PUT", "/", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestBookingUpdateStatus_EmptyStatus(t *testing.T) {
	h := &BookingHandler{}

	req := httptest.NewRequest("PUT", "/", strings.NewReader(`{"status":""}`))
	rr := httptest.NewRecorder()

	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "status")
}

func TestBookingEstimate_InvalidJSON(t *testing.T) {
	h := &BookingHandler{}

	req := httptest.NewRequest("POST", "/", strings.NewReader("!!"))
	rr := httptest.NewRecorder()

	h.Estimate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestBookingEstimate_InvalidServiceType(t *testing.T) {
	h := &BookingHandler{}

	body := `{"service_type":"limo","distance_km":10}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Estimate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "service_type")
}
