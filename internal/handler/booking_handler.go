package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
	"github.com/P0l1-0825/Go-destino/pkg/validator"
)

type BookingHandler struct {
	bookingSvc *service.BookingService
}

func NewBookingHandler(bookingSvc *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc}
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")

	var req domain.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if err := validator.ValidateRequired(req.PickupAddress, "pickup_address"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validator.ValidateRequired(req.DropoffAddress, "dropoff_address"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if !domain.ValidServiceType(string(req.ServiceType)) {
		response.Error(w, http.StatusBadRequest, "invalid service_type: must be taxi, shuttle, van, or bus")
		return
	}
	if req.PassengerCount < 1 || req.PassengerCount > 50 {
		response.Error(w, http.StatusBadRequest, "passenger_count must be between 1 and 50")
		return
	}

	booking, err := h.bookingSvc.Create(r.Context(), tenantID, userID, kioskID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, booking)
}

func (h *BookingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}

	booking, err := h.bookingSvc.GetByIDTenant(r.Context(), id, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "booking not found")
		return
	}
	response.JSON(w, http.StatusOK, booking)
}

func (h *BookingHandler) GetByNumber(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	number := r.PathValue("number")
	if number == "" {
		response.Error(w, http.StatusBadRequest, "booking number is required")
		return
	}

	booking, err := h.bookingSvc.GetByNumberTenant(r.Context(), number, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "booking not found")
		return
	}
	response.JSON(w, http.StatusOK, booking)
}

func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req domain.CancelBookingRequest
	// Body is optional for cancel
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.bookingSvc.Cancel(r.Context(), id, req.Reason); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *BookingHandler) AssignDriver(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req domain.AssignDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DriverID == "" || req.VehicleID == "" {
		response.Error(w, http.StatusBadRequest, "driver_id and vehicle_id are required")
		return
	}

	if err := h.bookingSvc.AssignDriver(r.Context(), id, req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "assigned"})
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

	if req.Status == "" {
		response.Error(w, http.StatusBadRequest, "status is required")
		return
	}

	if err := h.bookingSvc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		if strings.Contains(err.Error(), "invalid transition") || strings.Contains(err.Error(), "cannot transition") {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusBadRequest, err.Error())
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

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Support filtering
	filter := domain.ListBookingsFilter{
		TenantID:    tenantID,
		Status:      domain.BookingStatus(r.URL.Query().Get("status")),
		ServiceType: domain.ServiceType(r.URL.Query().Get("service_type")),
		UserID:      r.URL.Query().Get("user_id"),
		Limit:       limit,
		Offset:      offset,
	}

	bookings, total, err := h.bookingSvc.ListFiltered(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"bookings": bookings,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *BookingHandler) StartTrip(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.bookingSvc.StartTrip(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "invalid transition") || strings.Contains(err.Error(), "cannot transition") {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (h *BookingHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req domain.EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !domain.ValidServiceType(string(req.ServiceType)) {
		response.Error(w, http.StatusBadRequest, "invalid service_type")
		return
	}

	estimate, err := h.bookingSvc.Estimate(req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, estimate)
}
