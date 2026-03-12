package handler

import (
	"encoding/json"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type QRHandler struct {
	bookingSvc *service.BookingService
	ticketSvc  *service.TicketService
}

func NewQRHandler(bookingSvc *service.BookingService, ticketSvc *service.TicketService) *QRHandler {
	return &QRHandler{bookingSvc: bookingSvc, ticketSvc: ticketSvc}
}

// QRValidateRequest is the request body for QR validation.
type QRValidateRequest struct {
	Code string `json:"code"`
}

// QRValidateResponse is the response for QR validation.
type QRValidateResponse struct {
	Type    string      `json:"type"`    // "booking" or "ticket"
	Valid   bool        `json:"valid"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Validate handles POST /api/v1/qr/validate
// It accepts a QR code (booking number or ticket QR code) and validates it.
func (h *QRHandler) Validate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req QRValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	// Try as booking number first (GD-XXXXXXXX format, 8 chars)
	if tenantID != "" {
		booking, err := h.bookingSvc.GetByNumberTenant(r.Context(), req.Code, tenantID)
		if err == nil && booking != nil {
			// Found as booking — validate status
			validStatuses := map[domain.BookingStatus]bool{
				domain.BookingPending:   true,
				domain.BookingConfirmed: true,
				domain.BookingAssigned:  true,
			}

			if !validStatuses[booking.Status] {
				response.JSON(w, http.StatusOK, QRValidateResponse{
					Type:    "booking",
					Valid:   false,
					Message: "Reserva en estado " + string(booking.Status) + " — no se puede validar",
					Data:    booking,
				})
				return
			}

			response.JSON(w, http.StatusOK, QRValidateResponse{
				Type:    "booking",
				Valid:   true,
				Message: "Reserva válida",
				Data:    booking,
			})
			return
		}
	}

	// Fallback: try without tenant filter
	booking, err := h.bookingSvc.GetByNumber(r.Context(), req.Code)
	if err == nil && booking != nil {
		validStatuses := map[domain.BookingStatus]bool{
			domain.BookingPending:   true,
			domain.BookingConfirmed: true,
			domain.BookingAssigned:  true,
		}

		if !validStatuses[booking.Status] {
			response.JSON(w, http.StatusOK, QRValidateResponse{
				Type:    "booking",
				Valid:   false,
				Message: "Reserva en estado " + string(booking.Status) + " — no se puede validar",
				Data:    booking,
			})
			return
		}

		response.JSON(w, http.StatusOK, QRValidateResponse{
			Type:    "booking",
			Valid:   true,
			Message: "Reserva válida",
			Data:    booking,
		})
		return
	}

	// Try as ticket QR code (longer hex format)
	ticket, err := h.ticketSvc.ValidateTicket(r.Context(), req.Code)
	if err == nil && ticket != nil {
		response.JSON(w, http.StatusOK, QRValidateResponse{
			Type:    "ticket",
			Valid:   true,
			Message: "Ticket válido — marcado como usado",
			Data:    ticket,
		})
		return
	}

	// Nothing found
	response.JSON(w, http.StatusOK, QRValidateResponse{
		Type:    "",
		Valid:   false,
		Message: "Código QR no encontrado — verifique e intente de nuevo",
	})
}

// Lookup handles GET /api/v1/qr/lookup/{code}
// It looks up a booking or ticket by code without modifying state.
func (h *QRHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	// Try as booking number
	booking, err := h.bookingSvc.GetByNumber(r.Context(), code)
	if err == nil && booking != nil {
		response.JSON(w, http.StatusOK, QRValidateResponse{
			Type:    "booking",
			Valid:   booking.Status == domain.BookingConfirmed || booking.Status == domain.BookingAssigned || booking.Status == domain.BookingPending,
			Message: "Reserva encontrada",
			Data:    booking,
		})
		return
	}

	response.JSON(w, http.StatusOK, QRValidateResponse{
		Type:    "",
		Valid:   false,
		Message: "Código no encontrado",
	})
}
