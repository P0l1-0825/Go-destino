package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
	"github.com/google/uuid"
)

func hashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(h), err
}

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

// Get single user
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

// Create user (admin-created accounts)
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "email, password and name are required")
		return
	}

	exists, _ := h.userRepo.ExistsByEmail(r.Context(), tenantID, req.Email)
	if exists {
		response.Error(w, http.StatusConflict, "email already registered")
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         req.Role,
		SubRole:      req.SubRole,
		CompanyID:    req.CompanyID,
		Lang:         req.Lang,
		Active:       true,
	}
	if user.Role == "" {
		user.Role = domain.RoleUsuario
	}
	if user.Lang == "" {
		user.Lang = "es"
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.auditSvc.Log(r.Context(), tenantID, middleware.GetUserID(r.Context()), "create", "user", user.ID, "Created user: "+req.Email, r.RemoteAddr, r.UserAgent())
	response.JSON(w, http.StatusCreated, user)
}

// Update user profile fields
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	var req struct {
		Name      *string `json:"name"`
		Phone     *string `json:"phone"`
		SubRole   *string `json:"sub_role"`
		CompanyID *string `json:"company_id"`
		Lang      *string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.SubRole != nil {
		user.SubRole = *req.SubRole
	}
	if req.CompanyID != nil {
		user.CompanyID = *req.CompanyID
	}
	if req.Lang != nil {
		user.Lang = *req.Lang
	}

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.auditSvc.Log(r.Context(), tenantID, middleware.GetUserID(r.Context()), "update", "user", id, "Updated user profile", r.RemoteAddr, r.UserAgent())
	response.JSON(w, http.StatusOK, user)
}

// Activate user
func (h *AdminHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	if err := h.userRepo.Activate(r.Context(), id); err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	h.auditSvc.Log(r.Context(), tenantID, middleware.GetUserID(r.Context()), "activate", "user", id, "Activated user", r.RemoteAddr, r.UserAgent())
	response.JSON(w, http.StatusOK, map[string]string{"status": "activated"})
}

// Deactivate user
func (h *AdminHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	if err := h.userRepo.Deactivate(r.Context(), id); err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	h.auditSvc.Log(r.Context(), tenantID, middleware.GetUserID(r.Context()), "deactivate", "user", id, "Deactivated user", r.RemoteAddr, r.UserAgent())
	response.JSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

// Update user role
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	var req struct {
		Role domain.UserRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role == "" {
		response.Error(w, http.StatusBadRequest, "role is required")
		return
	}

	if err := h.userRepo.UpdateRole(r.Context(), id, req.Role); err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	h.auditSvc.Log(r.Context(), tenantID, middleware.GetUserID(r.Context()), "update_role", "user", id, "Changed role to: "+string(req.Role), r.RemoteAddr, r.UserAgent())
	response.JSON(w, http.StatusOK, map[string]string{"status": "role_updated", "role": string(req.Role)})
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
