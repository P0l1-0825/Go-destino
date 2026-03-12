package service

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/P0l1-0825/Go-destino/pkg/geo"
)

type FleetService struct {
	driverRepo  *repository.DriverRepository
	vehicleRepo *repository.VehicleRepository
}

func NewFleetService(driverRepo *repository.DriverRepository, vehicleRepo *repository.VehicleRepository) *FleetService {
	return &FleetService{driverRepo: driverRepo, vehicleRepo: vehicleRepo}
}

func (s *FleetService) RegisterDriver(ctx context.Context, tenantID string, req domain.RegisterDriverRequest) (*domain.Driver, error) {
	driver := &domain.Driver{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		UserID:        req.UserID,
		CompanyID:     req.CompanyID,
		LicenseNumber: req.LicenseNumber,
		SubRole:       req.SubRole,
		Status:        domain.DriverOffline,
		Rating:        5.0,
		TotalTrips:    0,
	}

	if err := s.driverRepo.Create(ctx, driver); err != nil {
		return nil, fmt.Errorf("registering driver: %w", err)
	}
	return driver, nil
}

func (s *FleetService) RegisterVehicle(ctx context.Context, tenantID string, req domain.RegisterVehicleRequest) (*domain.Vehicle, error) {
	vehicle := &domain.Vehicle{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		DriverID: req.DriverID,
		Plate:    req.Plate,
		Brand:    req.Brand,
		Model:    req.Model,
		Year:     req.Year,
		Color:    req.Color,
		Type:     req.Type,
		Capacity: req.Capacity,
		Status:   "active",
	}

	if err := s.vehicleRepo.Create(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("registering vehicle: %w", err)
	}
	return vehicle, nil
}

func (s *FleetService) UpdateDriverLocation(ctx context.Context, tenantID string, loc domain.DriverLocation) error {
	return s.driverRepo.UpdateLocation(ctx, loc.DriverID, tenantID, loc.Lat, loc.Lng, loc.Heading, loc.Speed)
}

func (s *FleetService) UpdateDriverStatus(ctx context.Context, driverID, tenantID string, status domain.DriverStatus) error {
	return s.driverRepo.UpdateStatus(ctx, driverID, tenantID, status)
}

func (s *FleetService) GetDriver(ctx context.Context, id, tenantID string) (*domain.Driver, error) {
	return s.driverRepo.GetByIDTenant(ctx, id, tenantID)
}

func (s *FleetService) GetDriverByUserID(ctx context.Context, userID string) (*domain.Driver, error) {
	return s.driverRepo.GetByUserID(ctx, userID)
}

func (s *FleetService) ListDrivers(ctx context.Context, tenantID string) ([]domain.Driver, error) {
	return s.driverRepo.ListByTenant(ctx, tenantID)
}

func (s *FleetService) ListDriversByCompany(ctx context.Context, companyID, tenantID string) ([]domain.Driver, error) {
	return s.driverRepo.ListByCompany(ctx, companyID, tenantID)
}

func (s *FleetService) VerifyDriverDocs(ctx context.Context, driverID, tenantID string, verified bool) error {
	return s.driverRepo.SetDocsVerified(ctx, driverID, tenantID, verified)
}

func (s *FleetService) RateDriver(ctx context.Context, driverID, tenantID string, newRating float64) error {
	driver, err := s.driverRepo.GetByIDTenant(ctx, driverID, tenantID)
	if err != nil {
		return err
	}
	// Weighted average: (old * count + new) / (count + 1)
	totalTrips := driver.TotalTrips + 1
	avgRating := (driver.Rating*float64(driver.TotalTrips) + newRating) / float64(totalTrips)
	avgRating = math.Round(avgRating*100) / 100
	return s.driverRepo.UpdateRating(ctx, driverID, tenantID, avgRating, totalTrips)
}

// FindNearbyDrivers returns available drivers within a radius (Haversine).
func (s *FleetService) FindNearbyDrivers(ctx context.Context, tenantID string, req domain.NearbyDriversRequest) ([]domain.Driver, error) {
	all, err := s.driverRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var nearby []domain.Driver
	for _, d := range all {
		if d.Status != domain.DriverAvailable {
			continue
		}
		if req.MinRating > 0 && d.Rating < req.MinRating {
			continue
		}
		dist := geo.Haversine(req.Lat, req.Lng, d.CurrentLat, d.CurrentLng)
		if dist <= req.RadiusKM {
			nearby = append(nearby, d)
		}
	}
	return nearby, nil
}

func (s *FleetService) GetActiveLocations(ctx context.Context, tenantID string) ([]domain.DriverLocation, error) {
	return s.driverRepo.GetActiveLocations(ctx, tenantID)
}

func (s *FleetService) GetVehicle(ctx context.Context, id, tenantID string) (*domain.Vehicle, error) {
	return s.vehicleRepo.GetByIDTenant(ctx, id, tenantID)
}

func (s *FleetService) GetVehicleByDriver(ctx context.Context, driverID, tenantID string) (*domain.Vehicle, error) {
	return s.vehicleRepo.GetByDriverIDTenant(ctx, driverID, tenantID)
}

func (s *FleetService) ListVehicles(ctx context.Context, tenantID string) ([]domain.Vehicle, error) {
	return s.vehicleRepo.ListByTenant(ctx, tenantID)
}

