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

type VoucherHandler struct {
	voucherSvc *service.VoucherService
}

func NewVoucherHandler(voucherSvc *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{voucherSvc: voucherSvc}
}

func (h *VoucherHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())
	var req domain.CreateVoucherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	voucher, err := h.voucherSvc.Create(r.Context(), tenantID, userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, voucher)
}

func (h *VoucherHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())
	var req domain.RedeemVoucherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.voucherSvc.Redeem(r.Context(), tenantID, userID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *VoucherHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	vouchers, err := h.voucherSvc.List(r.Context(), tenantID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, vouchers)
}

func (h *VoucherHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.Error(w, http.StatusBadRequest, "id is required")
		return
	}
	voucher, err := h.voucherSvc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "voucher not found")
		return
	}
	response.JSON(w, http.StatusOK, voucher)
}

func (h *VoucherHandler) GetByCode(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	code := r.PathValue("code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}
	voucher, err := h.voucherSvc.GetByCode(r.Context(), code, tenantID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "voucher not found")
		return
	}
	response.JSON(w, http.StatusOK, voucher)
}
