package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

const defaultCommissionRate = 0.05 // 5% commission

type ShiftService struct {
	shiftRepo *repository.ShiftRepository
}

func NewShiftService(shiftRepo *repository.ShiftRepository) *ShiftService {
	return &ShiftService{shiftRepo: shiftRepo}
}

func (s *ShiftService) OpenShift(ctx context.Context, tenantID, sellerID, airportID, terminalID, kioskID string) (*domain.ShiftRecord, error) {
	// Check if seller already has an open shift
	existing, err := s.shiftRepo.GetActive(ctx, sellerID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("seller already has an open shift: %s", existing.ID)
	}

	shift := &domain.ShiftRecord{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		SellerID:   sellerID,
		AirportID:  airportID,
		TerminalID: terminalID,
		KioskID:    kioskID,
	}

	if err := s.shiftRepo.Create(ctx, shift); err != nil {
		return nil, err
	}
	return shift, nil
}

func (s *ShiftService) CloseShift(ctx context.Context, shiftID string, totalSales, cashCollected, cardCollected int64, ticketsSold, bookingsCreated int) error {
	// Verify shift exists and is open
	shift, err := s.shiftRepo.GetByID(ctx, shiftID)
	if err != nil {
		return fmt.Errorf("shift not found: %w", err)
	}
	if shift.Status != "open" {
		return fmt.Errorf("shift is already closed")
	}

	// Validate totals
	if cashCollected+cardCollected != totalSales {
		return fmt.Errorf("cash (%d) + card (%d) must equal total sales (%d)", cashCollected, cardCollected, totalSales)
	}

	// Auto-calculate commission
	commissionCents := int64(float64(totalSales) * defaultCommissionRate)

	return s.shiftRepo.Close(ctx, shiftID, totalSales, cashCollected, cardCollected, commissionCents, ticketsSold, bookingsCreated)
}

func (s *ShiftService) GetActiveShift(ctx context.Context, sellerID string) (*domain.ShiftRecord, error) {
	return s.shiftRepo.GetActive(ctx, sellerID)
}

func (s *ShiftService) GetShiftByID(ctx context.Context, id string) (*domain.ShiftRecord, error) {
	return s.shiftRepo.GetByID(ctx, id)
}

func (s *ShiftService) ListShifts(ctx context.Context, sellerID string, limit int) ([]domain.ShiftRecord, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	return s.shiftRepo.ListBySeller(ctx, sellerID, limit)
}
