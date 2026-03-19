package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotificationHandler_Send_InvalidJSON(t *testing.T) {
	h := &NotificationHandler{}
	req := httptest.NewRequest("POST", "/api/v1/notifications/send", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.Send(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}


func TestNotificationHandler_MarkRead_EmptyID(t *testing.T) {
	h := &NotificationHandler{}

	req := httptest.NewRequest("PUT", "/api/v1/notifications//read", nil)
	req.SetPathValue("id", "")
	rr := httptest.NewRecorder()

	h.MarkRead(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "id is required")
}

