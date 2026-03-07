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

type BookingHandler struct {
	bookingSvc *service.BookingService
}

func NewBookingHandler(bookingSvc *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc}
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")

	var req domain.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	booking, err := h.bookingSvc.Create(r.Context(), tenantID, kioskID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, booking)
}

func (h *BookingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	booking, err := h.bookingSvc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "booking not found")
		return
	}
	response.JSON(w, http.StatusOK, booking)
}

func (h *BookingHandler) GetByNumber(w http.ResponseWriter, r *http.Request) {
	number := r.PathValue("number")
	booking, err := h.bookingSvc.GetByNumber(r.Context(), number)
	if err != nil {
		response.Error(w, http.StatusNotFound, "booking not found")
		return
	}
	response.JSON(w, http.StatusOK, booking)
}

func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.bookingSvc.Cancel(r.Context(), id); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *BookingHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Status domain.BookingStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.bookingSvc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	bookings, err := h.bookingSvc.ListByTenant(r.Context(), tenantID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, bookings)
}

func (h *BookingHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req domain.EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	estimate, err := h.bookingSvc.Estimate(req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, estimate)
}
