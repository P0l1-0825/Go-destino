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

// KioskMonitorHandler serves the kiosk monitoring and remote support endpoints.
type KioskMonitorHandler struct {
	monitorSvc *service.KioskMonitorService
}

func NewKioskMonitorHandler(monitorSvc *service.KioskMonitorService) *KioskMonitorHandler {
	return &KioskMonitorHandler{monitorSvc: monitorSvc}
}

// HeartbeatFull receives extended heartbeat with full telemetry data.
// PUT /api/v1/monitor/kiosks/{id}/heartbeat
func (h *KioskMonitorHandler) HeartbeatFull(w http.ResponseWriter, r *http.Request) {
	kioskID := r.PathValue("id")
	tenantID := middleware.GetTenantID(r.Context())

	var hb domain.KioskHeartbeatFull
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	hb.KioskID = kioskID

	if err := h.monitorSvc.ProcessHeartbeat(r.Context(), kioskID, tenantID, hb); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return any pending commands for the kiosk to execute
	commands, _ := h.monitorSvc.GetPendingCommands(r.Context(), kioskID)
	if commands == nil {
		commands = []domain.KioskRemoteCommand{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":           "ok",
		"pending_commands": commands,
	})
}

// FleetDashboard returns the monitoring overview for all kiosks.
// GET /api/v1/monitor/dashboard
func (h *KioskMonitorHandler) FleetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	dash, err := h.monitorSvc.GetFleetDashboard(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, dash)
}

// Diagnostics returns a comprehensive diagnostic report for a specific kiosk.
// GET /api/v1/monitor/kiosks/{id}/diagnostics
func (h *KioskMonitorHandler) Diagnostics(w http.ResponseWriter, r *http.Request) {
	kioskID := r.PathValue("id")

	diag, err := h.monitorSvc.GetDiagnostics(r.Context(), kioskID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, diag)
}

// TelemetryHistory returns telemetry data points for charts.
// GET /api/v1/monitor/kiosks/{id}/telemetry?hours=24
func (h *KioskMonitorHandler) TelemetryHistory(w http.ResponseWriter, r *http.Request) {
	kioskID := r.PathValue("id")
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours <= 0 {
		hours = 24
	}

	data, err := h.monitorSvc.GetTelemetryHistory(r.Context(), kioskID, hours)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if data == nil {
		data = []domain.KioskTelemetry{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"kiosk_id": kioskID,
		"hours":    hours,
		"points":   len(data),
		"data":     data,
	})
}

// Alerts returns active alerts across all kiosks for a tenant.
// GET /api/v1/monitor/alerts
func (h *KioskMonitorHandler) Alerts(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	alerts, err := h.monitorSvc.GetAlerts(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if alerts == nil {
		alerts = []domain.KioskAlert{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// AckAlert acknowledges an alert.
// PUT /api/v1/monitor/alerts/{id}/ack
func (h *KioskMonitorHandler) AckAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("id")
	userID := middleware.GetUserID(r.Context())

	if err := h.monitorSvc.AckAlert(r.Context(), alertID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

// ResolveAlert resolves an alert.
// PUT /api/v1/monitor/alerts/{id}/resolve
func (h *KioskMonitorHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("id")

	if err := h.monitorSvc.ResolveAlert(r.Context(), alertID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

// Events returns recent events for a kiosk.
// GET /api/v1/monitor/kiosks/{id}/events?limit=50
func (h *KioskMonitorHandler) Events(w http.ResponseWriter, r *http.Request) {
	kioskID := r.PathValue("id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	events, err := h.monitorSvc.GetEvents(r.Context(), kioskID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if events == nil {
		events = []domain.KioskEvent{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"kiosk_id": kioskID,
		"events":   events,
		"total":    len(events),
	})
}

// EventsByTenant returns events across all kiosks.
// GET /api/v1/monitor/events?severity=critical&limit=100
func (h *KioskMonitorHandler) EventsByTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	severity := r.URL.Query().Get("severity")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	events, err := h.monitorSvc.GetEventsByTenant(r.Context(), tenantID, severity, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if events == nil {
		events = []domain.KioskEvent{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// SendCommand sends a remote command to a kiosk.
// POST /api/v1/monitor/kiosks/{id}/commands
func (h *KioskMonitorHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	kioskID := r.PathValue("id")
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req domain.SendRemoteCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Command == "" {
		response.Error(w, http.StatusBadRequest, "command is required")
		return
	}

	cmd, err := h.monitorSvc.SendCommand(r.Context(), kioskID, tenantID, userID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, cmd)
}

// CommandHistory returns command history for a kiosk.
// GET /api/v1/monitor/kiosks/{id}/commands?limit=50
func (h *KioskMonitorHandler) CommandHistory(w http.ResponseWriter, r *http.Request) {
	kioskID := r.PathValue("id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	cmds, err := h.monitorSvc.GetCommandHistory(r.Context(), kioskID, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cmds == nil {
		cmds = []domain.KioskRemoteCommand{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"kiosk_id": kioskID,
		"commands": cmds,
		"total":    len(cmds),
	})
}

// CommandResult receives execution result from the kiosk.
// PUT /api/v1/monitor/commands/{id}/result
func (h *KioskMonitorHandler) CommandResult(w http.ResponseWriter, r *http.Request) {
	cmdID := r.PathValue("id")

	var req struct {
		Status string `json:"status"` // executed, failed
		Result string `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.monitorSvc.ReportCommandResult(r.Context(), cmdID, req.Status, req.Result); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// StreamMonitor provides real-time SSE stream of kiosk monitoring events.
// GET /api/v1/monitor/stream
func (h *KioskMonitorHandler) StreamMonitor(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.monitorSvc.Subscribe(tenantID)
	defer h.monitorSvc.Unsubscribe(tenantID, ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			data, _ := json.Marshal(event)
			w.Write([]byte("event: " + event.EventType + "\n"))
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}
