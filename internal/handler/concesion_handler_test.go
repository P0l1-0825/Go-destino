package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConcesionHandler_Create_InvalidJSON(t *testing.T) {
	h := &ConcesionHandler{}
	req := httptest.NewRequest("POST", "/api/v1/concesiones", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestConcesionHandler_Create_MissingName(t *testing.T) {
	h := &ConcesionHandler{}
	body := `{"code":"CONC-001"}`
	req := httptest.NewRequest("POST", "/api/v1/concesiones", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "name is required")
}

func TestConcesionHandler_Create_MissingCode(t *testing.T) {
	h := &ConcesionHandler{}
	body := `{"name":"Test Concesion"}`
	req := httptest.NewRequest("POST", "/api/v1/concesiones", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "code is required")
}

func TestConcesionHandler_Update_InvalidJSON(t *testing.T) {
	h := &ConcesionHandler{}
	req := httptest.NewRequest("PUT", "/api/v1/concesiones/123", strings.NewReader("{bad"))
	req.SetPathValue("id", "123")
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestConcesionHandler_AssignStaff_InvalidJSON(t *testing.T) {
	h := &ConcesionHandler{}
	req := httptest.NewRequest("POST", "/api/v1/concesiones/123/staff", strings.NewReader("{bad"))
	req.SetPathValue("id", "123")
	rr := httptest.NewRecorder()

	h.AssignStaff(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestConcesionHandler_AssignStaff_MissingUserID(t *testing.T) {
	h := &ConcesionHandler{}
	body := `{"staff_role":"operativo"}`
	req := httptest.NewRequest("POST", "/api/v1/concesiones/123/staff", strings.NewReader(body))
	req.SetPathValue("id", "123")
	rr := httptest.NewRecorder()

	h.AssignStaff(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "user_id is required")
}
