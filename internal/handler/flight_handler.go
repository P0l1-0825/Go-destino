package handler

import (
	"encoding/json"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type FlightHandler struct {
	flightSvc *service.FlightService
}

func NewFlightHandler(flightSvc *service.FlightService) *FlightHandler {
	return &FlightHandler{flightSvc: flightSvc}
}

func (h *FlightHandler) GetFlightStatus(w http.ResponseWriter, r *http.Request) {
	flightNumber := r.PathValue("number")
	info, err := h.flightSvc.GetFlightInfo(r.Context(), flightNumber)
	if err != nil {
		response.Error(w, http.StatusNotFound, "flight not found")
		return
	}
	response.JSON(w, http.StatusOK, info)
}

func (h *FlightHandler) ListArrivals(w http.ResponseWriter, r *http.Request) {
	airportCode := r.PathValue("code")
	flights, err := h.flightSvc.ListArrivals(r.Context(), airportCode)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, flights)
}

func (h *FlightHandler) ReportIROPS(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var event domain.IROPSEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.flightSvc.HandleIROPS(r.Context(), tenantID, event)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, result)
}
