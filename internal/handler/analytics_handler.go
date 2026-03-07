package handler

import (
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

type AnalyticsHandler struct {
	analyticsSvc *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsSvc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc}
}

func (h *AnalyticsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	airportID := r.URL.Query().Get("airport_id")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "today"
	}

	kpis, err := h.analyticsSvc.GetDashboardKPIs(r.Context(), tenantID, airportID, period)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, kpis)
}

func (h *AnalyticsHandler) Revenue(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	report, err := h.analyticsSvc.GetRevenueReport(r.Context(), tenantID, period)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, report)
}

func (h *AnalyticsHandler) BookingFunnel(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	funnel, err := h.analyticsSvc.GetBookingFunnel(r.Context(), tenantID, period)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, funnel)
}

func (h *AnalyticsHandler) SLO(w http.ResponseWriter, r *http.Request) {
	metrics := h.analyticsSvc.GetSLOMetrics()
	response.JSON(w, http.StatusOK, metrics)
}
