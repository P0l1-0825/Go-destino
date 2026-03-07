package handler

import (
	"encoding/json"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type TicketHandler struct {
	ticketSvc *service.TicketService
}

func NewTicketHandler(ticketSvc *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketSvc: ticketSvc}
}

func (h *TicketHandler) Purchase(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	kioskID := r.Header.Get("X-Kiosk-ID")

	var req domain.PurchaseTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	result, err := h.ticketSvc.PurchaseTickets(r.Context(), tenantID, kioskID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *TicketHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QRCode string `json:"qr_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ticket, err := h.ticketSvc.ValidateTicket(r.Context(), req.QRCode)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ticket, err := h.ticketSvc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "ticket not found")
		return
	}
	response.JSON(w, http.StatusOK, ticket)
}
