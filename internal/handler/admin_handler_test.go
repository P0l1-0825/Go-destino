package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminHandler_CreateTenant_InvalidJSON(t *testing.T) {
	h := &AdminHandler{}
	req := httptest.NewRequest("POST", "/api/v1/admin/tenants", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.CreateTenant(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAdminHandler_CreateUser_InvalidJSON(t *testing.T) {
	h := &AdminHandler{}
	req := httptest.NewRequest("POST", "/api/v1/admin/users", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAdminHandler_CreateUser_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"missing email", `{"password":"Pass1","name":"Test"}`, "email, password and name are required"},
		{"missing password", `{"email":"a@b.com","name":"Test"}`, "email, password and name are required"},
		{"missing name", `{"email":"a@b.com","password":"Pass1"}`, "email, password and name are required"},
		{"all empty", `{}`, "email, password and name are required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &AdminHandler{}
			req := httptest.NewRequest("POST", "/api/v1/admin/users", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()

			h.CreateUser(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
			assertErrorContains(t, rr, tc.want)
		})
	}
}

// Note: UpdateUser calls repo.GetByIDTenant before decoding body,
// so input validation tests require a mock repo. Covered by integration tests.

func TestAdminHandler_UpdateUserRole_InvalidJSON(t *testing.T) {
	h := &AdminHandler{}
	req := httptest.NewRequest("PUT", "/api/v1/admin/users/123/role", strings.NewReader("{bad"))
	req.SetPathValue("id", "123")
	rr := httptest.NewRecorder()

	h.UpdateUserRole(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAdminHandler_UpdateUserRole_MissingRole(t *testing.T) {
	h := &AdminHandler{}
	req := httptest.NewRequest("PUT", "/api/v1/admin/users/123/role", strings.NewReader(`{}`))
	req.SetPathValue("id", "123")
	rr := httptest.NewRecorder()

	h.UpdateUserRole(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "role is required")
}

func TestAdminHandler_CreateAirport_InvalidJSON(t *testing.T) {
	h := &AdminHandler{}
	req := httptest.NewRequest("POST", "/api/v1/admin/airports", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.CreateAirport(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAdminHandler_ListRoles(t *testing.T) {
	h := &AdminHandler{}
	req := httptest.NewRequest("GET", "/api/v1/admin/roles", nil)
	rr := httptest.NewRecorder()

	h.ListRoles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "SUPER_ADMIN") {
		t.Error("response should include SUPER_ADMIN role")
	}
}

func TestAdminHandler_ListPermissions(t *testing.T) {
	h := &AdminHandler{}
	req := httptest.NewRequest("GET", "/api/v1/admin/permissions", nil)
	rr := httptest.NewRecorder()

	h.ListPermissions(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "concesion.read") {
		t.Error("response should include concesion.read permission")
	}
}
