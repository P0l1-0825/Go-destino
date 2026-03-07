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

type AIHandler struct {
	aiSvc *service.AIService
}

func NewAIHandler(aiSvc *service.AIService) *AIHandler {
	return &AIHandler{aiSvc: aiSvc}
}

func (h *AIHandler) DemandForecast(w http.ResponseWriter, r *http.Request) {
	airportID := r.URL.Query().Get("airport_id")
	interval := 30
	if v := r.URL.Query().Get("interval"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			interval = i
		}
	}

	forecasts, err := h.aiSvc.ForecastDemand(r.Context(), airportID, interval)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, forecasts)
}

func (h *AIHandler) DynamicPricing(w http.ResponseWriter, r *http.Request) {
	var req domain.DynamicPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	price, err := h.aiSvc.CalculateDynamicPrice(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, price)
}

func (h *AIHandler) FraudCheck(w http.ResponseWriter, r *http.Request) {
	var req domain.FraudCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.aiSvc.CheckFraud(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req domain.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.aiSvc.Chat(r.Context(), userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *AIHandler) VerifyBiometric(w http.ResponseWriter, r *http.Request) {
	var req domain.BiometricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.aiSvc.VerifyBiometric(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AIHandler) OptimizeRoutes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookingIDs []string `json:"booking_ids"`
		DriverID   string   `json:"driver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.aiSvc.OptimizeRoutes(r.Context(), req.BookingIDs, req.DriverID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, result)
}
