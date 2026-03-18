package service

import (
	"context"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

// --- Kiosk mocks ---

type mockKioskRepo struct {
	CreateFn          func(ctx context.Context, k *domain.Kiosk) error
	GetByIDFn         func(ctx context.Context, id string) (*domain.Kiosk, error)
	UpdateHeartbeatFn func(ctx context.Context, id string) error
	UpdateStatusFn    func(ctx context.Context, id string, status domain.KioskStatus) error
	ListByTenantFn    func(ctx context.Context, tenantID string) ([]domain.Kiosk, error)
}

func (m *mockKioskRepo) Create(ctx context.Context, k *domain.Kiosk) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, k) }
	return nil
}
func (m *mockKioskRepo) GetByID(ctx context.Context, id string) (*domain.Kiosk, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(ctx, id) }
	return nil, nil
}
func (m *mockKioskRepo) UpdateHeartbeat(ctx context.Context, id string) error {
	if m.UpdateHeartbeatFn != nil { return m.UpdateHeartbeatFn(ctx, id) }
	return nil
}
func (m *mockKioskRepo) UpdateStatus(ctx context.Context, id string, status domain.KioskStatus) error {
	if m.UpdateStatusFn != nil { return m.UpdateStatusFn(ctx, id, status) }
	return nil
}
func (m *mockKioskRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Kiosk, error) {
	if m.ListByTenantFn != nil { return m.ListByTenantFn(ctx, tenantID) }
	return nil, nil
}

// --- Tests ---

func TestKioskService_Register(t *testing.T) {
	svc := NewKioskService(&mockKioskRepo{})
	kiosk, err := svc.Register(context.Background(), testutil.TestTenantID, domain.RegisterKioskRequest{
		Name: "Kiosk Terminal 1", Location: "Gate A12", AirportID: "MEX", TerminalID: "T1",
	})
	testutil.AssertNoError(t, err)
	if kiosk.TenantID != testutil.TestTenantID { t.Errorf("tenant = %s", kiosk.TenantID) }
	if kiosk.Status != domain.KioskOnline { t.Errorf("status = %s, want online", kiosk.Status) }
	if kiosk.ID == "" { t.Error("expected non-empty ID") }
}

func TestKioskService_Heartbeat(t *testing.T) {
	called := false
	svc := NewKioskService(&mockKioskRepo{
		UpdateHeartbeatFn: func(_ context.Context, id string) error {
			called = true
			if id != testutil.TestKioskID { t.Errorf("id = %s", id) }
			return nil
		},
	})
	err := svc.Heartbeat(context.Background(), testutil.TestKioskID)
	testutil.AssertNoError(t, err)
	if !called { t.Error("UpdateHeartbeat not called") }
}

func TestKioskService_UpdateStatus(t *testing.T) {
	svc := NewKioskService(&mockKioskRepo{
		UpdateStatusFn: func(_ context.Context, _ string, status domain.KioskStatus) error {
			if status != domain.KioskMaintenance { t.Errorf("status = %s", status) }
			return nil
		},
	})
	err := svc.UpdateStatus(context.Background(), testutil.TestKioskID, domain.KioskMaintenance)
	testutil.AssertNoError(t, err)
}

func TestKioskService_ListByTenant(t *testing.T) {
	svc := NewKioskService(&mockKioskRepo{
		ListByTenantFn: func(_ context.Context, tenantID string) ([]domain.Kiosk, error) {
			if tenantID != testutil.TestTenantID { t.Errorf("tenant = %s", tenantID) }
			return []domain.Kiosk{{ID: "k1", TenantID: testutil.TestTenantID}}, nil
		},
	})
	kiosks, err := svc.ListByTenant(context.Background(), testutil.TestTenantID)
	testutil.AssertNoError(t, err)
	if len(kiosks) != 1 { t.Fatalf("expected 1, got %d", len(kiosks)) }
}
