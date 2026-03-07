package domain

import "time"

type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "low"
	SeverityMedium   IncidentSeverity = "medium"
	SeverityHigh     IncidentSeverity = "high"
	SeverityCritical IncidentSeverity = "critical"
)

// SafetyIncident represents a reported safety event.
type SafetyIncident struct {
	ID          string           `json:"id" db:"id"`
	TenantID    string           `json:"tenant_id" db:"tenant_id"`
	BookingID   string           `json:"booking_id,omitempty" db:"booking_id"`
	DriverID    string           `json:"driver_id,omitempty" db:"driver_id"`
	ReportedBy  string           `json:"reported_by" db:"reported_by"`
	Type        string           `json:"type" db:"type"` // sos, accident, deviation, complaint
	Severity    IncidentSeverity `json:"severity" db:"severity"`
	Description string           `json:"description" db:"description"`
	Lat         float64          `json:"lat" db:"lat"`
	Lng         float64          `json:"lng" db:"lng"`
	Status      string           `json:"status" db:"status"` // open, investigating, resolved
	ResolvedBy  string           `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt  *time.Time       `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
}

// SOSAlert represents an emergency panic button press.
type SOSAlert struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	BookingID string    `json:"booking_id" db:"booking_id"`
	Lat       float64   `json:"lat" db:"lat"`
	Lng       float64   `json:"lng" db:"lng"`
	Status    string    `json:"status" db:"status"` // triggered, acknowledged, resolved
	LocalEmergencyNumber string `json:"local_emergency_number"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RideCheck represents an AI-triggered safety check.
type RideCheck struct {
	BookingID string `json:"booking_id"`
	TriggerReason string `json:"trigger_reason"` // long_stop, route_deviation, acceleration_anomaly
	UserResponse  string `json:"user_response,omitempty"` // ok, need_help, no_response
	Timestamp     int64  `json:"timestamp"`
}

// EmergencyNumbers by country.
var EmergencyNumbers = map[string]string{
	"MX": "911",
	"CO": "123",
	"PE": "105",
	"CL": "133",
	"AR": "911",
	"BR": "190",
}
