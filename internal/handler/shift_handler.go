package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
	"github.com/P0l1-0825/Go-destino/pkg/validator"
)

type ShiftHandler struct {
	shiftSvc *service.ShiftService
}

func NewShiftHandler(shiftSvc *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{shiftSvc: shiftSvc}
}

func (h *ShiftHandler) Open(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req struct {
		AirportID  string `json:"airport_id"`
		TerminalID string `json:"terminal_id"`
		KioskID    string `json:"kiosk_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateRequired(req.AirportID, "airport_id"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validator.ValidateRequired(req.KioskID, "kiosk_id"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	shift, err := h.shiftSvc.OpenShift(r.Context(), tenantID, userID, req.AirportID, req.TerminalID, req.KioskID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, shift)
}

func (h *ShiftHandler) Close(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "shift id is required")
		return
	}

	var req struct {
		TotalSales      int64 `json:"total_sales_cents"`
		CashCollected   int64 `json:"cash_collected_cents"`
		CardCollected   int64 `json:"card_collected_cents"`
		TicketsSold     int   `json:"tickets_sold"`
		BookingsCreated int   `json:"bookings_created"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TotalSales < 0 {
		response.Error(w, http.StatusBadRequest, "total_sales_cents cannot be negative")
		return
	}

	if err := h.shiftSvc.CloseShift(r.Context(), id, req.TotalSales, req.CashCollected, req.CardCollected, req.TicketsSold, req.BookingsCreated); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func (h *ShiftHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	shift, err := h.shiftSvc.GetActiveShift(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "no active shift")
		return
	}
	response.JSON(w, http.StatusOK, shift)
}

func (h *ShiftHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	shifts, err := h.shiftSvc.ListShifts(r.Context(), userID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, shifts)
}
