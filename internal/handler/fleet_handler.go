package handler

import (
	"encoding/json"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type FleetHandler struct {
	fleetSvc *service.FleetService
}

func NewFleetHandler(fleetSvc *service.FleetService) *FleetHandler {
	return &FleetHandler{fleetSvc: fleetSvc}
}

func (h *FleetHandler) RegisterDriver(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req domain.RegisterDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	driver, err := h.fleetSvc.RegisterDriver(r.Context(), tenantID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, driver)
}

func (h *FleetHandler) RegisterVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req domain.RegisterVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	vehicle, err := h.fleetSvc.RegisterVehicle(r.Context(), tenantID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, vehicle)
}

func (h *FleetHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var loc domain.DriverLocation
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.fleetSvc.UpdateDriverLocation(r.Context(), tenantID, loc); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *FleetHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	var req struct {
		Status domain.DriverStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.fleetSvc.UpdateDriverStatus(r.Context(), id, tenantID, req.Status); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (h *FleetHandler) GetDriver(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	driver, err := h.fleetSvc.GetDriver(r.Context(), id, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "driver not found")
		return
	}
	response.JSON(w, http.StatusOK, driver)
}

func (h *FleetHandler) ListDrivers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	drivers, err := h.fleetSvc.ListDrivers(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, drivers)
}

func (h *FleetHandler) NearbyDrivers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req domain.NearbyDriversRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RadiusKM <= 0 {
		req.RadiusKM = 10
	}
	drivers, err := h.fleetSvc.FindNearbyDrivers(r.Context(), tenantID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, drivers)
}

func (h *FleetHandler) RateDriver(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Rating float64 `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		response.Error(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}
	tenantID := middleware.GetTenantID(r.Context())
	if err := h.fleetSvc.RateDriver(r.Context(), id, tenantID, req.Rating); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "rated"})
}

func (h *FleetHandler) VerifyDocs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	var req struct {
		Verified bool `json:"verified"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.fleetSvc.VerifyDriverDocs(r.Context(), id, tenantID, req.Verified); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]bool{"verified": req.Verified})
}

func (h *FleetHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	vehicles, err := h.fleetSvc.ListVehicles(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, vehicles)
}

func (h *FleetHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	vehicle, err := h.fleetSvc.GetVehicle(r.Context(), id, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "vehicle not found")
		return
	}
	response.JSON(w, http.StatusOK, vehicle)
}
