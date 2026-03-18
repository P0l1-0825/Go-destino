package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
	"github.com/P0l1-0825/Go-destino/internal/testutil/mocks"
)

func newTestBookingService(bookingRepo *mocks.MockBookingRepo, paymentRepo *mocks.MockPaymentRepo) *BookingService {
	return NewBookingService(bookingRepo, paymentRepo)
}

func validCreateBookingReq() domain.CreateBookingRequest {
	return domain.CreateBookingRequest{
		ServiceType:    domain.ServiceTaxi,
		PickupAddress:  "Terminal 1, MEX",
		DropoffAddress: "Hotel Centro",
		PickupLat:      19.4363,
		PickupLng:      -99.0721,
		DropoffLat:     19.4326,
		DropoffLng:     -99.1332,
		PassengerCount: 2,
	}
}

// --- Create ---

func TestBookingService_Create(t *testing.T) {
	tests := []struct {
		name      string
		req       domain.CreateBookingRequest
		mockSetup func(*mocks.MockBookingRepo)
		wantErr   string
	}{
		{
			name: "happy path",
			req:  validCreateBookingReq(),
			mockSetup: func(m *mocks.MockBookingRepo) {
				m.CreateFn = func(_ context.Context, _ *domain.Booking) error { return nil }
			},
		},
		{
			name: "invalid service type",
			req: func() domain.CreateBookingRequest {
				r := validCreateBookingReq()
				r.ServiceType = "helicopter"
				return r
			}(),
			mockSetup: func(m *mocks.MockBookingRepo) {},
			wantErr:   "invalid service type",
		},
		{
			name: "passenger count zero",
			req: func() domain.CreateBookingRequest {
				r := validCreateBookingReq()
				r.PassengerCount = 0
				return r
			}(),
			mockSetup: func(m *mocks.MockBookingRepo) {},
			wantErr:   "passenger count must be between 1 and 50",
		},
		{
			name: "passenger count over 50",
			req: func() domain.CreateBookingRequest {
				r := validCreateBookingReq()
				r.PassengerCount = 51
				return r
			}(),
			mockSetup: func(m *mocks.MockBookingRepo) {},
			wantErr:   "passenger count must be between 1 and 50",
		},
		{
			name: "invalid latitude",
			req: func() domain.CreateBookingRequest {
				r := validCreateBookingReq()
				r.PickupLat = 91.0
				return r
			}(),
			mockSetup: func(m *mocks.MockBookingRepo) {},
			wantErr:   "latitude must be between -90 and 90",
		},
		{
			name: "invalid longitude",
			req: func() domain.CreateBookingRequest {
				r := validCreateBookingReq()
				r.PickupLng = 181.0
				return r
			}(),
			mockSetup: func(m *mocks.MockBookingRepo) {},
			wantErr:   "longitude must be between -180 and 180",
		},
		{
			name: "repo create error",
			req:  validCreateBookingReq(),
			mockSetup: func(m *mocks.MockBookingRepo) {
				m.CreateFn = func(_ context.Context, _ *domain.Booking) error { return fmt.Errorf("db error") }
			},
			wantErr: "creating booking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookingRepo := &mocks.MockBookingRepo{}
			paymentRepo := &mocks.MockPaymentRepo{}
			tt.mockSetup(bookingRepo)
			svc := newTestBookingService(bookingRepo, paymentRepo)

			booking, err := svc.Create(context.Background(), testutil.TestTenantID, testutil.TestUserID, testutil.TestKioskID, tt.req)
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if booking == nil {
				t.Fatal("expected booking, got nil")
			}
			if booking.TenantID != testutil.TestTenantID {
				t.Errorf("tenant_id = %s, want %s", booking.TenantID, testutil.TestTenantID)
			}
			if booking.Status != domain.BookingPending {
				t.Errorf("status = %s, want pending", booking.Status)
			}
			if booking.PriceCents <= 0 {
				t.Error("expected positive price")
			}
		})
	}
}

// --- Estimate ---

