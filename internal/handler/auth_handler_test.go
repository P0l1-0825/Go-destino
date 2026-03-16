package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/P0l1-0825/Go-destino/pkg/response"
)

func TestLogin_InvalidJSON(t *testing.T) {
	h := &AuthHandler{} // nil authSvc — we never reach it

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader("{bad json"))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestLogin_MissingEmail(t *testing.T) {
	h := &AuthHandler{}

	body := `{"email":"","password":"secret123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "email")
}

func TestLogin_MissingPassword(t *testing.T) {
	h := &AuthHandler{}

	body := `{"email":"test@example.com","password":""}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "password")
}

func TestLogin_InvalidEmail(t *testing.T) {
	h := &AuthHandler{}

	body := `{"email":"not-an-email","password":"secret123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "email")
}

func TestRegister_InvalidJSON(t *testing.T) {
	h := &AuthHandler{}

	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestRegister_MissingFields(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		errMsg string
	}{
		{"missing email", `{"email":"","password":"Secret123","name":"John"}`, "email"},
		{"invalid email", `{"email":"bad","password":"Secret123","name":"John"}`, "email"},
		{"short password", `{"email":"t@t.com","password":"ab","name":"John"}`, "password"},
		{"missing name", `{"email":"t@t.com","password":"Secret123","name":""}`, "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandler{}
			req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.Register(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
			assertErrorContains(t, rr, tt.errMsg)
		})
	}
}

func TestLogout_MissingAuthHeader(t *testing.T) {
	h := &AuthHandler{}

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestRefreshToken_InvalidJSON(t *testing.T) {
	h := &AuthHandler{}

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.RefreshToken(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestRefreshToken_EmptyToken(t *testing.T) {
	h := &AuthHandler{}

	body := `{"refresh_token":""}`
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.RefreshToken(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChangePassword_InvalidJSON(t *testing.T) {
	h := &AuthHandler{}

	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", strings.NewReader("!!"))
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChangePassword_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing old_password", `{"old_password":"","new_password":"NewPass123"}`},
		{"short new_password", `{"old_password":"OldPass123","new_password":"ab"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandler{}
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.ChangePassword(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestResetPassword_InvalidJSON(t *testing.T) {
	h := &AuthHandler{}

	req := httptest.NewRequest("POST", "/", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.ResetPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestResetPassword_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing token", `{"token":"","new_password":"NewPass123"}`},
		{"short password", `{"token":"abc123","new_password":"ab"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &AuthHandler{}
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.ResetPassword(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

func TestRequestPasswordReset_InvalidEmail(t *testing.T) {
	h := &AuthHandler{}

	body := `{"email":"not-valid"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.RequestPasswordReset(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid email, got %d", rr.Code)
	}
}

func TestMe_NoClaims(t *testing.T) {
	h := &AuthHandler{}

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		wantIP   string
	}{
		{
			"X-Forwarded-For single",
			map[string]string{"X-Forwarded-For": "1.2.3.4"},
			"127.0.0.1:8080",
			"1.2.3.4",
		},
		{
			"X-Forwarded-For multiple",
			map[string]string{"X-Forwarded-For": "1.2.3.4, 5.6.7.8"},
			"127.0.0.1:8080",
			"1.2.3.4",
		},
		{
			"X-Real-IP",
			map[string]string{"X-Real-IP": "10.0.0.1"},
			"127.0.0.1:8080",
			"10.0.0.1",
		},
		{
			"RemoteAddr with port",
			nil,
			"192.168.1.1:12345",
			"192.168.1.1",
		},
		{
			"RemoteAddr without port",
			nil,
			"192.168.1.1",
			"192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remote
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := clientIP(req)
			if got != tt.wantIP {
				t.Errorf("clientIP() = %q, want %q", got, tt.wantIP)
			}
		})
	}
}

// assertErrorContains checks the response body contains the expected error message.
func assertErrorContains(t *testing.T, rr *httptest.ResponseRecorder, msg string) {
	t.Helper()
	var resp response.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !strings.Contains(strings.ToLower(resp.Error), strings.ToLower(msg)) {
		t.Errorf("error should contain %q, got %q", msg, resp.Error)
	}
}
