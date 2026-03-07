package handler

import (
	"encoding/json"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type KioskHandler struct {
	kioskSvc *service.KioskService
}

func NewKioskHandler(kioskSvc *service.KioskService) *KioskHandler {
	return &KioskHandler{kioskSvc: kioskSvc}
}

func (h *KioskHandler) Register(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req domain.RegisterKioskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	kiosk, err := h.kioskSvc.Register(r.Context(), tenantID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, kiosk)
}

func (h *KioskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	kiosk, err := h.kioskSvc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "kiosk not found")
		return
	}
	response.JSON(w, http.StatusOK, kiosk)
}

func (h *KioskHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.kioskSvc.Heartbeat(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *KioskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Status domain.KioskStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.kioskSvc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (h *KioskHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	kiosks, err := h.kioskSvc.ListByTenant(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, kiosks)
}
