package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

func TestBookingRepo_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewBookingRepository(db)

	b := testutil.NewTestBooking()

	mock.ExpectExec("INSERT INTO bookings").
		WithArgs(b.ID, b.BookingNumber, b.TenantID, b.UserID, b.KioskID, b.RouteID,
			b.DriverID, b.VehicleID, b.Status, b.ServiceType,
			b.PickupAddress, b.DropoffAddress, b.PickupLat, b.PickupLng,
			b.DropoffLat, b.DropoffLng, b.PassengerCount,
			b.PriceCents, b.Currency, b.PaymentID, b.FlightNumber, b.ScheduledAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), b)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBookingRepo_GetByIDTenant_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewBookingRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(testutil.BookingColumns).
		AddRow(testutil.TestBookingID, "GD-TEST0001", testutil.TestTenantID, testutil.TestUserID,
			testutil.TestKioskID, "", "", "", "pending", "taxi",
			"Terminal 1", "Hotel", 19.43, -99.07, 19.43, -99.13, 2,
			15000, "MXN", "", "", "",
			nil, nil, nil, nil, now, now)

	mock.ExpectQuery("SELECT .+ FROM bookings WHERE id = \\$1 AND tenant_id = \\$2").
		WithArgs(testutil.TestBookingID, testutil.TestTenantID).
		WillReturnRows(rows)

	booking, err := repo.GetByIDTenant(context.Background(), testutil.TestBookingID, testutil.TestTenantID)
	if err != nil {
		t.Fatalf("GetByIDTenant: %v", err)
	}
	if booking.TenantID != testutil.TestTenantID {
		t.Errorf("tenant_id = %s, want %s", booking.TenantID, testutil.TestTenantID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBookingRepo_UpdateStatus_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewBookingRepository(db)

	mock.ExpectExec("UPDATE bookings SET status = \\$1.+WHERE id = \\$2 AND tenant_id = \\$3").
		WithArgs(domain.BookingConfirmed, testutil.TestBookingID, testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(context.Background(), testutil.TestBookingID, testutil.TestTenantID, domain.BookingConfirmed)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBookingRepo_UpdateStatus_NotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewBookingRepository(db)

	mock.ExpectExec("UPDATE bookings SET status").
		WithArgs(domain.BookingConfirmed, "nonexistent", testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := repo.UpdateStatus(context.Background(), "nonexistent", testutil.TestTenantID, domain.BookingConfirmed)
	if err == nil {
		t.Fatal("expected error for 0 rows affected")
	}
}

func TestBookingRepo_AssignDriver_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewBookingRepository(db)

	mock.ExpectExec("UPDATE bookings SET driver_id = \\$1, vehicle_id = \\$2.+WHERE id = \\$3 AND tenant_id = \\$4").
		WithArgs(testutil.TestDriverID, testutil.TestVehicleID, testutil.TestBookingID, testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AssignDriver(context.Background(), testutil.TestBookingID, testutil.TestTenantID, testutil.TestDriverID, testutil.TestVehicleID)
	if err != nil {
		t.Fatalf("AssignDriver: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBookingRepo_SetCancelled_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewBookingRepository(db)

	mock.ExpectExec("UPDATE bookings SET status = 'cancelled', cancel_reason = \\$1.+WHERE id = \\$2 AND tenant_id = \\$3").
		WithArgs("changed mind", testutil.TestBookingID, testutil.TestTenantID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetCancelled(context.Background(), testutil.TestBookingID, testutil.TestTenantID, "changed mind")
	if err != nil {
		t.Fatalf("SetCancelled: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBookingRepo_ListByTenant_TenantIsolation(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := NewBookingRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows(testutil.BookingColumns).
		AddRow(testutil.TestBookingID, "GD-TEST0001", testutil.TestTenantID, testutil.TestUserID,
			testutil.TestKioskID, "", "", "", "pending", "taxi",
			"Terminal 1", "Hotel", 19.43, -99.07, 19.43, -99.13, 2,
			15000, "MXN", "", "", "",
			nil, nil, nil, nil, now, now)

	mock.ExpectQuery("SELECT .+ FROM bookings WHERE tenant_id = \\$1").
		WithArgs(testutil.TestTenantID, 50).
		WillReturnRows(rows)

	bookings, err := repo.ListByTenant(context.Background(), testutil.TestTenantID, 50)
	if err != nil {
		t.Fatalf("ListByTenant: %v", err)
	}
	if len(bookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(bookings))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
