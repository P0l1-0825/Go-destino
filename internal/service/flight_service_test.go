package service

import (
	"context"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

func TestFlightService_GetFlightInfo(t *testing.T) {
	svc := &FlightService{}

	tests := []struct {
		name         string
		flightNumber string
		wantAirline  string
		wantStatus   domain.FlightStatus
	}{
		{"Aeromexico", "AM2341", "Aeromexico", domain.FlightLanded},
		{"VivaAerobus", "VB3201", "VivaAerobus", domain.FlightLanded},
		{"Volaris", "Y4812", "Volaris", domain.FlightLanded},
		{"United", "UA1923", "United", domain.FlightLanded},
		{"Delta", "DL500", "Delta", domain.FlightLanded},
		{"Unknown prefix", "XX999", "Airline(XX)", domain.FlightLanded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := svc.GetFlightInfo(context.Background(), tt.flightNumber)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.FlightNumber != tt.flightNumber {
				t.Errorf("FlightNumber = %q, want %q", info.FlightNumber, tt.flightNumber)
			}
			if info.Airline != tt.wantAirline {
				t.Errorf("Airline = %q, want %q", info.Airline, tt.wantAirline)
			}
			if info.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", info.Status, tt.wantStatus)
			}
			if info.Gate == "" {
				t.Error("Gate should not be empty")
			}
			if info.Terminal == "" {
				t.Error("Terminal should not be empty")
			}
		})
	}
}

func TestFlightService_ListArrivals(t *testing.T) {
	svc := &FlightService{}

	flights, err := svc.ListArrivals(context.Background(), "CUN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 4 {
		t.Errorf("expected 4 flights, got %d", len(flights))
	}

	// Verify all have the target airport as destination
	for _, f := range flights {
		if f.Destination != "CUN" {
			t.Errorf("Destination = %q, want CUN", f.Destination)
		}
	}

	// Verify different statuses are represented
	statusSet := make(map[domain.FlightStatus]bool)
	for _, f := range flights {
		statusSet[f.Status] = true
	}
	if len(statusSet) < 3 {
		t.Errorf("expected at least 3 different statuses, got %d", len(statusSet))
	}
}

func TestFlightService_HandleIROPS_ShortDelay(t *testing.T) {
	svc := &FlightService{}

	event := domain.IROPSEvent{
		FlightNumber: "AM100",
		EventType:    "delay",
		DelayMinutes: 20,
	}

	result, err := svc.HandleIROPS(context.Background(), "tenant-1", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Error("ID should be generated")
	}
	if len(result.AutoActions) != 0 {
		t.Errorf("expected no auto-actions for 20 min delay, got %v", result.AutoActions)
	}
}

func TestFlightService_HandleIROPS_MediumDelay(t *testing.T) {
	svc := &FlightService{}

	event := domain.IROPSEvent{
		FlightNumber: "UA500",
		EventType:    "delay",
		DelayMinutes: 60,
	}

	result, err := svc.HandleIROPS(context.Background(), "tenant-1", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.AutoActions) != 2 {
		t.Errorf("expected 2 auto-actions for 60 min delay, got %d: %v", len(result.AutoActions), result.AutoActions)
	}
}

func TestFlightService_HandleIROPS_LongDelay(t *testing.T) {
	svc := &FlightService{}

	event := domain.IROPSEvent{
		FlightNumber: "DL300",
		EventType:    "delay",
		DelayMinutes: 180,
	}

	result, err := svc.HandleIROPS(context.Background(), "tenant-1", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.AutoActions) != 4 {
		t.Errorf("expected 4 auto-actions for 180 min delay, got %d: %v", len(result.AutoActions), result.AutoActions)
	}
}

func TestFlightService_HandleIROPS_Cancellation(t *testing.T) {
	svc := &FlightService{}

	event := domain.IROPSEvent{
		FlightNumber: "VB100",
		EventType:    "cancellation",
		DelayMinutes: 0,
	}

	result, err := svc.HandleIROPS(context.Background(), "tenant-1", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.AutoActions) != 3 {
		t.Errorf("expected 3 auto-actions for cancellation, got %d: %v", len(result.AutoActions), result.AutoActions)
	}
}

func TestFlightService_HandleIROPS_Diversion(t *testing.T) {
	svc := &FlightService{}

	event := domain.IROPSEvent{
		FlightNumber: "AA200",
		EventType:    "diversion",
		DelayMinutes: 0,
	}

	result, err := svc.HandleIROPS(context.Background(), "tenant-1", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasCancel := false
	for _, a := range result.AutoActions {
		if a == "auto_cancel_bookings" {
			hasCancel = true
		}
	}
	if !hasCancel {
		t.Error("diversion should trigger auto_cancel_bookings")
	}
}

func TestInferAirline(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"AM", "Aeromexico"},
		{"VB", "VivaAerobus"},
		{"Y4", "Volaris"},
		{"UA", "United"},
		{"AA", "American"},
		{"DL", "Delta"},
		{"LA", "LATAM"},
		{"AV", "Avianca"},
		{"CM", "Copa"},
		{"G3", "GOL"},
		{"AD", "Azul"},
		{"XX", "Airline(XX)"},
		{"Z", "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := inferAirline(tt.code)
			if got != tt.want {
				t.Errorf("inferAirline(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
