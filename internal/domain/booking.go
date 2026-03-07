package domain

import "time"

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingConfirmed BookingStatus = "confirmed"
	BookingAssigned  BookingStatus = "assigned"
	BookingStarted   BookingStatus = "started"
	BookingCompleted BookingStatus = "completed"
	BookingCancelled BookingStatus = "cancelled"
)

type ServiceType string

const (
	ServiceTaxi    ServiceType = "taxi"
	ServiceShuttle ServiceType = "shuttle"
	ServiceVan     ServiceType = "van"
	ServiceBus     ServiceType = "bus"
)

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
	ScheduledAt     *time.Time    `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt       *time.Time    `json:"started_at,omitempty" db:"started_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty" db:"completed_at"`
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
