package service

import (
	"context"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/testutil"
)

// --- Fleet mocks ---

type mockDriverRepo struct {
	CreateFn             func(ctx context.Context, d *domain.Driver) error
	GetByIDTenantFn      func(ctx context.Context, id, tenantID string) (*domain.Driver, error)
	GetByUserIDFn        func(ctx context.Context, userID string) (*domain.Driver, error)
	UpdateLocationFn     func(ctx context.Context, driverID, tenantID string, lat, lng, heading, speed float64) error
	UpdateStatusFn       func(ctx context.Context, driverID, tenantID string, status domain.DriverStatus) error
	UpdateRatingFn       func(ctx context.Context, driverID, tenantID string, rating float64, totalTrips int) error
	SetDocsVerifiedFn    func(ctx context.Context, driverID, tenantID string, verified bool) error
	ListByTenantFn       func(ctx context.Context, tenantID string) ([]domain.Driver, error)
	ListByCompanyFn      func(ctx context.Context, companyID, tenantID string) ([]domain.Driver, error)
	GetActiveLocationsFn func(ctx context.Context, tenantID string) ([]domain.DriverLocation, error)
}

func (m *mockDriverRepo) Create(ctx context.Context, d *domain.Driver) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, d) }
	return nil
}
func (m *mockDriverRepo) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Driver, error) {
	if m.GetByIDTenantFn != nil { return m.GetByIDTenantFn(ctx, id, tenantID) }
	return nil, nil
}
func (m *mockDriverRepo) GetByUserID(ctx context.Context, userID string) (*domain.Driver, error) {
	if m.GetByUserIDFn != nil { return m.GetByUserIDFn(ctx, userID) }
	return nil, nil
}
func (m *mockDriverRepo) UpdateLocation(ctx context.Context, driverID, tenantID string, lat, lng, heading, speed float64) error {
	if m.UpdateLocationFn != nil { return m.UpdateLocationFn(ctx, driverID, tenantID, lat, lng, heading, speed) }
	return nil
}
func (m *mockDriverRepo) UpdateStatus(ctx context.Context, driverID, tenantID string, status domain.DriverStatus) error {
	if m.UpdateStatusFn != nil { return m.UpdateStatusFn(ctx, driverID, tenantID, status) }
	return nil
}
func (m *mockDriverRepo) UpdateRating(ctx context.Context, driverID, tenantID string, rating float64, totalTrips int) error {
	if m.UpdateRatingFn != nil { return m.UpdateRatingFn(ctx, driverID, tenantID, rating, totalTrips) }
	return nil
}
func (m *mockDriverRepo) SetDocsVerified(ctx context.Context, driverID, tenantID string, verified bool) error {
	if m.SetDocsVerifiedFn != nil { return m.SetDocsVerifiedFn(ctx, driverID, tenantID, verified) }
	return nil
}
func (m *mockDriverRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Driver, error) {
	if m.ListByTenantFn != nil { return m.ListByTenantFn(ctx, tenantID) }
	return nil, nil
}
func (m *mockDriverRepo) ListByCompany(ctx context.Context, companyID, tenantID string) ([]domain.Driver, error) {
	if m.ListByCompanyFn != nil { return m.ListByCompanyFn(ctx, companyID, tenantID) }
	return nil, nil
}
func (m *mockDriverRepo) GetActiveLocations(ctx context.Context, tenantID string) ([]domain.DriverLocation, error) {
	if m.GetActiveLocationsFn != nil { return m.GetActiveLocationsFn(ctx, tenantID) }
	return nil, nil
}

type mockVehicleRepo struct {
	CreateFn              func(ctx context.Context, v *domain.Vehicle) error
	GetByIDTenantFn       func(ctx context.Context, id, tenantID string) (*domain.Vehicle, error)
	GetByDriverIDTenantFn func(ctx context.Context, driverID, tenantID string) (*domain.Vehicle, error)
	ListByTenantFn        func(ctx context.Context, tenantID string) ([]domain.Vehicle, error)
}

