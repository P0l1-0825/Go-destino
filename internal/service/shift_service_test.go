package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

// --- Shift mocks ---

type mockShiftRepo struct {
	CreateFn       func(ctx context.Context, s *domain.ShiftRecord) error
	GetByIDFn      func(ctx context.Context, id string) (*domain.ShiftRecord, error)
	GetActiveFn    func(ctx context.Context, sellerID string) (*domain.ShiftRecord, error)
	CloseFn        func(ctx context.Context, id string, totalSales, cashCollected, cardCollected, commissionCents int64, ticketsSold, bookingsCreated int) error
	ListBySellerFn func(ctx context.Context, sellerID string, limit int) ([]domain.ShiftRecord, error)
	ListByKioskFn  func(ctx context.Context, kioskID string, limit int) ([]domain.ShiftRecord, error)
}

func (m *mockShiftRepo) Create(ctx context.Context, s *domain.ShiftRecord) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, s) }
	return nil
}
func (m *mockShiftRepo) GetByID(ctx context.Context, id string) (*domain.ShiftRecord, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(ctx, id) }
	return nil, fmt.Errorf("not found")
}
func (m *mockShiftRepo) GetActive(ctx context.Context, sellerID string) (*domain.ShiftRecord, error) {
	if m.GetActiveFn != nil { return m.GetActiveFn(ctx, sellerID) }
	return nil, fmt.Errorf("no active shift")
}
func (m *mockShiftRepo) Close(ctx context.Context, id string, totalSales, cashCollected, cardCollected, commissionCents int64, ticketsSold, bookingsCreated int) error {
	if m.CloseFn != nil { return m.CloseFn(ctx, id, totalSales, cashCollected, cardCollected, commissionCents, ticketsSold, bookingsCreated) }
	return nil
}
func (m *mockShiftRepo) ListBySeller(ctx context.Context, sellerID string, limit int) ([]domain.ShiftRecord, error) {
	if m.ListBySellerFn != nil { return m.ListBySellerFn(ctx, sellerID, limit) }
	return nil, nil
}
func (m *mockShiftRepo) ListByKiosk(ctx context.Context, kioskID string, limit int) ([]domain.ShiftRecord, error) {
	if m.ListByKioskFn != nil { return m.ListByKioskFn(ctx, kioskID, limit) }
	return nil, nil
}

// --- Tests ---

func TestShiftService_OpenShift(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		svc := NewShiftService(&mockShiftRepo{
			GetActiveFn: func(_ context.Context, _ string) (*domain.ShiftRecord, error) {
				return nil, fmt.Errorf("no active shift")
			},
		})
		shift, err := svc.OpenShift(context.Background(), testutil.TestTenantID, "seller1", "airport1", "T1", testutil.TestKioskID)
		testutil.AssertNoError(t, err)
		if shift.TenantID != testutil.TestTenantID { t.Errorf("tenant = %s", shift.TenantID) }
	})

	t.Run("already has open shift", func(t *testing.T) {
		svc := NewShiftService(&mockShiftRepo{
			GetActiveFn: func(_ context.Context, _ string) (*domain.ShiftRecord, error) {
				return &domain.ShiftRecord{ID: "existing", Status: "open"}, nil
			},
		})
		_, err := svc.OpenShift(context.Background(), testutil.TestTenantID, "seller1", "airport1", "T1", testutil.TestKioskID)
		testutil.AssertError(t, err, "already has an open shift")
	})
}

func TestShiftService_CloseShift(t *testing.T) {
	tests := []struct {
		name    string
		shift   *domain.ShiftRecord
		total   int64
		cash    int64
		card    int64
		wantErr string
	}{
		{
			name:  "happy path",
			shift: &domain.ShiftRecord{ID: "s1", Status: "open"},
			total: 10000, cash: 6000, card: 4000,
		},
		{
			name:    "already closed",
			shift:   &domain.ShiftRecord{ID: "s2", Status: "closed"},
			total:   10000, cash: 6000, card: 4000,
			wantErr: "already closed",
		},
		{
			name:    "totals mismatch",
			shift:   &domain.ShiftRecord{ID: "s3", Status: "open"},
			total:   10000, cash: 5000, card: 4000,
			wantErr: "must equal total sales",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewShiftService(&mockShiftRepo{
				GetByIDFn: func(_ context.Context, _ string) (*domain.ShiftRecord, error) { return tt.shift, nil },
				CloseFn: func(_ context.Context, _ string, _, _, _, commissionCents int64, _, _ int) error {
					expected := int64(float64(tt.total) * 0.05)
					if commissionCents != expected {
						t.Errorf("commission = %d, want %d", commissionCents, expected)
					}
					return nil
				},
			})
			err := svc.CloseShift(context.Background(), tt.shift.ID, tt.total, tt.cash, tt.card, 10, 5)
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
		})
	}
}

func TestShiftService_ListShifts_LimitCap(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		wantLimit int
	}{
		{"default", 0, 30},
		{"custom", 50, 50},
		{"cap", 200, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewShiftService(&mockShiftRepo{
				ListBySellerFn: func(_ context.Context, _ string, limit int) ([]domain.ShiftRecord, error) {
					if limit != tt.wantLimit { t.Errorf("limit = %d, want %d", limit, tt.wantLimit) }
					return nil, nil
				},
			})
			_, _ = svc.ListShifts(context.Background(), "seller1", tt.input)
		})
	}
}
