package handler

import (
	"encoding/json"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
	"github.com/P0l1-0825/Go-destino/pkg/validator"
)

// KioskUXHandler serves the kiosk-optimized endpoints for fast UX.
type KioskUXHandler struct {
	kioskUXSvc *service.KioskUXService
}

func NewKioskUXHandler(kioskUXSvc *service.KioskUXService) *KioskUXHandler {
	return &KioskUXHandler{kioskUXSvc: kioskUXSvc}
}

// Suggestions returns AI-powered smart suggestions for the kiosk home screen.
// GET /api/v1/kiosk/suggestions?lang=es
func (h *KioskUXHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	lang := r.URL.Query().Get("lang")

	suggestions, err := h.kioskUXSvc.GetSmartSuggestions(r.Context(), kioskID, tenantID, lang)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, suggestions)
}

// FlightLookup returns flight info + transport options.
// GET /api/v1/kiosk/flights/{number}?lang=es
func (h *KioskUXHandler) FlightLookup(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	flightNumber := r.PathValue("number")
	if flightNumber == "" {
		response.Error(w, http.StatusBadRequest, "flight number is required")
		return
	}

	lang := r.URL.Query().Get("lang")

	result, err := h.kioskUXSvc.LookupFlight(r.Context(), flightNumber, kioskID, tenantID, lang)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// RecommendService returns AI service recommendations based on context.
// POST /api/v1/kiosk/recommend
func (h *KioskUXHandler) RecommendService(w http.ResponseWriter, r *http.Request) {
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	var req struct {
		Passengers int     `json:"passengers"`
		DropoffLat float64 `json:"dropoff_lat"`
		DropoffLng float64 `json:"dropoff_lng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Passengers < 1 {
		req.Passengers = 1
	}

	recs, err := h.kioskUXSvc.RecommendService(r.Context(), req.Passengers, req.DropoffLat, req.DropoffLng, kioskID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"recommendations": recs,
		"passengers":      req.Passengers,
	})
}

// QuickBook performs a streamlined one-tap booking.
// POST /api/v1/kiosk/quick-book
func (h *KioskUXHandler) QuickBook(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	sellerID := middleware.GetUserID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	var req domain.QuickBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DropoffLat == 0 && req.DropoffLng == 0 {
		response.Error(w, http.StatusBadRequest, "dropoff coordinates required")
		return
	}
	if req.ServiceType == "" {
		response.Error(w, http.StatusBadRequest, "service_type required")
		return
	}
	if req.PaymentMethod == "" {
		response.Error(w, http.StatusBadRequest, "payment_method required")
		return
	}

	result, err := h.kioskUXSvc.QuickBook(r.Context(), kioskID, tenantID, sellerID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

// GetReceipt generates/retrieves a receipt for a booking.
// GET /api/v1/kiosk/receipts/{bookingId}?lang=es
func (h *KioskUXHandler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	bookingID := r.PathValue("bookingId")
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "booking ID required")
		return
	}

	lang := r.URL.Query().Get("lang")

	receipt, err := h.kioskUXSvc.GetReceipt(r.Context(), bookingID, kioskID, lang)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, receipt)
}

// StartSession begins tracking a kiosk user session.
// POST /api/v1/kiosk/sessions
func (h *KioskUXHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	var req struct {
		Lang string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Lang = "es"
	}

	session, err := h.kioskUXSvc.StartSession(r.Context(), kioskID, tenantID, req.Lang)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, session)
}

// EndSession marks a session as completed/abandoned.
// PUT /api/v1/kiosk/sessions/{id}/end
func (h *KioskUXHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		response.Error(w, http.StatusBadRequest, "session ID required")
		return
	}

	var req struct {
		Outcome   string `json:"outcome"` // completed, abandoned, timeout
		BookingID string `json:"booking_id,omitempty"`
		Steps     int    `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.kioskUXSvc.EndSession(r.Context(), sessionID, req.Outcome, req.BookingID, req.Steps); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "session ended"})
}

// --- Transport Card Endpoints ---

// CardBalance checks the balance of a transport card.
// GET /api/v1/kiosk/cards/{number}/balance
func (h *KioskUXHandler) CardBalance(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	cardNumber := r.PathValue("number")
	if cardNumber == "" {
		response.Error(w, http.StatusBadRequest, "card number required")
		return
	}

	card, err := h.kioskUXSvc.GetCardBalance(r.Context(), tenantID, cardNumber)
	if err != nil {
		response.Error(w, http.StatusNotFound, "card not found")
		return
	}

	response.JSON(w, http.StatusOK, card)
}

// RechargeCard adds balance to a transport card.
// POST /api/v1/kiosk/cards/recharge
func (h *KioskUXHandler) RechargeCard(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	var req domain.RechargeCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.ValidateRequired(req.CardNumber, "card_number"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.AmountCents <= 0 {
		response.Error(w, http.StatusBadRequest, "amount must be positive")
		return
	}

	result, err := h.kioskUXSvc.RechargeCard(r.Context(), tenantID, kioskID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// IssueCard creates a new transport card with optional initial balance.
// POST /api/v1/kiosk/cards/issue
func (h *KioskUXHandler) IssueCard(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "X-Kiosk-ID header required")
		return
	}

	var req struct {
		InitialBalance int64 `json:"initial_balance_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.InitialBalance = 0
	}

	card, err := h.kioskUXSvc.IssueCard(r.Context(), tenantID, kioskID, req.InitialBalance)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, card)
}

// Estimate returns a quick price estimate for the kiosk display.
// POST /api/v1/kiosk/estimate
func (h *KioskUXHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServiceType    string  `json:"service_type"`
		DropoffLat     float64 `json:"dropoff_lat"`
		DropoffLng     float64 `json:"dropoff_lng"`
		PassengerCount int     `json:"passenger_count"`
		KioskID        string  `json:"kiosk_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	kioskID := req.KioskID
	if kioskID == "" {
		kioskID = r.Header.Get("X-Kiosk-ID")
	}
	if kioskID == "" {
		response.Error(w, http.StatusBadRequest, "kiosk_id or X-Kiosk-ID header required")
		return
	}

	pax := req.PassengerCount
	if pax < 1 {
		pax = 1
	}

	// Get all service type estimates
	types := []domain.ServiceType{domain.ServiceTaxi, domain.ServiceShuttle, domain.ServiceVan, domain.ServiceBus}
	if req.ServiceType != "" {
		types = []domain.ServiceType{domain.ServiceType(req.ServiceType)}
	}

	type estimate struct {
		ServiceType string `json:"service_type"`
		PriceCents  int64  `json:"price_cents"`
		Currency    string `json:"currency"`
		ETAMinutes  int    `json:"eta_minutes"`
		Distance    string `json:"distance"`
	}

	recs, err := h.kioskUXSvc.RecommendService(r.Context(), pax, req.DropoffLat, req.DropoffLng, kioskID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	estimates := make([]estimate, 0, len(types))
	for _, rec := range recs {
		estimates = append(estimates, estimate{
			ServiceType: string(rec.ServiceType),
			PriceCents:  rec.PriceCents,
			Currency:    rec.Currency,
			ETAMinutes:  rec.ETAMinutes,
			Distance:    "",
		})
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"estimates":  estimates,
		"passengers": pax,
	})
}
