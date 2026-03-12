package domain

import "time"

type FlightStatus string

const (
	FlightScheduled FlightStatus = "scheduled"
	FlightEnRoute   FlightStatus = "en_route"
	FlightLanded    FlightStatus = "landed"
	FlightDelayed   FlightStatus = "delayed"
	FlightCancelled FlightStatus = "cancelled"
	FlightDiverted  FlightStatus = "diverted"
)

// FlightInfo represents flight data from Cirium/FlightAware.
type FlightInfo struct {
	FlightNumber   string       `json:"flight_number"`
	Airline        string       `json:"airline"`
	Origin         string       `json:"origin"`
	Destination    string       `json:"destination"`
	Status         FlightStatus `json:"status"`
	ScheduledAt    time.Time    `json:"scheduled_at"`
	EstimatedAt    *time.Time   `json:"estimated_at,omitempty"`
	ActualAt       *time.Time   `json:"actual_at,omitempty"`
	Gate           string       `json:"gate,omitempty"`
	Terminal       string       `json:"terminal,omitempty"`
	BaggageBelt    string       `json:"baggage_belt,omitempty"`
	DelayMinutes   int          `json:"delay_minutes"`
	LastUpdated    time.Time    `json:"last_updated"`
}

// IROPSEvent represents an Irregular Operations event.
type IROPSEvent struct {
	ID            string    `json:"id" db:"id"`
	FlightNumber  string    `json:"flight_number" db:"flight_number"`
	EventType     string    `json:"event_type" db:"event_type"` // delay, cancel, divert
	DelayMinutes  int       `json:"delay_minutes" db:"delay_minutes"`
	AffectedBookings int   `json:"affected_bookings"`
	AutoActions   []string  `json:"auto_actions"` // notify_passenger, adjust_driver, cancel_booking
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type AirportTerminal struct {
	ID         string `json:"id" db:"id"`
	AirportID  string `json:"airport_id" db:"airport_id"`
	Name       string `json:"name" db:"name"`
	PickupZone string `json:"pickup_zone" db:"pickup_zone"`
}
