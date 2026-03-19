package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type ConcesionHandler struct {
	concesionSvc *service.ConcesionService
}

func NewConcesionHandler(concesionSvc *service.ConcesionService) *ConcesionHandler {
	return &ConcesionHandler{concesionSvc: concesionSvc}
}

// Create registers a new concesion.
func (h *ConcesionHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req domain.CreateConcesionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	c, err := h.concesionSvc.Create(r.Context(), tenantID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, c)
}

// GetByID retrieves a concesion by ID.
func (h *ConcesionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	c, err := h.concesionSvc.GetByID(r.Context(), id, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "concesion not found")
		return
	}
	response.JSON(w, http.StatusOK, c)
}

// List returns paginated concesiones.
func (h *ConcesionHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}

	f := domain.ListConcesionesFilter{
		TenantID: tenantID,
		Status:   domain.ConcesionStatus(r.URL.Query().Get("status")),
		Type:     domain.ConcesionType(r.URL.Query().Get("type")),
		Search:   r.URL.Query().Get("search"),
		Limit:    limit,
		Offset:   offset,
	}

	concesiones, total, err := h.concesionSvc.List(r.Context(), f)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSONWithMeta(w, http.StatusOK, concesiones, total, limit, offset)
}

// Update modifies a concesion.
func (h *ConcesionHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	var req domain.UpdateConcesionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.concesionSvc.Update(r.Context(), id, tenantID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, c)
}

// Delete removes a concesion.
func (h *ConcesionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	if err := h.concesionSvc.Delete(r.Context(), id, tenantID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListStaff returns all users assigned to a concesion.
func (h *ConcesionHandler) ListStaff(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	staff, err := h.concesionSvc.ListStaff(r.Context(), id, tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, staff)
}

// AssignStaff adds a user to a concesion.
func (h *ConcesionHandler) AssignStaff(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	var req domain.AssignStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		response.Error(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.concesionSvc.AssignStaff(r.Context(), id, tenantID, req); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

// RemoveStaff removes a user from a concesion.
func (h *ConcesionHandler) RemoveStaff(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := r.PathValue("userId")

	if err := h.concesionSvc.RemoveStaff(r.Context(), userID, tenantID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
