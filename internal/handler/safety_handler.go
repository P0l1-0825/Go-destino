package handler

import (
	"encoding/json"
	"net/http"

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

func (h *SafetyHandler) GetEmergencyNumbers(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, domain.EmergencyNumbers)
}
