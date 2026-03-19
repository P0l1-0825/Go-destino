package service

import (
	"context"
	"fmt"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

type ConcesionService struct {
	concesionRepo *repository.ConcesionRepository
	auditSvc      *AuditService
}

func NewConcesionService(concesionRepo *repository.ConcesionRepository, auditSvc *AuditService) *ConcesionService {
	return &ConcesionService{
		concesionRepo: concesionRepo,
		auditSvc:      auditSvc,
	}
}

// Create registers a new concesion.
func (s *ConcesionService) Create(ctx context.Context, tenantID string, req domain.CreateConcesionRequest) (*domain.Concesion, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.MaxVehicles <= 0 {
		req.MaxVehicles = 50
	}
	if req.MaxDrivers <= 0 {
		req.MaxDrivers = 50
	}

	c := &domain.Concesion{
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        req.Code,
		RFC:         req.RFC,
		Type:        req.Type,
		Status:      domain.ConcesionPending,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		MaxVehicles: req.MaxVehicles,
		MaxDrivers:  req.MaxDrivers,
	}
	if c.Type == "" {
		c.Type = domain.ConcesionMixed
	}

	if err := s.concesionRepo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("creating concesion: %w", err)
	}

	if s.auditSvc != nil {
		go s.auditSvc.Log(context.Background(), tenantID, "", "concesion.created", "concesion", c.ID, c.Name, "", "")
	}

	return c, nil
}

// GetByID retrieves a concesion with aggregated stats.
func (s *ConcesionService) GetByID(ctx context.Context, id, tenantID string) (*domain.Concesion, error) {
	c, err := s.concesionRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	// Populate stats
	c.DriverCount, _ = s.concesionRepo.CountDrivers(ctx, id, tenantID)
	c.VehicleCount, _ = s.concesionRepo.CountVehicles(ctx, id, tenantID)
	c.StaffCount, _ = s.concesionRepo.CountStaff(ctx, id, tenantID)
	return c, nil
}

// List returns paginated concesiones with filters.
func (s *ConcesionService) List(ctx context.Context, f domain.ListConcesionesFilter) ([]domain.Concesion, int, error) {
	return s.concesionRepo.List(ctx, f)
}

// Update modifies a concesion.
func (s *ConcesionService) Update(ctx context.Context, id, tenantID string, req domain.UpdateConcesionRequest) (*domain.Concesion, error) {
	c, err := s.concesionRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("concesion not found: %w", err)
	}

	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Phone != nil {
		c.Phone = *req.Phone
	}
	if req.Email != nil {
		c.Email = *req.Email
	}
	if req.Address != nil {
		c.Address = *req.Address
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	if req.MaxVehicles != nil {
		c.MaxVehicles = *req.MaxVehicles
	}
	if req.MaxDrivers != nil {
		c.MaxDrivers = *req.MaxDrivers
	}
	if req.LogoURL != nil {
		c.LogoURL = *req.LogoURL
	}
	if req.Notes != nil {
		c.Notes = *req.Notes
	}

	if err := s.concesionRepo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("updating concesion: %w", err)
	}

	return c, nil
}

// Delete removes a concesion.
func (s *ConcesionService) Delete(ctx context.Context, id, tenantID string) error {
	return s.concesionRepo.Delete(ctx, id, tenantID)
}

// ListStaff returns all users assigned to a concesion.
func (s *ConcesionService) ListStaff(ctx context.Context, concesionID, tenantID string) ([]domain.User, error) {
	return s.concesionRepo.ListStaff(ctx, concesionID, tenantID)
}

// AssignStaff adds a user to a concesion with a specific staff role.
func (s *ConcesionService) AssignStaff(ctx context.Context, concesionID, tenantID string, req domain.AssignStaffRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	// Validate staff role
	valid := false
	for _, r := range domain.ValidStaffRoles() {
		if r == req.StaffRole {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid staff role: %s (valid: administrativo, operativo, taxista)", req.StaffRole)
	}

	if err := s.concesionRepo.AssignStaff(ctx, req.UserID, concesionID, tenantID); err != nil {
		return fmt.Errorf("assigning staff: %w", err)
	}

	// If assigning administrativo, also set as manager
	if req.StaffRole == domain.StaffAdministrativo {
		_ = s.concesionRepo.SetManager(ctx, concesionID, req.UserID, tenantID)
	}

	if s.auditSvc != nil {
		go s.auditSvc.Log(context.Background(), tenantID, "", "concesion.staff.assigned", "concesion", concesionID, req.UserID, string(req.StaffRole), "")
	}

	return nil
}

// RemoveStaff removes a user from a concesion.
func (s *ConcesionService) RemoveStaff(ctx context.Context, userID, tenantID string) error {
	return s.concesionRepo.RemoveStaff(ctx, userID, tenantID)
}
