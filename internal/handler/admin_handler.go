package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
	"github.com/google/uuid"
)

type AdminHandler struct {
	tenantRepo  *repository.TenantRepository
	userRepo    *repository.UserRepository
	airportRepo *repository.AirportRepository
	auditSvc    *service.AuditService
}

func NewAdminHandler(
	tenantRepo *repository.TenantRepository,
	userRepo *repository.UserRepository,
	airportRepo *repository.AirportRepository,
	auditSvc *service.AuditService,
) *AdminHandler {
	return &AdminHandler{
		tenantRepo:  tenantRepo,
		userRepo:    userRepo,
		airportRepo: airportRepo,
		auditSvc:    auditSvc,
	}
}

// Tenant management
func (h *AdminHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant := &domain.Tenant{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Slug:   req.Slug,
		Active: true,
		Plan:   req.Plan,
	}

	if err := h.tenantRepo.Create(r.Context(), tenant); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.auditSvc.Log(r.Context(), tenant.ID, middleware.GetUserID(r.Context()), "create", "tenant", tenant.ID, "Created tenant: "+req.Name, r.RemoteAddr, r.UserAgent())
	response.JSON(w, http.StatusCreated, tenant)
}

func (h *AdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.tenantRepo.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tenants)
}

func (h *AdminHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenant, err := h.tenantRepo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "tenant not found")
		return
	}
	response.JSON(w, http.StatusOK, tenant)
}

// User management
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	users, err := h.userRepo.ListByTenant(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, users)
}

// Airport management
func (h *AdminHandler) CreateAirport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req domain.Airport
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.ID = uuid.New().String()
	req.TenantID = tenantID
	req.Active = true

	if err := h.airportRepo.Create(r.Context(), &req); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.auditSvc.Log(r.Context(), tenantID, middleware.GetUserID(r.Context()), "create", "airport", req.ID, "Created airport: "+req.Code, r.RemoteAddr, r.UserAgent())
	response.JSON(w, http.StatusCreated, req)
}

func (h *AdminHandler) ListAirports(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	airports, err := h.airportRepo.ListByTenant(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, airports)
}

func (h *AdminHandler) GetAirport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	airport, err := h.airportRepo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "airport not found")
		return
	}
	response.JSON(w, http.StatusOK, airport)
}

// Audit log
func (h *AdminHandler) AuditLog(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	entries, err := h.auditSvc.ListByTenant(r.Context(), tenantID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, entries)
}

// Permissions introspection
func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles := []map[string]interface{}{}
	for role, perms := range domain.RolePermissions {
		roles = append(roles, map[string]interface{}{
			"role":             role,
			"permission_count": len(perms),
			"permissions":      perms,
		})
	}
	response.JSON(w, http.StatusOK, roles)
}

func (h *AdminHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, domain.AllPermissions())
}
