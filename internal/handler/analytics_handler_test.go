package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

func TestAnalyticsHandler_SLO(t *testing.T) {
	svc := service.NewAnalyticsService(nil)
	h := NewAnalyticsHandler(svc)

	req := httptest.NewRequest("GET", "/api/v1/analytics/slo", nil)
	rr := httptest.NewRecorder()

	h.SLO(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var env response.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !env.Success {
		t.Error("expected success=true")
	}
}
