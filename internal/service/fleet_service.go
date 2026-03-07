package service

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
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

func (s *FleetService) UpdateDriverLocation(ctx context.Context, loc domain.DriverLocation) error {
	return s.driverRepo.UpdateLocation(ctx, loc.DriverID, loc.Lat, loc.Lng, loc.Heading, loc.Speed)
}

func (s *FleetService) UpdateDriverStatus(ctx context.Context, driverID string, status domain.DriverStatus) error {
	return s.driverRepo.UpdateStatus(ctx, driverID, status)
}

func (s *FleetService) GetDriver(ctx context.Context, id string) (*domain.Driver, error) {
	return s.driverRepo.GetByID(ctx, id)
}

func (s *FleetService) GetDriverByUserID(ctx context.Context, userID string) (*domain.Driver, error) {
	return s.driverRepo.GetByUserID(ctx, userID)
}

func (s *FleetService) ListDrivers(ctx context.Context, tenantID string) ([]domain.Driver, error) {
	return s.driverRepo.ListByTenant(ctx, tenantID)
}

func (s *FleetService) ListDriversByCompany(ctx context.Context, companyID string) ([]domain.Driver, error) {
	return s.driverRepo.ListByCompany(ctx, companyID)
}

func (s *FleetService) VerifyDriverDocs(ctx context.Context, driverID string, verified bool) error {
	return s.driverRepo.SetDocsVerified(ctx, driverID, verified)
}

func (s *FleetService) RateDriver(ctx context.Context, driverID string, newRating float64) error {
	driver, err := s.driverRepo.GetByID(ctx, driverID)
	if err != nil {
		return err
	}
	// Weighted average: (old * count + new) / (count + 1)
	totalTrips := driver.TotalTrips + 1
	avgRating := (driver.Rating*float64(driver.TotalTrips) + newRating) / float64(totalTrips)
	avgRating = math.Round(avgRating*100) / 100
	return s.driverRepo.UpdateRating(ctx, driverID, avgRating, totalTrips)
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
		dist := haversine(req.Lat, req.Lng, d.CurrentLat, d.CurrentLng)
		if dist <= req.RadiusKM {
			nearby = append(nearby, d)
		}
	}
	return nearby, nil
}

func (s *FleetService) GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error) {
	return s.vehicleRepo.GetByID(ctx, id)
}

func (s *FleetService) GetVehicleByDriver(ctx context.Context, driverID string) (*domain.Vehicle, error) {
	return s.vehicleRepo.GetByDriverID(ctx, driverID)
}

func (s *FleetService) ListVehicles(ctx context.Context, tenantID string) ([]domain.Vehicle, error) {
	return s.vehicleRepo.ListByTenant(ctx, tenantID)
}

// haversine computes distance in km between two lat/lng points.
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
