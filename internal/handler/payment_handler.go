package handler

import (
	"net/http"
	"strconv"

	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type PaymentHandler struct {
	paymentSvc *service.PaymentService
}

func NewPaymentHandler(paymentSvc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentSvc: paymentSvc}
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
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

	payments, err := h.paymentSvc.ListPayments(r.Context(), tenantID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *PaymentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}

	payment, err := h.paymentSvc.GetPayment(r.Context(), id, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "payment not found")
		return
	}
	response.JSON(w, http.StatusOK, payment)
}

func (h *PaymentHandler) GetByBooking(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	bookingID := r.PathValue("bookingId")
	if bookingID == "" {
		response.Error(w, http.StatusBadRequest, "bookingId is required")
		return
	}

	payment, err := h.paymentSvc.GetPaymentByBooking(r.Context(), bookingID, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "payment not found for booking")
		return
	}
	response.JSON(w, http.StatusOK, payment)
}

func (h *PaymentHandler) Refund(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())
	paymentID := r.PathValue("id")
	if paymentID == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}

	reason := r.URL.Query().Get("reason")
	if reason == "" {
		reason = "admin refund"
	}

	refund, err := h.paymentSvc.RefundPayment(r.Context(), paymentID, tenantID, userID, reason, "es")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, refund)
}