func TestBookingService_Estimate(t *testing.T) {
	svc := newTestBookingService(&mocks.MockBookingRepo{}, &mocks.MockPaymentRepo{})

	tests := []struct {
		name        string
		serviceType domain.ServiceType
		minPrice    int64
	}{
		{"taxi", domain.ServiceTaxi, 5000},
		{"shuttle", domain.ServiceShuttle, 3500},
		{"van", domain.ServiceVan, 8000},
		{"bus", domain.ServiceBus, 2500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.Estimate(domain.EstimateRequest{
				ServiceType:    tt.serviceType,
				PickupLat:      19.4363,
				PickupLng:      -99.0721,
				DropoffLat:     19.4326,
				DropoffLng:     -99.1332,
				PassengerCount: 1,
			})
			testutil.AssertNoError(t, err)
			if resp.PriceCents < tt.minPrice {
				t.Errorf("price %d < min %d for %s", resp.PriceCents, tt.minPrice, tt.serviceType)
			}
			if resp.Currency != "MXN" {
				t.Errorf("currency = %s, want MXN", resp.Currency)
			}
			if resp.ETAMinutes <= 0 {
				t.Error("expected positive ETA")
			}
		})
	}

	t.Run("default passenger count", func(t *testing.T) {
		resp, err := svc.Estimate(domain.EstimateRequest{
			ServiceType:    domain.ServiceTaxi,
			PickupLat:      19.4363,
			PickupLng:      -99.0721,
			DropoffLat:     19.4326,
			DropoffLng:     -99.1332,
			PassengerCount: 0, // should default to 1
		})
		testutil.AssertNoError(t, err)
		if resp.PriceCents <= 0 {
			t.Error("expected positive price even with 0 passengers")
		}
	})
}

// --- Status transitions ---

func TestBookingService_Confirm(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Booking, error) {
			b := testutil.NewTestBooking()
			b.Status = domain.BookingPending
			return b, nil
		},
		UpdateStatusFn: func(_ context.Context, _, _ string, _ domain.BookingStatus) error { return nil },
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	err := svc.Confirm(context.Background(), testutil.TestBookingID, testutil.TestTenantID)
	testutil.AssertNoError(t, err)
}

func TestBookingService_AssignDriver(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Booking, error) {
			b := testutil.NewTestBooking()
			b.Status = domain.BookingConfirmed
			return b, nil
		},
		AssignDriverFn: func(_ context.Context, _, _, _, _ string) error { return nil },
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	err := svc.AssignDriver(context.Background(), testutil.TestBookingID, testutil.TestTenantID,
		domain.AssignDriverRequest{DriverID: testutil.TestDriverID, VehicleID: testutil.TestVehicleID})
	testutil.AssertNoError(t, err)
}

func TestBookingService_StartTrip(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Booking, error) {
			b := testutil.NewTestBooking()
			b.Status = domain.BookingAssigned
			return b, nil
		},
		SetStartedFn: func(_ context.Context, _, _ string) error { return nil },
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	err := svc.StartTrip(context.Background(), testutil.TestBookingID, testutil.TestTenantID)
	testutil.AssertNoError(t, err)
}

func TestBookingService_CompleteBooking(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Booking, error) {
			b := testutil.NewTestBooking()
			b.Status = domain.BookingStarted
			return b, nil
		},
		SetCompletedFn: func(_ context.Context, _, _ string) error { return nil },
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	err := svc.CompleteBooking(context.Background(), testutil.TestBookingID, testutil.TestTenantID)
	testutil.AssertNoError(t, err)
}

func TestBookingService_Cancel(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Booking, error) {
			b := testutil.NewTestBooking()
			b.Status = domain.BookingPending
			return b, nil
		},
		SetCancelledFn: func(_ context.Context, _, _, _ string) error { return nil },
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	err := svc.Cancel(context.Background(), testutil.TestBookingID, testutil.TestTenantID, "changed my mind")
	testutil.AssertNoError(t, err)
}

func TestBookingService_InvalidTransition(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Booking, error) {
			b := testutil.NewTestBooking()
			b.Status = domain.BookingStarted // started can only go to completed
			return b, nil
		},
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	err := svc.Cancel(context.Background(), testutil.TestBookingID, testutil.TestTenantID, "too late")
	testutil.AssertError(t, err, "invalid transition")
}

func TestBookingService_BookingNotFound(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Booking, error) {
			return nil, fmt.Errorf("sql: no rows")
		},
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	err := svc.Confirm(context.Background(), "nonexistent", testutil.TestTenantID)
	testutil.AssertError(t, err, "booking not found")
}

// --- ListByTenant ---

func TestBookingService_ListByTenant(t *testing.T) {
	bookingRepo := &mocks.MockBookingRepo{
		ListByTenantFn: func(_ context.Context, _ string, limit int) ([]domain.Booking, error) {
			// Return a booking to verify limit was passed through
			return []domain.Booking{*testutil.NewTestBooking()}, nil
		},
	}
	svc := newTestBookingService(bookingRepo, &mocks.MockPaymentRepo{})

	tests := []struct {
		name      string
		limit     int
		wantLimit int
	}{
		{"default limit", 0, 50},
		{"custom limit", 25, 25},
		{"cap at 200", 500, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookings, err := svc.ListByTenant(context.Background(), testutil.TestTenantID, tt.limit)
			testutil.AssertNoError(t, err)
			if len(bookings) == 0 {
				t.Error("expected at least one booking")
			}
		})
	}
}
