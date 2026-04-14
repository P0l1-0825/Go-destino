package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type SafetyHandler struct {
	safetySvc *service.SafetyService
}

func NewSafetyHandler(safetySvc *service.SafetyService) *SafetyHandler {
	return &SafetyHandler{safetySvc: safetySvc}
}

func (h *SafetyHandler) ReportIncident(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())
	var req domain.SafetyIncident
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.TenantID = tenantID
	req.ReportedBy = userID
	incident, err := h.safetySvc.ReportIncident(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, incident)
}

func (h *SafetyHandler) TriggerSOS(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())
	var req struct {
		BookingID string  `json:"booking_id"`
		Lat       float64 `json:"lat"`
		Lng       float64 `json:"lng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	alert, err := h.safetySvc.TriggerSOS(r.Context(), tenantID, userID, req.BookingID, req.Lat, req.Lng)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, alert)
}

func (h *SafetyHandler) ResolveSOS(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.safetySvc.ResolveSOS(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (h *SafetyHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	incidents, err := h.safetySvc.ListIncidents(r.Context(), tenantID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, incidents)
}

func (h *SafetyHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}
	incident, err := h.safetySvc.GetIncident(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "incident not found")
		return
	}
	response.JSON(w, http.StatusOK, incident)
}

func (h *SafetyHandler) GetEmergencyNumbers(w http.ResponseWriter, r *http.Request) {
	// Emergency numbers are compiled into the binary at build time and never
	// change at runtime; cache aggressively (1 hour) with ETag support.
	response.CachedJSON(w, r, http.StatusOK, domain.EmergencyNumbers, time.Hour)
}
