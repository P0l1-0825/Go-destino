package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

type KioskService struct {
	kioskRepo *repository.KioskRepository
}

func NewKioskService(kioskRepo *repository.KioskRepository) *KioskService {
	return &KioskService{kioskRepo: kioskRepo}
}

func (s *KioskService) Register(ctx context.Context, tenantID string, req domain.RegisterKioskRequest) (*domain.Kiosk, error) {
	kiosk := &domain.Kiosk{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Name:       req.Name,
		Location:   req.Location,
		AirportID:  req.AirportID,
		TerminalID: req.TerminalID,
		Status:     domain.KioskOnline,
	}

	if err := s.kioskRepo.Create(ctx, kiosk); err != nil {
		return nil, fmt.Errorf("registering kiosk: %w", err)
	}

	return kiosk, nil
}

func (s *KioskService) GetByID(ctx context.Context, id string) (*domain.Kiosk, error) {
	return s.kioskRepo.GetByID(ctx, id)
}

func (s *KioskService) Heartbeat(ctx context.Context, id string) error {
	return s.kioskRepo.UpdateHeartbeat(ctx, id)
}

func (s *KioskService) UpdateStatus(ctx context.Context, id string, status domain.KioskStatus) error {
	return s.kioskRepo.UpdateStatus(ctx, id, status)
}

func (s *KioskService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Kiosk, error) {
	return s.kioskRepo.ListByTenant(ctx, tenantID)
}
