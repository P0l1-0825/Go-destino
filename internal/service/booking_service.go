package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

type BookingService struct {
	bookingRepo *repository.BookingRepository
	paymentRepo *repository.PaymentRepository
}

func NewBookingService(bookingRepo *repository.BookingRepository, paymentRepo *repository.PaymentRepository) *BookingService {
	return &BookingService{bookingRepo: bookingRepo, paymentRepo: paymentRepo}
}

func (s *BookingService) Create(ctx context.Context, tenantID, kioskID string, req domain.CreateBookingRequest) (*domain.Booking, error) {
	bookingNumber, err := generateBookingNumber()
	if err != nil {
		return nil, fmt.Errorf("generating booking number: %w", err)
	}

	booking := &domain.Booking{
		ID:             uuid.New().String(),
		BookingNumber:  bookingNumber,
		TenantID:       tenantID,
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
		FlightNumber:   req.FlightNumber,
		ScheduledAt:    req.ScheduledAt,
		Currency:       "MXN",
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

func (s *BookingService) Cancel(ctx context.Context, id string) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	if booking.Status == domain.BookingCompleted || booking.Status == domain.BookingCancelled {
		return fmt.Errorf("booking cannot be cancelled in status %s", booking.Status)
	}

	return s.bookingRepo.UpdateStatus(ctx, id, domain.BookingCancelled)
}

func (s *BookingService) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	return s.bookingRepo.UpdateStatus(ctx, id, status)
}

func (s *BookingService) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.Booking, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.bookingRepo.ListByTenant(ctx, tenantID, limit)
}

func (s *BookingService) Estimate(req domain.EstimateRequest) (*domain.EstimateResponse, error) {
	// Simplified pricing: base price + distance factor
	baseCents := int64(5000) // $50.00 MXN base
	if req.ServiceType == domain.ServiceVan {
		baseCents = 8000
	} else if req.ServiceType == domain.ServiceShuttle {
		baseCents = 3500
	} else if req.ServiceType == domain.ServiceBus {
		baseCents = 2500
	}

	return &domain.EstimateResponse{
		PriceCents: baseCents * int64(max(req.PassengerCount, 1)),
		Currency:   "MXN",
		ETAMinutes: 15,
		Distance:   "~20 km",
	}, nil
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *BookingService) CompleteBooking(ctx context.Context, id string) error {
	now := time.Now()
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if booking.Status != domain.BookingStarted {
		return fmt.Errorf("booking must be in started status to complete")
	}
	_ = now
	return s.bookingRepo.UpdateStatus(ctx, id, domain.BookingCompleted)
}
