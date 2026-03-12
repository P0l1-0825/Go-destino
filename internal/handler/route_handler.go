package handler

import (
	"encoding/json"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type RouteHandler struct {
	routeSvc *service.RouteService
}

func NewRouteHandler(routeSvc *service.RouteService) *RouteHandler {
	return &RouteHandler{routeSvc: routeSvc}
}

func (h *RouteHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req domain.CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	route, err := h.routeSvc.Create(r.Context(), tenantID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, route)
}

func (h *RouteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	route, err := h.routeSvc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "route not found")
		return
	}
	response.JSON(w, http.StatusOK, route)
}

func (h *RouteHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	transportType := r.URL.Query().Get("transport_type")
	if transportType != "" {
		routes, err := h.routeSvc.ListByTransportType(r.Context(), tenantID, domain.TransportType(transportType))
		if err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		response.JSON(w, http.StatusOK, routes)
		return
	}

	routes, err := h.routeSvc.ListByTenant(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, routes)
}

func (h *RouteHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}

	var req domain.CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	route, err := h.routeSvc.Update(r.Context(), id, tenantID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, route)
}

func (h *RouteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.routeSvc.Deactivate(r.Context(), id, tenantID); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}
