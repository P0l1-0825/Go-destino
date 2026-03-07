package domain

import "time"

// --- Kiosk AI UX Models ---

// KioskSuggestion represents an AI-generated suggestion for the kiosk screen.
type KioskSuggestion struct {
	Type        string `json:"type"`         // "destination", "service", "promo"
	Title       string `json:"title"`        // Display title
	Subtitle    string `json:"subtitle"`     // Short description
	IconURL     string `json:"icon_url"`     // Display icon
	ServiceType string `json:"service_type"` // taxi, shuttle, van, bus
	DropoffLat  float64 `json:"dropoff_lat,omitempty"`
	DropoffLng  float64 `json:"dropoff_lng,omitempty"`
	DropoffName string `json:"dropoff_name,omitempty"`
	PriceCents  int64  `json:"price_cents,omitempty"`
	Currency    string `json:"currency,omitempty"`
	ETAMinutes  int    `json:"eta_minutes,omitempty"`
	Priority    int    `json:"priority"` // lower = higher priority
}

// KioskSuggestionsResponse is what the kiosk screen receives for smart display.
type KioskSuggestionsResponse struct {
	Suggestions    []KioskSuggestion `json:"suggestions"`
	PopularRoutes  []PopularRoute    `json:"popular_routes"`
	DemandLevel    string            `json:"demand_level"`    // low, normal, high, surge
	WelcomeMessage string            `json:"welcome_message"` // Localized greeting
	Lang           string            `json:"lang"`
	GeneratedAt    time.Time         `json:"generated_at"`
}

// PopularRoute is a frequently used route shown on the kiosk.
type PopularRoute struct {
	Name       string `json:"name"`
	Code       string `json:"code,omitempty"`
	Origin     string `json:"origin"`
	Destination string `json:"destination"`
	PriceCents int64  `json:"price_cents"`
	Currency   string `json:"currency"`
	Frequency  int    `json:"frequency"` // times booked today
}

// FlightLookupResponse bundles flight info with transport options.
type FlightLookupResponse struct {
	FlightNumber  string             `json:"flight_number"`
	Airline       string             `json:"airline"`
	Origin        string             `json:"origin"`
	Status        string             `json:"status"` // on_time, delayed, landed, cancelled
	ArrivalTime   *time.Time         `json:"arrival_time,omitempty"`
	Terminal      string             `json:"terminal"`
	Gate          string             `json:"gate,omitempty"`
	Passengers    int                `json:"estimated_passengers"`
	TransportOptions []TransportOption `json:"transport_options"`
}

// TransportOption is a suggested transport for a flight arrival.
type TransportOption struct {
	ServiceType   ServiceType `json:"service_type"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	PriceCents    int64       `json:"price_cents"`
	Currency      string      `json:"currency"`
	ETAMinutes    int         `json:"eta_minutes"`
	MaxPassengers int         `json:"max_passengers"`
	Recommended   bool        `json:"recommended"`
	ReasonTag     string      `json:"reason_tag,omitempty"` // "best_value", "fastest", "group"
}

// QuickBookRequest is a streamlined booking request optimized for kiosk UX.
// Requires minimal input: just pick a destination and pay.
type QuickBookRequest struct {
	DestinationID  string      `json:"destination_id,omitempty"` // Pre-defined destination
	DropoffAddress string      `json:"dropoff_address,omitempty"`
	DropoffLat     float64     `json:"dropoff_lat"`
	DropoffLng     float64     `json:"dropoff_lng"`
	ServiceType    ServiceType `json:"service_type"`
	PassengerCount int         `json:"passenger_count"`
	PaymentMethod  string      `json:"payment_method"` // cash, card, qr, transport_card
	CardNumber     string      `json:"card_number,omitempty"` // For transport_card payment
	FlightNumber   string      `json:"flight_number,omitempty"`
	Lang           string      `json:"lang,omitempty"`
}

// QuickBookResponse returns everything the kiosk needs to display confirmation.
type QuickBookResponse struct {
	Booking       Booking         `json:"booking"`
	Payment       Payment         `json:"payment"`
	Receipt       KioskReceipt    `json:"receipt"`
	QRCode        string          `json:"qr_code"`
	Message       string          `json:"message"`
	ETAMinutes    int             `json:"eta_minutes"`
}

// KioskReceipt is the printable receipt for kiosk transactions.
type KioskReceipt struct {
	ReceiptNumber string    `json:"receipt_number"`
	KioskID       string    `json:"kiosk_id"`
	KioskName     string    `json:"kiosk_name"`
	TenantID      string    `json:"tenant_id"`
	BookingNumber string    `json:"booking_number,omitempty"`
	TicketIDs     []string  `json:"ticket_ids,omitempty"`
	ServiceType   string    `json:"service_type"`
	Pickup        string    `json:"pickup"`
	Dropoff       string    `json:"dropoff"`
	Passengers    int       `json:"passengers"`
	PriceCents    int64     `json:"price_cents"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"payment_method"`
	PaymentRef    string    `json:"payment_ref"`
	FlightNumber  string    `json:"flight_number,omitempty"`
	QRCode        string    `json:"qr_code"`
	IssuedAt      time.Time `json:"issued_at"`
	Footer        string    `json:"footer"` // Legal/promo text
}

// KioskSession tracks an active user interaction at a kiosk.
type KioskSession struct {
	ID         string     `json:"id" db:"id"`
	KioskID    string     `json:"kiosk_id" db:"kiosk_id"`
	TenantID   string     `json:"tenant_id" db:"tenant_id"`
	Lang       string     `json:"lang" db:"lang"`
	StartedAt  time.Time  `json:"started_at" db:"started_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	BookingID  string     `json:"booking_id,omitempty" db:"booking_id"`
	TicketIDs  string     `json:"ticket_ids,omitempty" db:"ticket_ids"` // comma-separated
	Outcome    string     `json:"outcome" db:"outcome"` // completed, abandoned, timeout
	StepCount  int        `json:"step_count" db:"step_count"`
	DurationMs int64      `json:"duration_ms" db:"duration_ms"`
}

// ServiceRecommendation is what AI returns when recommending a service type.
type ServiceRecommendation struct {
	ServiceType   ServiceType `json:"service_type"`
	Confidence    float64     `json:"confidence"` // 0-1
	Reason        string      `json:"reason"`
	PriceCents    int64       `json:"price_cents"`
	Currency      string      `json:"currency"`
	ETAMinutes    int         `json:"eta_minutes"`
}
