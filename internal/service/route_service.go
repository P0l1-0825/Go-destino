package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

type RouteService struct {
	routeRepo *repository.RouteRepository
}

func NewRouteService(routeRepo *repository.RouteRepository) *RouteService {
	return &RouteService{routeRepo: routeRepo}
}

func (s *RouteService) Create(ctx context.Context, tenantID string, req domain.CreateRouteRequest) (*domain.Route, error) {
	route := &domain.Route{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Name:          req.Name,
		Code:          req.Code,
		TransportType: req.TransportType,
		Origin:        req.Origin,
		Destination:   req.Destination,
		PriceCents:    req.PriceCents,
		Currency:      req.Currency,
		Active:        true,
	}

	if err := s.routeRepo.Create(ctx, route); err != nil {
		return nil, fmt.Errorf("creating route: %w", err)
	}

	return route, nil
}

func (s *RouteService) GetByID(ctx context.Context, id string) (*domain.Route, error) {
	return s.routeRepo.GetByID(ctx, id)
}

func (s *RouteService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Route, error) {
	return s.routeRepo.ListByTenant(ctx, tenantID)
}

func (s *RouteService) ListByTransportType(ctx context.Context, tenantID string, transportType domain.TransportType) ([]domain.Route, error) {
	return s.routeRepo.ListByTransportType(ctx, tenantID, transportType)
}
