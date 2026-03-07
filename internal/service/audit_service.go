package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

type AuditService struct {
	auditRepo *repository.AuditRepository
}

func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

func (s *AuditService) Log(ctx context.Context, tenantID, userID, action, resource, resourceID, details, ip, ua string) {
	entry := &domain.AuditLogEntry{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  ip,
		UserAgent:  ua,
	}
	// Fire and forget — audit logging should not block
	go func() {
		_ = s.auditRepo.Create(context.Background(), entry)
	}()
}

func (s *AuditService) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.AuditLogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.auditRepo.ListByTenant(ctx, tenantID, limit)
}

func (s *AuditService) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.auditRepo.ListByUser(ctx, userID, limit)
}
