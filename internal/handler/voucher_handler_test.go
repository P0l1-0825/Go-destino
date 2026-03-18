package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVoucherHandler_Create_InvalidJSON(t *testing.T) {
	h := &VoucherHandler{}
	req := httptest.NewRequest("POST", "/api/v1/vouchers", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
	assertErrorContains(t, rr, "invalid request body")
}

func TestVoucherHandler_Redeem_InvalidJSON(t *testing.T) {
	h := &VoucherHandler{}
	req := httptest.NewRequest("POST", "/api/v1/vouchers/redeem", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()
	h.Redeem(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}

func TestVoucherHandler_GetByID_Missing(t *testing.T) {
	h := &VoucherHandler{}
	req := httptest.NewRequest("GET", "/api/v1/vouchers/", nil)
	rr := httptest.NewRecorder()
	h.GetByID(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}

func TestVoucherHandler_GetByCode_Missing(t *testing.T) {
	h := &VoucherHandler{}
	req := httptest.NewRequest("GET", "/api/v1/vouchers/code/", nil)
	rr := httptest.NewRecorder()
	h.GetByCode(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}
