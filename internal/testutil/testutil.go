package testutil

import (
	"strings"
	"testing"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/config"
	"github.com/P0l1-0825/Go-destino/internal/domain"
)

// Standard test constants reused across all test files.
const (
	TestTenantID  = "00000000-0000-0000-0000-000000000001"
	TestUserID    = "00000000-0000-0000-0000-aaaaaaaaaaaa"
	TestDriverID  = "00000000-0000-0000-0000-dddddddddddd"
	TestVehicleID = "00000000-0000-0000-0000-vvvvvvvvvvvv"
	TestBookingID = "00000000-0000-0000-0000-bbbbbbbbbbbb"
	TestPaymentID = "00000000-0000-0000-0000-pppppppppppp"
	TestKioskID   = "00000000-0000-0000-0000-kkkkkkkkkkkk"
	TestEmail     = "test@godestino.com"
	TestJWTSecret = "test-jwt-secret-must-be-at-least-32-characters-long"
)

// NewTestJWTConfig returns a JWTConfig suitable for unit tests.
func NewTestJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:     TestJWTSecret,
		ExpireHour: 1,
	}
}

// NewTestUser returns a User with sensible test defaults.
func NewTestUser() *domain.User {
	now := time.Now()
	return &domain.User{
		ID:           TestUserID,
		TenantID:     TestTenantID,
		Email:        TestEmail,
		Phone:        "+525512345678",
		PasswordHash: "$2a$12$LJ3m4ys1Q1K7K7K7K7K7K7K7K7K7K7K7K7K7K7K7K7K7K7K7K7K", // dummy bcrypt
		Name:         "Test User",
		Role:         domain.RoleUsuario,
		Lang:         "es",
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewTestBooking returns a Booking with sensible test defaults.
func NewTestBooking() *domain.Booking {
	now := time.Now()
	return &domain.Booking{
		ID:             TestBookingID,
		BookingNumber:  "GD-TEST0001",
		TenantID:       TestTenantID,
		UserID:         TestUserID,
		KioskID:        TestKioskID,
		Status:         domain.BookingPending,
		ServiceType:    domain.ServiceTaxi,
		PickupAddress:  "Terminal 1, MEX",
		DropoffAddress: "Hotel Centro, CDMX",
		PickupLat:      19.4363,
		PickupLng:      -99.0721,
		DropoffLat:     19.4326,
		DropoffLng:     -99.1332,
		PassengerCount: 2,
		PriceCents:     15000,
		Currency:       "MXN",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewTestPayment returns a Payment with sensible test defaults.
func NewTestPayment() *domain.Payment {
	now := time.Now()
	return &domain.Payment{
		ID:          TestPaymentID,
		TenantID:    TestTenantID,
		BookingID:   TestBookingID,
		KioskID:     TestKioskID,
		UserID:      TestUserID,
		Method:      domain.PaymentCard,
		Status:      domain.PaymentCompleted,
		AmountCents: 15000,
		Currency:    "MXN",
		Reference:   "PAY-test123456",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AssertError checks that err is non-nil and contains msg.
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", msg)
	}
	if !strings.Contains(err.Error(), msg) {
		t.Fatalf("expected error containing %q, got %q", msg, err.Error())
	}
}

// AssertNoError fails the test if err is non-nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
