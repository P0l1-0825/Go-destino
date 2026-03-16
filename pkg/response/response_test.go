package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"hello": "world"}

	JSON(rr, http.StatusOK, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestJSON_Created(t *testing.T) {
	rr := httptest.NewRecorder()
	JSON(rr, http.StatusCreated, "created")

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp APIResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Success {
		t.Error("201 should be success=true")
	}
}

func TestError_Response(t *testing.T) {
	rr := httptest.NewRecorder()
	Error(rr, http.StatusBadRequest, "something went wrong")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "something went wrong" {
		t.Errorf("error = %q, want %q", resp.Error, "something went wrong")
	}
}

func TestError_ServerError(t *testing.T) {
	rr := httptest.NewRecorder()
	Error(rr, http.StatusInternalServerError, "internal error")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}

	var resp APIResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Success {
		t.Error("500 should be success=false")
	}
}
