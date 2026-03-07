package domain

import (
	"fmt"
	"time"
)

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingConfirmed BookingStatus = "confirmed"
	BookingAssigned  BookingStatus = "assigned"
	BookingStarted   BookingStatus = "started"
	BookingCompleted BookingStatus = "completed"
	BookingCancelled BookingStatus = "cancelled"
)

// ValidBookingTransition enforces the booking state machine:
//
//	pending → confirmed → assigned → started → completed
//	pending/confirmed/assigned → cancelled
func ValidBookingTransition(from, to BookingStatus) error {
	allowed := map[BookingStatus][]BookingStatus{
		BookingPending:   {BookingConfirmed, BookingCancelled},
		BookingConfirmed: {BookingAssigned, BookingCancelled},
		BookingAssigned:  {BookingStarted, BookingCancelled},
		BookingStarted:   {BookingCompleted},
	}

	targets, ok := allowed[from]
	if !ok {
		return fmt.Errorf("cannot transition from terminal status %s", from)
	}
	for _, t := range targets {
		if t == to {
			return nil
		}
	}
	return fmt.Errorf("invalid transition: %s → %s", from, to)
}

type ServiceType string

const (
	ServiceTaxi    ServiceType = "taxi"
	ServiceShuttle ServiceType = "shuttle"
	ServiceVan     ServiceType = "van"
	ServiceBus     ServiceType = "bus"
)

// ValidServiceType returns true if the service type is known.
func ValidServiceType(s string) bool {
	switch ServiceType(s) {
	case ServiceTaxi, ServiceShuttle, ServiceVan, ServiceBus:
		return true
	}
	return false
}

// Booking represents a transport reservation (airport transfer, ride, etc.).
type Booking struct {
	ID              string        `json:"id" db:"id"`
	BookingNumber   string        `json:"booking_number" db:"booking_number"`
	TenantID        string        `json:"tenant_id" db:"tenant_id"`
	UserID          string        `json:"user_id,omitempty" db:"user_id"`
	KioskID         string        `json:"kiosk_id,omitempty" db:"kiosk_id"`
	RouteID         string        `json:"route_id,omitempty" db:"route_id"`
	DriverID        string        `json:"driver_id,omitempty" db:"driver_id"`
	VehicleID       string        `json:"vehicle_id,omitempty" db:"vehicle_id"`
	Status          BookingStatus `json:"status" db:"status"`
	ServiceType     ServiceType   `json:"service_type" db:"service_type"`
	PickupAddress   string        `json:"pickup_address" db:"pickup_address"`
	DropoffAddress  string        `json:"dropoff_address" db:"dropoff_address"`
	PickupLat       float64       `json:"pickup_lat" db:"pickup_lat"`
	PickupLng       float64       `json:"pickup_lng" db:"pickup_lng"`
	DropoffLat      float64       `json:"dropoff_lat" db:"dropoff_lat"`
	DropoffLng      float64       `json:"dropoff_lng" db:"dropoff_lng"`
	PassengerCount  int           `json:"passenger_count" db:"passenger_count"`
	PriceCents      int64         `json:"price_cents" db:"price_cents"`
	Currency        string        `json:"currency" db:"currency"`
	PaymentID       string        `json:"payment_id,omitempty" db:"payment_id"`
	FlightNumber    string        `json:"flight_number,omitempty" db:"flight_number"`
	CancelReason    string        `json:"cancel_reason,omitempty" db:"cancel_reason"`
	ScheduledAt     *time.Time    `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt       *time.Time    `json:"started_at,omitempty" db:"started_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty" db:"completed_at"`
	CancelledAt     *time.Time    `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

type CreateBookingRequest struct {
	RouteID        string      `json:"route_id"`
	ServiceType    ServiceType `json:"service_type"`
	PickupAddress  string      `json:"pickup_address"`
	DropoffAddress string      `json:"dropoff_address"`
	PickupLat      float64     `json:"pickup_lat"`
	PickupLng      float64     `json:"pickup_lng"`
	DropoffLat     float64     `json:"dropoff_lat"`
	DropoffLng     float64     `json:"dropoff_lng"`
	PassengerCount int         `json:"passenger_count"`
	FlightNumber   string      `json:"flight_number,omitempty"`
	ScheduledAt    *time.Time  `json:"scheduled_at,omitempty"`
	PaymentMethod  string      `json:"payment_method"`
}

type EstimateRequest struct {
	ServiceType    ServiceType `json:"service_type"`
	PickupLat      float64     `json:"pickup_lat"`
	PickupLng      float64     `json:"pickup_lng"`
	DropoffLat     float64     `json:"dropoff_lat"`
	DropoffLng     float64     `json:"dropoff_lng"`
	PassengerCount int         `json:"passenger_count"`
}

type EstimateResponse struct {
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`
	ETAMinutes int    `json:"eta_minutes"`
	Distance   string `json:"distance"`
}

// AssignDriverRequest is used to assign a driver+vehicle to a booking.
type AssignDriverRequest struct {
	DriverID  string `json:"driver_id"`
	VehicleID string `json:"vehicle_id"`
}

// CancelBookingRequest carries an optional reason for cancellation.
type CancelBookingRequest struct {
	Reason string `json:"reason"`
}

// ListBookingsFilter provides filtering for booking queries.
type ListBookingsFilter struct {
	TenantID    string
	UserID      string
	Status      BookingStatus
	ServiceType ServiceType
	Offset      int
	Limit       int
}
