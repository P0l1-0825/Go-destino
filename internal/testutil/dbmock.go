package testutil

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// NewMockDB creates a new sqlmock database connection for testing.
func NewMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// User columns returned by SELECT queries.
var UserColumns = []string{
	"id", "tenant_id", "email", "phone", "password_hash", "name", "role",
	"sub_role", "company_id", "lang", "active", "mfa_enabled",
	"created_at", "updated_at", "last_login",
}

// Booking columns returned by SELECT queries.
var BookingColumns = []string{
	"id", "booking_number", "tenant_id", "user_id", "kiosk_id", "route_id",
	"driver_id", "vehicle_id", "status", "service_type",
	"pickup_address", "dropoff_address", "pickup_lat", "pickup_lng",
	"dropoff_lat", "dropoff_lng", "passenger_count",
	"price_cents", "currency", "payment_id", "flight_number", "cancel_reason",
	"scheduled_at", "started_at", "completed_at", "cancelled_at", "created_at", "updated_at",
}

// Payment columns returned by SELECT queries.
var PaymentColumns = []string{
	"id", "tenant_id", "booking_id", "kiosk_id", "user_id",
	"method", "status", "amount_cents", "currency",
	"reference", "failure_reason", "refunded_at",
	"created_at", "updated_at",
}
