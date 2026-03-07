package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/P0l1-0825/Go-destino/pkg/geo"
)

type BookingService struct {
	bookingRepo *repository.BookingRepository
	paymentRepo *repository.PaymentRepository
}

func NewBookingService(bookingRepo *repository.BookingRepository, paymentRepo *repository.PaymentRepository) *BookingService {
	return &BookingService{bookingRepo: bookingRepo, paymentRepo: paymentRepo}
}

func (s *BookingService) Create(ctx context.Context, tenantID, userID, kioskID string, req domain.CreateBookingRequest) (*domain.Booking, error) {
	if !domain.ValidServiceType(string(req.ServiceType)) {
		return nil, fmt.Errorf("invalid service type: %s", req.ServiceType)
	}
	if req.PassengerCount < 1 || req.PassengerCount > 50 {
		return nil, fmt.Errorf("passenger count must be between 1 and 50")
	}
	if req.PickupLat < -90 || req.PickupLat > 90 || req.DropoffLat < -90 || req.DropoffLat > 90 {
		return nil, fmt.Errorf("latitude must be between -90 and 90")
	}
	if req.PickupLng < -180 || req.PickupLng > 180 || req.DropoffLng < -180 || req.DropoffLng > 180 {
		return nil, fmt.Errorf("longitude must be between -180 and 180")
	}

	bookingNumber, err := generateBookingNumber()
	if err != nil {
		return nil, fmt.Errorf("generating booking number: %w", err)
	}

	// Calculate price from estimate
	estimate, err := s.Estimate(domain.EstimateRequest{
		ServiceType:    req.ServiceType,
		PickupLat:      req.PickupLat,
		PickupLng:      req.PickupLng,
		DropoffLat:     req.DropoffLat,
		DropoffLng:     req.DropoffLng,
		PassengerCount: req.PassengerCount,
	})
	if err != nil {
		return nil, fmt.Errorf("estimating price: %w", err)
	}

	booking := &domain.Booking{
		ID:             uuid.New().String(),
		BookingNumber:  bookingNumber,
		TenantID:       tenantID,
		UserID:         userID,
		KioskID:        kioskID,
		RouteID:        req.RouteID,
		Status:         domain.BookingPending,
		ServiceType:    req.ServiceType,
		PickupAddress:  req.PickupAddress,
		DropoffAddress: req.DropoffAddress,
		PickupLat:      req.PickupLat,
		PickupLng:      req.PickupLng,
		DropoffLat:     req.DropoffLat,
		DropoffLng:     req.DropoffLng,
		PassengerCount: req.PassengerCount,
		PriceCents:     estimate.PriceCents,
		Currency:       estimate.Currency,
		FlightNumber:   req.FlightNumber,
		ScheduledAt:    req.ScheduledAt,
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("creating booking: %w", err)
	}

	return booking, nil
}

func (s *BookingService) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	return s.bookingRepo.GetByID(ctx, id)
}

func (s *BookingService) GetByNumber(ctx context.Context, number string) (*domain.Booking, error) {
	return s.bookingRepo.GetByNumber(ctx, number)
}

func (s *BookingService) Confirm(ctx context.Context, id string) error {
	return s.transitionStatus(ctx, id, domain.BookingConfirmed)
}

func (s *BookingService) AssignDriver(ctx context.Context, id string, req domain.AssignDriverRequest) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if err := domain.ValidBookingTransition(booking.Status, domain.BookingAssigned); err != nil {
		return err
	}
	return s.bookingRepo.AssignDriver(ctx, id, req.DriverID, req.VehicleID)
}

func (s *BookingService) StartTrip(ctx context.Context, id string) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if err := domain.ValidBookingTransition(booking.Status, domain.BookingStarted); err != nil {
		return err
	}
	return s.bookingRepo.SetStarted(ctx, id)
}

func (s *BookingService) CompleteBooking(ctx context.Context, id string) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if err := domain.ValidBookingTransition(booking.Status, domain.BookingCompleted); err != nil {
		return err
	}
	return s.bookingRepo.SetCompleted(ctx, id)
}

func (s *BookingService) Cancel(ctx context.Context, id, reason string) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if err := domain.ValidBookingTransition(booking.Status, domain.BookingCancelled); err != nil {
		return err
	}
	return s.bookingRepo.SetCancelled(ctx, id, reason)
}

func (s *BookingService) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if err := domain.ValidBookingTransition(booking.Status, status); err != nil {
		return err
	}

	switch status {
	case domain.BookingStarted:
		return s.bookingRepo.SetStarted(ctx, id)
	case domain.BookingCompleted:
		return s.bookingRepo.SetCompleted(ctx, id)
	case domain.BookingCancelled:
		return s.bookingRepo.SetCancelled(ctx, id, "")
	default:
		return s.bookingRepo.UpdateStatus(ctx, id, status)
	}
}

func (s *BookingService) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.Booking, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.bookingRepo.ListByTenant(ctx, tenantID, limit)
}

func (s *BookingService) ListFiltered(ctx context.Context, f domain.ListBookingsFilter) ([]domain.Booking, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 200 {
		f.Limit = 200
	}
	return s.bookingRepo.ListFiltered(ctx, f)
}

func (s *BookingService) Estimate(req domain.EstimateRequest) (*domain.EstimateResponse, error) {
	if req.PassengerCount < 1 {
		req.PassengerCount = 1
	}

	baseCents := int64(5000) // $50.00 MXN base
	switch req.ServiceType {
	case domain.ServiceVan:
		baseCents = 8000
	case domain.ServiceShuttle:
		baseCents = 3500
	case domain.ServiceBus:
		baseCents = 2500
	}

	// Haversine distance factor
	distKm := geo.Haversine(req.PickupLat, req.PickupLng, req.DropoffLat, req.DropoffLng)
	distFactor := int64(distKm * 300) // ~$3.00 MXN per km

	price := (baseCents + distFactor) * int64(req.PassengerCount)

	etaMinutes := int(distKm/0.8) + 5 // rough estimate: ~48 km/h avg + 5 min pickup

	return &domain.EstimateResponse{
		PriceCents: price,
		Currency:   "MXN",
		ETAMinutes: etaMinutes,
		Distance:   fmt.Sprintf("%.1f km", distKm),
	}, nil
}

func (s *BookingService) transitionStatus(ctx context.Context, id string, target domain.BookingStatus) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if err := domain.ValidBookingTransition(booking.Status, target); err != nil {
		return err
	}
	return s.bookingRepo.UpdateStatus(ctx, id, target)
}

func generateBookingNumber() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 8)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return "GD-" + string(result), nil
}

