package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouteHandler_Create_InvalidJSON(t *testing.T) {
	h := &RouteHandler{}
	req := httptest.NewRequest("POST", "/api/v1/routes", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
	assertErrorContains(t, rr, "invalid request body")
}

func TestRouteHandler_Update_InvalidJSON(t *testing.T) {
	h := &RouteHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/v1/routes/{id}", h.Update)
	req := httptest.NewRequest("PUT", "/api/v1/routes/r1", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}

func TestRouteHandler_Delete_MissingID(t *testing.T) {
	h := &RouteHandler{}
	req := httptest.NewRequest("DELETE", "/api/v1/routes/", nil)
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	if rr.Code != http.StatusBadRequest { t.Errorf("status = %d, want 400", rr.Code) }
}
