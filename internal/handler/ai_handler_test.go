package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIHandler_DynamicPricing_InvalidJSON(t *testing.T) {
	h := &AIHandler{}
	req := httptest.NewRequest("POST", "/api/v1/ai/pricing", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.DynamicPricing(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAIHandler_FraudCheck_InvalidJSON(t *testing.T) {
	h := &AIHandler{}
	req := httptest.NewRequest("POST", "/api/v1/ai/fraud", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.FraudCheck(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAIHandler_Chat_InvalidJSON(t *testing.T) {
	h := &AIHandler{}
	req := httptest.NewRequest("POST", "/api/v1/ai/chat", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.Chat(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAIHandler_VerifyBiometric_InvalidJSON(t *testing.T) {
	h := &AIHandler{}
	req := httptest.NewRequest("POST", "/api/v1/ai/biometric", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.VerifyBiometric(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}

func TestAIHandler_OptimizeRoutes_InvalidJSON(t *testing.T) {
	h := &AIHandler{}
	req := httptest.NewRequest("POST", "/api/v1/ai/routes", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()

	h.OptimizeRoutes(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	assertErrorContains(t, rr, "invalid request body")
}
