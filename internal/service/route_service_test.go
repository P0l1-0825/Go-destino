package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

// --- Route mocks ---

type mockRouteRepo struct {
	CreateFn              func(ctx context.Context, r *domain.Route) error
	GetByIDFn             func(ctx context.Context, id string) (*domain.Route, error)
	UpdateFn              func(ctx context.Context, r *domain.Route) error
	DeactivateFn          func(ctx context.Context, id, tenantID string) error
	ListByTenantFn        func(ctx context.Context, tenantID string) ([]domain.Route, error)
	ListByTransportTypeFn func(ctx context.Context, tenantID string, t domain.TransportType) ([]domain.Route, error)
}

func (m *mockRouteRepo) Create(ctx context.Context, r *domain.Route) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, r) }
	return nil
}
func (m *mockRouteRepo) GetByID(ctx context.Context, id string) (*domain.Route, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(ctx, id) }
	return nil, fmt.Errorf("not found")
}
func (m *mockRouteRepo) Update(ctx context.Context, r *domain.Route) error {
	if m.UpdateFn != nil { return m.UpdateFn(ctx, r) }
	return nil
}
func (m *mockRouteRepo) Deactivate(ctx context.Context, id, tenantID string) error {
	if m.DeactivateFn != nil { return m.DeactivateFn(ctx, id, tenantID) }
	return nil
}
func (m *mockRouteRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Route, error) {
	if m.ListByTenantFn != nil { return m.ListByTenantFn(ctx, tenantID) }
	return nil, nil
}
func (m *mockRouteRepo) ListByTransportType(ctx context.Context, tenantID string, t domain.TransportType) ([]domain.Route, error) {
	if m.ListByTransportTypeFn != nil { return m.ListByTransportTypeFn(ctx, tenantID, t) }
	return nil, nil
}

// --- Tests ---

func TestRouteService_Create(t *testing.T) {
	svc := NewRouteService(&mockRouteRepo{})
	route, err := svc.Create(context.Background(), testutil.TestTenantID, domain.CreateRouteRequest{
		Name: "MEX → Centro", Code: "MEX-CTR", TransportType: "taxi",
		Origin: "MEX Airport", Destination: "Centro CDMX", PriceCents: 50000, Currency: "MXN",
	})
	testutil.AssertNoError(t, err)
	if route.TenantID != testutil.TestTenantID { t.Errorf("tenant = %s", route.TenantID) }
	if !route.Active { t.Error("route should be active") }
}

func TestRouteService_Update_TenantIsolation(t *testing.T) {
	t.Run("same tenant", func(t *testing.T) {
		svc := NewRouteService(&mockRouteRepo{
			GetByIDFn: func(_ context.Context, _ string) (*domain.Route, error) {
				return &domain.Route{ID: "r1", TenantID: testutil.TestTenantID}, nil
			},
		})
		_, err := svc.Update(context.Background(), "r1", testutil.TestTenantID, domain.CreateRouteRequest{Name: "Updated"})
		testutil.AssertNoError(t, err)
	})

	t.Run("cross-tenant blocked", func(t *testing.T) {
		svc := NewRouteService(&mockRouteRepo{
			GetByIDFn: func(_ context.Context, _ string) (*domain.Route, error) {
				return &domain.Route{ID: "r1", TenantID: "other-tenant"}, nil
			},
		})
		_, err := svc.Update(context.Background(), "r1", testutil.TestTenantID, domain.CreateRouteRequest{Name: "Hack"})
		testutil.AssertError(t, err, "route not found")
	})
}

func TestRouteService_Update_NotFound(t *testing.T) {
	svc := NewRouteService(&mockRouteRepo{
		GetByIDFn: func(_ context.Context, _ string) (*domain.Route, error) {
			return nil, fmt.Errorf("not found")
		},
	})
	_, err := svc.Update(context.Background(), "nonexistent", testutil.TestTenantID, domain.CreateRouteRequest{})
	testutil.AssertError(t, err, "route not found")
}