func (m *mockVehicleRepo) Create(ctx context.Context, v *domain.Vehicle) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, v) }
	return nil
}
func (m *mockVehicleRepo) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Vehicle, error) {
	if m.GetByIDTenantFn != nil { return m.GetByIDTenantFn(ctx, id, tenantID) }
	return nil, nil
}
func (m *mockVehicleRepo) GetByDriverIDTenant(ctx context.Context, driverID, tenantID string) (*domain.Vehicle, error) {
	if m.GetByDriverIDTenantFn != nil { return m.GetByDriverIDTenantFn(ctx, driverID, tenantID) }
	return nil, nil
}
func (m *mockVehicleRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Vehicle, error) {
	if m.ListByTenantFn != nil { return m.ListByTenantFn(ctx, tenantID) }
	return nil, nil
}

// --- Tests ---

func TestFleetService_RegisterDriver(t *testing.T) {
	svc := NewFleetService(&mockDriverRepo{}, &mockVehicleRepo{})
	driver, err := svc.RegisterDriver(context.Background(), testutil.TestTenantID, domain.RegisterDriverRequest{
		UserID: testutil.TestUserID, LicenseNumber: "LIC-123",
	})
	testutil.AssertNoError(t, err)
	if driver.TenantID != testutil.TestTenantID { t.Errorf("tenant = %s", driver.TenantID) }
	if driver.Status != domain.DriverOffline { t.Errorf("status = %s, want offline", driver.Status) }
	if driver.Rating != 5.0 { t.Errorf("rating = %f, want 5.0", driver.Rating) }
}

func TestFleetService_RegisterVehicle(t *testing.T) {
	svc := NewFleetService(&mockDriverRepo{}, &mockVehicleRepo{})
	vehicle, err := svc.RegisterVehicle(context.Background(), testutil.TestTenantID, domain.RegisterVehicleRequest{
		DriverID: testutil.TestDriverID, Plate: "ABC-123", Brand: "Toyota", Model: "Camry",
		Year: 2024, Color: "White", Type: "sedan", Capacity: 4,
	})
	testutil.AssertNoError(t, err)
	if vehicle.TenantID != testutil.TestTenantID { t.Errorf("tenant = %s", vehicle.TenantID) }
	if vehicle.Status != "active" { t.Errorf("status = %s", vehicle.Status) }
}

func TestFleetService_RateDriver(t *testing.T) {
	svc := NewFleetService(&mockDriverRepo{
		GetByIDTenantFn: func(_ context.Context, _, _ string) (*domain.Driver, error) {
			return &domain.Driver{ID: "d1", Rating: 4.5, TotalTrips: 10}, nil
		},
		UpdateRatingFn: func(_ context.Context, _, _ string, rating float64, totalTrips int) error {
			if totalTrips != 11 { t.Errorf("totalTrips = %d, want 11", totalTrips) }
			// Expected: (4.5*10 + 5.0) / 11 ≈ 4.55
			if rating < 4.54 || rating > 4.56 { t.Errorf("rating = %f, want ~4.55", rating) }
			return nil
		},
	}, &mockVehicleRepo{})

	err := svc.RateDriver(context.Background(), "d1", testutil.TestTenantID, 5.0)
	testutil.AssertNoError(t, err)
}

func TestFleetService_FindNearbyDrivers(t *testing.T) {
	svc := NewFleetService(&mockDriverRepo{
		ListByTenantFn: func(_ context.Context, _ string) ([]domain.Driver, error) {
			return []domain.Driver{
				{ID: "d1", Status: domain.DriverAvailable, CurrentLat: 19.4363, CurrentLng: -99.0721, Rating: 4.8},
				{ID: "d2", Status: domain.DriverOffline, CurrentLat: 19.4363, CurrentLng: -99.0721, Rating: 4.5},   // offline
				{ID: "d3", Status: domain.DriverAvailable, CurrentLat: 20.0, CurrentLng: -100.0, Rating: 4.9},        // too far
				{ID: "d4", Status: domain.DriverAvailable, CurrentLat: 19.437, CurrentLng: -99.073, Rating: 3.0},      // low rating
			}, nil
		},
	}, &mockVehicleRepo{})

	drivers, err := svc.FindNearbyDrivers(context.Background(), testutil.TestTenantID, domain.NearbyDriversRequest{
		Lat: 19.4363, Lng: -99.0721, RadiusKM: 5.0, MinRating: 4.0,
	})
	testutil.AssertNoError(t, err)
	if len(drivers) != 1 { t.Fatalf("expected 1 nearby driver, got %d", len(drivers)) }
	if drivers[0].ID != "d1" { t.Errorf("expected d1, got %s", drivers[0].ID) }
}
