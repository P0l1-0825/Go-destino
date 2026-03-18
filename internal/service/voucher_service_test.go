package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

// --- Voucher mocks ---

type mockVoucherRepo struct {
	CreateFn         func(ctx context.Context, v *domain.Voucher) error
	GetByIDFn        func(ctx context.Context, id string) (*domain.Voucher, error)
	GetByCodeFn      func(ctx context.Context, code string) (*domain.Voucher, error)
	GetByCodeTenantFn func(ctx context.Context, code, tenantID string) (*domain.Voucher, error)
	RedeemFn         func(ctx context.Context, id, redeemedBy string) error
	ListFn           func(ctx context.Context, tenantID string, limit, offset int) ([]domain.Voucher, error)
}

func (m *mockVoucherRepo) Create(ctx context.Context, v *domain.Voucher) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, v) }
	return nil
}
func (m *mockVoucherRepo) GetByID(ctx context.Context, id string) (*domain.Voucher, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(ctx, id) }
	return nil, nil
}
func (m *mockVoucherRepo) GetByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	if m.GetByCodeFn != nil { return m.GetByCodeFn(ctx, code) }
	return nil, nil
}
func (m *mockVoucherRepo) GetByCodeTenant(ctx context.Context, code, tenantID string) (*domain.Voucher, error) {
	if m.GetByCodeTenantFn != nil { return m.GetByCodeTenantFn(ctx, code, tenantID) }
	return nil, nil
}
func (m *mockVoucherRepo) Redeem(ctx context.Context, id, redeemedBy string) error {
	if m.RedeemFn != nil { return m.RedeemFn(ctx, id, redeemedBy) }
	return nil
}
func (m *mockVoucherRepo) List(ctx context.Context, tenantID string, limit, offset int) ([]domain.Voucher, error) {
	if m.ListFn != nil { return m.ListFn(ctx, tenantID, limit, offset) }
	return nil, nil
}

type mockPaymentCreator struct {
	CreateFn func(ctx context.Context, p *domain.Payment) error
}

func (m *mockPaymentCreator) Create(ctx context.Context, p *domain.Payment) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, p) }
	return nil
}

// --- Tests ---

func TestVoucherService_Create(t *testing.T) {
	svc := NewVoucherService(
		&mockVoucherRepo{CreateFn: func(_ context.Context, v *domain.Voucher) error {
			if v.Status != domain.VoucherActive { t.Errorf("status = %s, want active", v.Status) }
			if v.Code == "" { t.Error("expected non-empty code") }
			return nil
		}},
		&mockPaymentCreator{},
	)

	voucher, err := svc.Create(context.Background(), testutil.TestTenantID, testutil.TestUserID,
		domain.CreateVoucherRequest{BookingID: testutil.TestBookingID, AmountCents: 15000, Currency: "MXN"})
	testutil.AssertNoError(t, err)
	if voucher.TenantID != testutil.TestTenantID { t.Errorf("tenant = %s", voucher.TenantID) }
	if voucher.AmountCents != 15000 { t.Errorf("amount = %d", voucher.AmountCents) }
}

func TestVoucherService_Redeem(t *testing.T) {
	tests := []struct {
		name      string
		voucher   *domain.Voucher
		amount    int64
		wantErr   string
	}{
		{
			name: "happy path",
			voucher: &domain.Voucher{ID: "v1", Status: domain.VoucherActive, AmountCents: 10000, Currency: "MXN", ExpiresAt: time.Now().Add(1 * time.Hour)},
			amount: 10000,
		},
		{
			name: "change returned",
			voucher: &domain.Voucher{ID: "v2", Status: domain.VoucherActive, AmountCents: 8000, Currency: "MXN", ExpiresAt: time.Now().Add(1 * time.Hour)},
			amount: 10000,
		},
		{
			name:    "already redeemed",
			voucher: &domain.Voucher{ID: "v3", Status: domain.VoucherRedeemed, AmountCents: 10000, ExpiresAt: time.Now().Add(1 * time.Hour)},
			amount:  10000,
			wantErr: "voucher is redeemed",
		},
		{
			name:    "expired",
			voucher: &domain.Voucher{ID: "v4", Status: domain.VoucherActive, AmountCents: 10000, ExpiresAt: time.Now().Add(-1 * time.Hour)},
			amount:  10000,
			wantErr: "voucher has expired",
		},
		{
			name:    "insufficient cash",
			voucher: &domain.Voucher{ID: "v5", Status: domain.VoucherActive, AmountCents: 10000, ExpiresAt: time.Now().Add(1 * time.Hour)},
			amount:  5000,
			wantErr: "insufficient cash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewVoucherService(
				&mockVoucherRepo{
					GetByCodeFn: func(_ context.Context, _ string) (*domain.Voucher, error) { return tt.voucher, nil },
					RedeemFn: func(_ context.Context, _, _ string) error { return nil },
				},
				&mockPaymentCreator{},
			)

			resp, err := svc.Redeem(context.Background(), testutil.TestTenantID, testutil.TestUserID,
				domain.RedeemVoucherRequest{Code: "V-test", AmountReceived: tt.amount})
			if tt.wantErr != "" {
				testutil.AssertError(t, err, tt.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			if resp.Change != tt.amount-tt.voucher.AmountCents {
				t.Errorf("change = %d, want %d", resp.Change, tt.amount-tt.voucher.AmountCents)
			}
		})
	}
}

func TestVoucherService_Redeem_NotFound(t *testing.T) {
	svc := NewVoucherService(
		&mockVoucherRepo{
			GetByCodeFn: func(_ context.Context, _ string) (*domain.Voucher, error) { return nil, fmt.Errorf("not found") },
		},
		&mockPaymentCreator{},
	)
	_, err := svc.Redeem(context.Background(), testutil.TestTenantID, testutil.TestUserID,
		domain.RedeemVoucherRequest{Code: "INVALID"})
	testutil.AssertError(t, err, "voucher not found")
}

func TestVoucherService_List_DefaultLimit(t *testing.T) {
	called := false
	svc := NewVoucherService(
		&mockVoucherRepo{
			ListFn: func(_ context.Context, _ string, limit, _ int) ([]domain.Voucher, error) {
				called = true
				if limit != 50 { t.Errorf("limit = %d, want 50", limit) }
				return nil, nil
			},
		},
		&mockPaymentCreator{},
	)
	_, _ = svc.List(context.Background(), testutil.TestTenantID, 0, 0)
	if !called { t.Error("ListFn not called") }
}
