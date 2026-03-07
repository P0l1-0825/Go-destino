package service

import (
	"context"
	"fmt"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/google/uuid"
)

type FlightService struct {
	bookingRepo *repository.BookingRepository
	notifSvc    *NotificationService
}

func NewFlightService(bookingRepo *repository.BookingRepository, notifSvc *NotificationService) *FlightService {
	return &FlightService{
		bookingRepo: bookingRepo,
		notifSvc:    notifSvc,
	}
}

// GetFlightInfo simulates fetching flight data from Cirium/FlightAware.
// In production, integrate with real flight data APIs.
func (s *FlightService) GetFlightInfo(_ context.Context, flightNumber string) (*domain.FlightInfo, error) {
	now := time.Now()
	actualAt := now.Add(-15 * time.Minute)
	return &domain.FlightInfo{
		FlightNumber: flightNumber,
		Airline:      inferAirline(flightNumber),
		Status:       domain.FlightLanded,
		ScheduledAt:  now.Add(-30 * time.Minute),
		ActualAt:     &actualAt,
		Gate:         "A12",
		Terminal:     "1",
		BaggageBelt:  "3",
		DelayMinutes: 15,
		LastUpdated:  now,
	}, nil
}

// ListArrivals returns simulated arrivals for an airport.
func (s *FlightService) ListArrivals(_ context.Context, airportCode string) ([]domain.FlightInfo, error) {
	now := time.Now()
	t1 := now.Add(-15 * time.Minute)
	t2 := now.Add(65 * time.Minute)
	flights := []domain.FlightInfo{
		{FlightNumber: "AM2341", Airline: "Aeromexico", Origin: "GDL", Destination: airportCode, Status: domain.FlightLanded, ScheduledAt: now.Add(-30 * time.Minute), ActualAt: &t1, Terminal: "1", Gate: "A12"},
		{FlightNumber: "VB3201", Airline: "VivaAerobus", Origin: "CUN", Destination: airportCode, Status: domain.FlightEnRoute, ScheduledAt: now.Add(45 * time.Minute), Terminal: "2", Gate: "B7"},
		{FlightNumber: "Y4812", Airline: "Volaris", Origin: "MTY", Destination: airportCode, Status: domain.FlightDelayed, ScheduledAt: now.Add(20 * time.Minute), ActualAt: &t2, DelayMinutes: 45, Terminal: "1", Gate: "C3"},
		{FlightNumber: "UA1923", Airline: "United", Origin: "IAH", Destination: airportCode, Status: domain.FlightScheduled, ScheduledAt: now.Add(2 * time.Hour), Terminal: "1", Gate: "D15"},
	}
	return flights, nil
}

// HandleIROPS processes Irregular Operations events.
// Automatically adjusts affected bookings and sends notifications.
func (s *FlightService) HandleIROPS(ctx context.Context, tenantID string, event domain.IROPSEvent) (*domain.IROPSEvent, error) {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()

	// Determine auto-actions based on delay
	if event.DelayMinutes > 30 {
		event.AutoActions = append(event.AutoActions,
			"notify_affected_passengers",
			"adjust_pickup_times",
		)
	}
	if event.DelayMinutes > 120 {
		event.AutoActions = append(event.AutoActions,
			"offer_rebooking",
			"alert_operations_team",
		)
	}
	if event.EventType == "cancellation" || event.EventType == "diversion" {
		event.AutoActions = append(event.AutoActions,
			"auto_cancel_bookings",
			"process_refunds",
			"notify_drivers_cancel",
		)
	}

	// Count affected bookings (simulated — in production, query by flight_number)
	event.AffectedBookings = 0

	return &event, nil
}

func inferAirline(flightNumber string) string {
	if len(flightNumber) < 2 {
		return "Unknown"
	}
	prefixes := map[string]string{
		"AM": "Aeromexico", "VB": "VivaAerobus", "Y4": "Volaris",
		"UA": "United", "AA": "American", "DL": "Delta",
		"LA": "LATAM", "AV": "Avianca", "CM": "Copa",
		"G3": "GOL", "AD": "Azul",
	}
	prefix := flightNumber[:2]
	if airline, ok := prefixes[prefix]; ok {
		return airline
	}
	return fmt.Sprintf("Airline(%s)", prefix)
}
