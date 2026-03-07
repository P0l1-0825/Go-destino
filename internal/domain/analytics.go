package domain

import "time"

// DashboardKPIs represents real-time KPIs for an airport/tenant.
type DashboardKPIs struct {
	TenantID         string  `json:"tenant_id"`
	AirportID        string  `json:"airport_id,omitempty"`
	Period           string  `json:"period"` // today, week, month
	TotalBookings    int     `json:"total_bookings"`
	ActiveBookings   int     `json:"active_bookings"`
	CompletedTrips   int     `json:"completed_trips"`
	CancelledTrips   int     `json:"cancelled_trips"`
	RevenueCents     int64   `json:"revenue_cents"`
	Currency         string  `json:"currency"`
	AvgRating        float64 `json:"avg_rating"`
	ActiveDrivers    int     `json:"active_drivers"`
	OnlineKiosks     int     `json:"online_kiosks"`
	AvgETAMinutes    float64 `json:"avg_eta_minutes"`
	ConversionRate   float64 `json:"conversion_rate"`   // bookings / estimates
	CancellationRate float64 `json:"cancellation_rate"`
	TicketsSold      int     `json:"tickets_sold"`
	CardRecharges    int     `json:"card_recharges"`
}

// RevenueReport represents revenue breakdown.
type RevenueReport struct {
	Period      string          `json:"period"`
	TotalCents  int64           `json:"total_cents"`
	Currency    string          `json:"currency"`
	ByService   []ServiceRevenue `json:"by_service"`
	ByPayMethod []MethodRevenue  `json:"by_pay_method"`
	ByDay       []DailyRevenue   `json:"by_day"`
}

type ServiceRevenue struct {
	ServiceType string `json:"service_type"`
	TotalCents  int64  `json:"total_cents"`
	Count       int    `json:"count"`
}

type MethodRevenue struct {
	Method     string `json:"method"`
	TotalCents int64  `json:"total_cents"`
	Count      int    `json:"count"`
}

type DailyRevenue struct {
	Date       string `json:"date"`
	TotalCents int64  `json:"total_cents"`
	Count      int    `json:"count"`
}

// BookingFunnel tracks conversion from search to completed trip.
type BookingFunnel struct {
	Period     string `json:"period"`
	Searches   int    `json:"searches"`
	Estimates  int    `json:"estimates"`
	Created    int    `json:"created"`
	Confirmed  int    `json:"confirmed"`
	Completed  int    `json:"completed"`
	Cancelled  int    `json:"cancelled"`
}

// DriverPerformance represents a driver's metrics.
type DriverPerformance struct {
	DriverID       string  `json:"driver_id"`
	DriverName     string  `json:"driver_name"`
	TotalTrips     int     `json:"total_trips"`
	CompletedTrips int     `json:"completed_trips"`
	CancelledTrips int     `json:"cancelled_trips"`
	AvgRating      float64 `json:"avg_rating"`
	TotalRevenue   int64   `json:"total_revenue_cents"`
	AvgTripMinutes float64 `json:"avg_trip_minutes"`
	OnlineHours    float64 `json:"online_hours"`
}

// SLOMetrics represents service level objective metrics.
type SLOMetrics struct {
	Service       string  `json:"service"`
	UptimePercent float64 `json:"uptime_percent"`
	P50LatencyMS  int     `json:"p50_latency_ms"`
	P99LatencyMS  int     `json:"p99_latency_ms"`
	ErrorRate     float64 `json:"error_rate"`
	ErrorBudget   float64 `json:"error_budget_remaining"`
}

// AuditLogEntry represents an auditable action.
type AuditLogEntry struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	ResourceID string   `json:"resource_id" db:"resource_id"`
	Details   string    `json:"details" db:"details"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ShiftRecord represents a POS seller work shift.
type ShiftRecord struct {
	ID             string     `json:"id" db:"id"`
	TenantID       string     `json:"tenant_id" db:"tenant_id"`
	SellerID       string     `json:"seller_id" db:"seller_id"`
	AirportID      string     `json:"airport_id" db:"airport_id"`
	TerminalID     string     `json:"terminal_id" db:"terminal_id"`
	KioskID        string     `json:"kiosk_id" db:"kiosk_id"`
	Status         string     `json:"status" db:"status"` // open, closed
	OpenedAt       time.Time  `json:"opened_at" db:"opened_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	TotalSales     int64      `json:"total_sales_cents" db:"total_sales_cents"`
	CashCollected  int64      `json:"cash_collected_cents" db:"cash_collected_cents"`
	CardCollected  int64      `json:"card_collected_cents" db:"card_collected_cents"`
	TicketsSold    int        `json:"tickets_sold" db:"tickets_sold"`
	BookingsCreated int       `json:"bookings_created" db:"bookings_created"`
	CommissionCents int64     `json:"commission_cents" db:"commission_cents"`
}

// Commission represents earned commission per sale.
type Commission struct {
	ID         string    `json:"id" db:"id"`
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	SellerID   string    `json:"seller_id" db:"seller_id"`
	BookingID  string    `json:"booking_id,omitempty" db:"booking_id"`
	TicketID   string    `json:"ticket_id,omitempty" db:"ticket_id"`
	AmountCents int64    `json:"amount_cents" db:"amount_cents"`
	Currency   string    `json:"currency" db:"currency"`
	Status     string    `json:"status" db:"status"` // pending, paid
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
