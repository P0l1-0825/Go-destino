package domain

import "time"

// DemandForecast represents predicted demand for a time window.
type DemandForecast struct {
	AirportID   string    `json:"airport_id"`
	Timestamp   time.Time `json:"timestamp"`
	IntervalMin int       `json:"interval_min"` // 30
	Predicted   int       `json:"predicted_demand"`
	Confidence  float64   `json:"confidence"`
	Factors     []string  `json:"factors"` // ["high_flight_arrivals", "weekend", "holiday"]
}

// DynamicPrice represents AI-calculated pricing.
type DynamicPrice struct {
	BasePrice    int64   `json:"base_price_cents"`
	FinalPrice   int64   `json:"final_price_cents"`
	Multiplier   float64 `json:"multiplier"` // 0.8x - 2.0x
	Currency     string  `json:"currency"`
	Rationale    string  `json:"rationale"`
	DemandLevel  string  `json:"demand_level"` // low, normal, high, surge
	ValidUntil   int64   `json:"valid_until"`  // unix timestamp, price lock
}

type DynamicPriceRequest struct {
	AirportID   string      `json:"airport_id"`
	ServiceType ServiceType `json:"service_type"`
	PickupLat   float64     `json:"pickup_lat"`
	PickupLng   float64     `json:"pickup_lng"`
	DropoffLat  float64     `json:"dropoff_lat"`
	DropoffLng  float64     `json:"dropoff_lng"`
	PassengerCount int     `json:"passenger_count"`
	ScheduledAt *time.Time  `json:"scheduled_at,omitempty"`
}

// FraudCheck represents fraud analysis result.
type FraudCheck struct {
	PaymentID string   `json:"payment_id"`
	UserID    string   `json:"user_id"`
	Score     float64  `json:"score"`      // 0-100
	Decision  string   `json:"decision"`   // approve, review, block
	Flags     []string `json:"flags"`      // ["velocity_anomaly", "geo_mismatch"]
	Timestamp int64    `json:"timestamp"`
}

type FraudCheckRequest struct {
	PaymentID   string `json:"payment_id"`
	UserID      string `json:"user_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	Method      string `json:"method"`
	IPAddress   string `json:"ip_address"`
	DeviceID    string `json:"device_id"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

// ChatMessage represents AI chatbot interaction.
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	BookingID string    `json:"booking_id,omitempty"`
	Role      string    `json:"role"` // user, assistant
	Content   string    `json:"content"`
	Lang      string    `json:"lang"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatRequest struct {
	Message   string `json:"message"`
	Lang      string `json:"lang"`
	BookingID string `json:"booking_id,omitempty"`
	Context   string `json:"context,omitempty"`
}

type ChatResponse struct {
	Reply            string   `json:"reply"`
	Sources          []string `json:"sources,omitempty"`
	SuggestedActions []string `json:"suggested_actions,omitempty"`
	Lang             string   `json:"lang"`
}

// RouteOptimization is the result of AI route optimization.
type RouteOptimization struct {
	BookingIDs    []string      `json:"booking_ids"`
	DriverID      string        `json:"driver_id"`
	OptimalOrder  []string      `json:"optimal_order"`
	TotalDistance  float64       `json:"total_distance_km"`
	TotalDuration int           `json:"total_duration_min"`
	Savings       float64       `json:"savings_percent"`
	Waypoints     []Waypoint    `json:"waypoints"`
}

type Waypoint struct {
	BookingID string  `json:"booking_id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Type      string  `json:"type"` // pickup, dropoff
	ETA       int     `json:"eta_min"`
}

// BiometricVerification represents driver selfie verification.
type BiometricVerification struct {
	DriverID   string  `json:"driver_id"`
	Verified   bool    `json:"verified"`
	Confidence float64 `json:"confidence"`
	Message    string  `json:"message"`
	Timestamp  int64   `json:"timestamp"`
}

type BiometricRequest struct {
	DriverID     string `json:"driver_id"`
	SelfieBase64 string `json:"selfie_base64"`
}
