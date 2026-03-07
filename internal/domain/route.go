package domain

import "time"

type TransportType string

const (
	TransportBus     TransportType = "bus"
	TransportMetro   TransportType = "metro"
	TransportTaxi    TransportType = "taxi"
	TransportShuttle TransportType = "shuttle"
	TransportFerry   TransportType = "ferry"
)

// Route represents a transport route operated by a tenant.
type Route struct {
	ID            string        `json:"id" db:"id"`
	TenantID      string        `json:"tenant_id" db:"tenant_id"`
	Name          string        `json:"name" db:"name"`
	Code          string        `json:"code" db:"code"`
	TransportType TransportType `json:"transport_type" db:"transport_type"`
	Origin        string        `json:"origin" db:"origin"`
	Destination   string        `json:"destination" db:"destination"`
	Stops         []Stop        `json:"stops,omitempty"`
	Schedule      []Schedule    `json:"schedule,omitempty"`
	PriceCents    int64         `json:"price_cents" db:"price_cents"`
	Currency      string        `json:"currency" db:"currency"`
	Active        bool          `json:"active" db:"active"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// Stop represents a location along a route.
type Stop struct {
	ID       string  `json:"id" db:"id"`
	RouteID  string  `json:"route_id" db:"route_id"`
	Name     string  `json:"name" db:"name"`
	Lat      float64 `json:"lat" db:"lat"`
	Lng      float64 `json:"lng" db:"lng"`
	Sequence int     `json:"sequence" db:"sequence"`
}

// Schedule represents a departure time for a route.
type Schedule struct {
	ID        string `json:"id" db:"id"`
	RouteID   string `json:"route_id" db:"route_id"`
	DayOfWeek int    `json:"day_of_week" db:"day_of_week"` // 0=Sun, 6=Sat
	Departure string `json:"departure" db:"departure"`     // HH:MM format
	Active    bool   `json:"active" db:"active"`
}

type CreateRouteRequest struct {
	Name          string        `json:"name"`
	Code          string        `json:"code"`
	TransportType TransportType `json:"transport_type"`
	Origin        string        `json:"origin"`
	Destination   string        `json:"destination"`
	PriceCents    int64         `json:"price_cents"`
	Currency      string        `json:"currency"`
}
