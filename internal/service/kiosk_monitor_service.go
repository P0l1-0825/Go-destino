package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
)

// KioskMonitorService provides comprehensive kiosk monitoring, alerting,
// remote support, and real-time status streaming.
type KioskMonitorService struct {
	monitorRepo *repository.KioskMonitorRepository
	kioskRepo   *repository.KioskRepository
	notifSvc    *NotificationService

	// SSE subscribers for real-time monitoring
	mu          sync.RWMutex
	subscribers map[string][]chan domain.KioskMonitorEvent // tenantID → channels
}

func NewKioskMonitorService(
	monitorRepo *repository.KioskMonitorRepository,
	kioskRepo *repository.KioskRepository,
	notifSvc *NotificationService,
) *KioskMonitorService {
	svc := &KioskMonitorService{
		monitorRepo: monitorRepo,
		kioskRepo:   kioskRepo,
		notifSvc:    notifSvc,
		subscribers: make(map[string][]chan domain.KioskMonitorEvent),
	}

	// Start background offline detector
	go svc.offlineDetectorLoop()

	// Start telemetry cleanup (keep 7 days)
	go svc.telemetryCleanupLoop()

	return svc
}

// ProcessHeartbeat handles an extended heartbeat, persists telemetry, and evaluates alerts.
func (s *KioskMonitorService) ProcessHeartbeat(ctx context.Context, kioskID, tenantID string, hb domain.KioskHeartbeatFull) error {
	// Update kiosk status to online
	_ = s.kioskRepo.UpdateHeartbeat(ctx, kioskID)

	// Persist telemetry
	telemetry := &domain.KioskTelemetry{
		ID:            uuid.New().String(),
		KioskID:       kioskID,
		TenantID:      tenantID,
		CPUPercent:    hb.CPUPercent,
		MemoryPercent: hb.MemoryPercent,
		DiskPercent:   hb.DiskPercent,
		Temperature:   hb.Temperature,
		PaperLevel:    hb.PaperLevel,
		PrinterOK:     hb.PrinterOK,
		ScannerOK:     hb.ScannerOK,
		NetworkType:   hb.NetworkType,
		NetworkMbps:   hb.NetworkMbps,
		UptimeSec:     hb.UptimeSec,
		AppVersion:    hb.AppVersion,
		OSVersion:     hb.OSVersion,
		ScreenOn:      hb.ScreenOn,
		ErrorCount:    hb.ErrorCount,
		CollectedAt:   time.Now(),
	}

	if err := s.monitorRepo.SaveTelemetry(ctx, telemetry); err != nil {
		log.Printf("[MONITOR] failed to save telemetry for %s: %v", kioskID, err)
	}

	// Auto-resolve offline alert if kiosk is back
	_ = s.monitorRepo.ResolveAlertsByType(ctx, kioskID, "offline")

	// Log online event
	s.logEvent(ctx, kioskID, tenantID, "heartbeat", "info", fmt.Sprintf("Heartbeat received: paper=%d%% temp=%.1f°C cpu=%.0f%%", hb.PaperLevel, hb.Temperature, hb.CPUPercent), "")

	// Evaluate alert rules
	s.evaluateAlerts(ctx, kioskID, tenantID, hb)

	// Broadcast to SSE subscribers
	s.broadcast(tenantID, domain.KioskMonitorEvent{
		EventType: "heartbeat",
		KioskID:   kioskID,
		Timestamp: time.Now(),
		Data:      telemetry,
	})

	return nil
}

// GetDiagnostics returns a comprehensive diagnostic report for a kiosk.
func (s *KioskMonitorService) GetDiagnostics(ctx context.Context, kioskID string) (*domain.KioskDiagnostics, error) {
	kiosk, err := s.kioskRepo.GetByID(ctx, kioskID)
	if err != nil {
		return nil, fmt.Errorf("kiosk not found: %w", err)
	}

	telemetry, _ := s.monitorRepo.GetLatestTelemetry(ctx, kioskID)
	alerts, _ := s.monitorRepo.GetActiveAlerts(ctx, kioskID)
	events, _ := s.monitorRepo.ListEvents(ctx, kioskID, 20)
	pending, _ := s.monitorRepo.GetPendingCommands(ctx, kioskID)
	sessionStats, _ := s.monitorRepo.SessionStatsForKiosk(ctx, kioskID)
	uptimePct, _ := s.monitorRepo.CalcUptimePercent(ctx, kioskID, 24)

	if alerts == nil {
		alerts = []domain.KioskAlert{}
	}
	if events == nil {
		events = []domain.KioskEvent{}
	}
	if pending == nil {
		pending = []domain.KioskRemoteCommand{}
	}
	if sessionStats == nil {
		sessionStats = &domain.KioskSessionStats{}
	}

	status := s.classifyHealth(kiosk, telemetry, alerts)
	healthScore := s.calculateHealthScore(kiosk, telemetry, alerts, uptimePct)

	return &domain.KioskDiagnostics{
		Kiosk:           *kiosk,
		Status:          status,
		LastTelemetry:   telemetry,
		ActiveAlerts:    alerts,
		RecentEvents:    events,
		PendingCommands: pending,
		SessionStats:    *sessionStats,
		UptimePercent:   uptimePct,
		HealthScore:     healthScore,
	}, nil
}

// GetFleetDashboard returns the monitoring overview for all kiosks of a tenant.
func (s *KioskMonitorService) GetFleetDashboard(ctx context.Context, tenantID string) (*domain.KioskFleetDashboard, error) {
	kiosks, err := s.kioskRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing kiosks: %w", err)
	}

	critAlerts, warnAlerts, _ := s.monitorRepo.CountActiveAlerts(ctx, tenantID)

	dash := &domain.KioskFleetDashboard{
		TotalKiosks:    len(kiosks),
		CriticalAlerts: critAlerts,
		WarningAlerts:  warnAlerts,
		Kiosks:         make([]domain.KioskSummary, 0, len(kiosks)),
		GeneratedAt:    time.Now(),
	}

	totalScore := 0
	for _, k := range kiosks {
		telemetry, _ := s.monitorRepo.GetLatestTelemetry(ctx, k.ID)
		alerts, _ := s.monitorRepo.GetActiveAlerts(ctx, k.ID)
		sessionsToday, _ := s.monitorRepo.CountSessionsToday(ctx, k.ID)

		if alerts == nil {
			alerts = []domain.KioskAlert{}
		}

		uptimePct, _ := s.monitorRepo.CalcUptimePercent(ctx, k.ID, 24)
		healthScore := s.calculateHealthScore(&k, telemetry, alerts, uptimePct)
		totalScore += healthScore

		summary := domain.KioskSummary{
			ID:            k.ID,
			Name:          k.Name,
			Location:      k.Location,
			AirportID:     k.AirportID,
			TerminalID:    k.TerminalID,
			Status:        string(k.Status),
			HealthScore:   healthScore,
			ActiveAlerts:  len(alerts),
			LastHeartbeat: k.LastHeartbeat,
			SessionsToday: sessionsToday,
		}

		if telemetry != nil {
			summary.PaperLevel = telemetry.PaperLevel
			summary.Temperature = telemetry.Temperature
			summary.AppVersion = telemetry.AppVersion
			summary.UptimeSec = telemetry.UptimeSec
		}

		dash.Kiosks = append(dash.Kiosks, summary)

		switch k.Status {
		case domain.KioskOnline:
			if healthScore < 50 {
				dash.DegradedCount++
			} else {
				dash.OnlineCount++
			}
		case domain.KioskOffline:
			dash.OfflineCount++
		case domain.KioskMaintenance:
			dash.MaintenanceCount++
		}
	}

	if len(kiosks) > 0 {
		dash.AvgHealthScore = totalScore / len(kiosks)
	}

	return dash, nil
}

// GetTelemetryHistory returns telemetry data points for charts.
func (s *KioskMonitorService) GetTelemetryHistory(ctx context.Context, kioskID string, hours int) ([]domain.KioskTelemetry, error) {
	if hours <= 0 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	return s.monitorRepo.GetTelemetryHistory(ctx, kioskID, since, 500)
}

// GetAlerts returns active alerts for a tenant.
func (s *KioskMonitorService) GetAlerts(ctx context.Context, tenantID string) ([]domain.KioskAlert, error) {
	return s.monitorRepo.GetActiveAlertsByTenant(ctx, tenantID)
}

// AckAlert acknowledges an alert.
func (s *KioskMonitorService) AckAlert(ctx context.Context, alertID, userID string) error {
	return s.monitorRepo.AckAlert(ctx, alertID, userID)
}

// ResolveAlert resolves an alert.
func (s *KioskMonitorService) ResolveAlert(ctx context.Context, alertID string) error {
	return s.monitorRepo.ResolveAlert(ctx, alertID)
}

// GetEvents returns recent events for a kiosk.
func (s *KioskMonitorService) GetEvents(ctx context.Context, kioskID string, limit int) ([]domain.KioskEvent, error) {
	return s.monitorRepo.ListEvents(ctx, kioskID, limit)
}

// GetEventsByTenant returns events across all kiosks for a tenant.
func (s *KioskMonitorService) GetEventsByTenant(ctx context.Context, tenantID, severity string, limit int) ([]domain.KioskEvent, error) {
	return s.monitorRepo.ListEventsByTenant(ctx, tenantID, severity, limit)
}

// SendCommand sends a remote command to a kiosk.
func (s *KioskMonitorService) SendCommand(ctx context.Context, kioskID, tenantID, userID string, req domain.SendRemoteCommandRequest) (*domain.KioskRemoteCommand, error) {
	if !domain.ValidRemoteCommands[req.Command] {
		return nil, fmt.Errorf("unknown command: %s", req.Command)
	}

	cmd := &domain.KioskRemoteCommand{
		ID:       uuid.New().String(),
		KioskID:  kioskID,
		TenantID: tenantID,
		Command:  req.Command,
		Params:   req.Params,
		Status:   "pending",
		IssuedBy: userID,
		IssuedAt: time.Now(),
	}

	if err := s.monitorRepo.CreateCommand(ctx, cmd); err != nil {
		return nil, fmt.Errorf("creating command: %w", err)
	}

	// Log event
	s.logEvent(ctx, kioskID, tenantID, "command_issued", "info",
		fmt.Sprintf("Remote command '%s' issued by %s", req.Command, userID), req.Params)

	// If it's set_maintenance or clear_maintenance, apply immediately
	switch req.Command {
	case "set_maintenance":
		_ = s.kioskRepo.UpdateStatus(ctx, kioskID, domain.KioskMaintenance)
		s.logEvent(ctx, kioskID, tenantID, "maintenance_start", "info", "Kiosk set to maintenance mode", "")
	case "clear_maintenance":
		_ = s.kioskRepo.UpdateStatus(ctx, kioskID, domain.KioskOnline)
		s.logEvent(ctx, kioskID, tenantID, "maintenance_end", "info", "Kiosk returned to online mode", "")
	}

	// Broadcast command to SSE for real-time pickup
	s.broadcast(tenantID, domain.KioskMonitorEvent{
		EventType: "command",
		KioskID:   kioskID,
		Timestamp: time.Now(),
		Data:      cmd,
	})

	return cmd, nil
}

// ReportCommandResult updates a command's result (called by the kiosk after execution).
func (s *KioskMonitorService) ReportCommandResult(ctx context.Context, cmdID, status, result string) error {
	return s.monitorRepo.UpdateCommandStatus(ctx, cmdID, status, result)
}

// GetPendingCommands returns pending commands for a kiosk (polled by the kiosk client).
func (s *KioskMonitorService) GetPendingCommands(ctx context.Context, kioskID string) ([]domain.KioskRemoteCommand, error) {
	return s.monitorRepo.GetPendingCommands(ctx, kioskID)
}

// GetCommandHistory returns command history for a kiosk.
func (s *KioskMonitorService) GetCommandHistory(ctx context.Context, kioskID string, limit int) ([]domain.KioskRemoteCommand, error) {
	return s.monitorRepo.ListCommands(ctx, kioskID, limit)
}

// --- SSE Subscriptions ---

func (s *KioskMonitorService) Subscribe(tenantID string) chan domain.KioskMonitorEvent {
	ch := make(chan domain.KioskMonitorEvent, 20)
	s.mu.Lock()
	s.subscribers[tenantID] = append(s.subscribers[tenantID], ch)
	s.mu.Unlock()
	return ch
}

func (s *KioskMonitorService) Unsubscribe(tenantID string, ch chan domain.KioskMonitorEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := s.subscribers[tenantID]
	for i, sub := range subs {
		if sub == ch {
			s.subscribers[tenantID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(s.subscribers[tenantID]) == 0 {
		delete(s.subscribers, tenantID)
	}
}

func (s *KioskMonitorService) broadcast(tenantID string, event domain.KioskMonitorEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.subscribers[tenantID] {
		select {
		case ch <- event:
		default:
			// Drop if subscriber is slow
		}
	}
}

// --- Alert Evaluation ---

func (s *KioskMonitorService) evaluateAlerts(ctx context.Context, kioskID, tenantID string, hb domain.KioskHeartbeatFull) {
	// Paper low (< 15%)
	if hb.PaperLevel < 15 {
		s.raiseAlertIfNew(ctx, kioskID, tenantID, "paper_low", "warning",
			fmt.Sprintf("Paper level critically low: %d%%", hb.PaperLevel))
	} else if hb.PaperLevel >= 30 {
		_ = s.monitorRepo.ResolveAlertsByType(ctx, kioskID, "paper_low")
	}

	// High temperature (> 65°C)
	if hb.Temperature > 65 {
		sev := "warning"
		if hb.Temperature > 80 {
			sev = "critical"
		}
		s.raiseAlertIfNew(ctx, kioskID, tenantID, "high_temp", sev,
			fmt.Sprintf("Temperature alert: %.1f°C", hb.Temperature))
	} else if hb.Temperature <= 60 {
		_ = s.monitorRepo.ResolveAlertsByType(ctx, kioskID, "high_temp")
	}

	// Printer error
	if !hb.PrinterOK {
		s.raiseAlertIfNew(ctx, kioskID, tenantID, "printer_error", "critical",
			"Printer is not responding")
	} else {
		_ = s.monitorRepo.ResolveAlertsByType(ctx, kioskID, "printer_error")
	}

	// Scanner error
	if !hb.ScannerOK {
		s.raiseAlertIfNew(ctx, kioskID, tenantID, "scanner_error", "warning",
			"QR/NFC scanner is not responding")
	} else {
		_ = s.monitorRepo.ResolveAlertsByType(ctx, kioskID, "scanner_error")
	}

	// Disk full (> 90%)
	if hb.DiskPercent > 90 {
		sev := "warning"
		if hb.DiskPercent > 95 {
			sev = "critical"
		}
		s.raiseAlertIfNew(ctx, kioskID, tenantID, "disk_full", sev,
			fmt.Sprintf("Disk usage high: %.1f%%", hb.DiskPercent))
	} else if hb.DiskPercent <= 85 {
		_ = s.monitorRepo.ResolveAlertsByType(ctx, kioskID, "disk_full")
	}

	// High error rate
	if hb.ErrorCount > 10 {
		s.raiseAlertIfNew(ctx, kioskID, tenantID, "high_error_rate", "warning",
			fmt.Sprintf("High error count since last heartbeat: %d errors", hb.ErrorCount))
	}

	// Slow network (< 1 Mbps)
	if hb.NetworkMbps > 0 && hb.NetworkMbps < 1.0 {
		s.raiseAlertIfNew(ctx, kioskID, tenantID, "slow_network", "warning",
			fmt.Sprintf("Slow network: %.2f Mbps", hb.NetworkMbps))
	} else if hb.NetworkMbps >= 5.0 {
		_ = s.monitorRepo.ResolveAlertsByType(ctx, kioskID, "slow_network")
	}
}

func (s *KioskMonitorService) raiseAlertIfNew(ctx context.Context, kioskID, tenantID, alertType, severity, message string) {
	exists, _ := s.monitorRepo.HasActiveAlert(ctx, kioskID, alertType)
	if exists {
		return
	}

	alert := &domain.KioskAlert{
		ID:        uuid.New().String(),
		KioskID:   kioskID,
		TenantID:  tenantID,
		AlertType: alertType,
		Severity:  severity,
		Message:   message,
		Active:    true,
		CreatedAt: time.Now(),
	}

	if err := s.monitorRepo.CreateAlert(ctx, alert); err != nil {
		log.Printf("[MONITOR] failed to create alert: %v", err)
		return
	}

	// Log the event
	s.logEvent(ctx, kioskID, tenantID, "alert_raised", severity, message, alertType)

	// Broadcast to SSE
	s.broadcast(tenantID, domain.KioskMonitorEvent{
		EventType: "alert",
		KioskID:   kioskID,
		Timestamp: time.Now(),
		Data:      alert,
	})

	// Send notification for critical alerts
	if severity == "critical" {
		go func() {
			_, _ = s.notifSvc.Send(context.Background(), tenantID, domain.SendNotificationRequest{
				Channel: domain.ChannelPush,
				Title:   fmt.Sprintf("CRITICAL: Kiosk %s", alertType),
				Body:    message,
				Data:    map[string]string{"type": "kiosk_alert", "kiosk_id": kioskID, "alert_type": alertType},
			})
		}()
	}
}

// --- Background Loops ---

func (s *KioskMonitorService) offlineDetectorLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.detectOfflineKiosks()
	}
}

func (s *KioskMonitorService) detectOfflineKiosks() {
	ctx := context.Background()
	// We can't easily iterate all tenants without a tenant repo, so we rely on
	// the kiosk heartbeat timestamp. Any kiosk that hasn't heartbeated in 5 min
	// is considered offline.
	// This is done at the DB level by checking last_heartbeat.

	// For now, log a check. In production, run:
	// UPDATE kiosks SET status = 'offline' WHERE status = 'online'
	//   AND last_heartbeat < NOW() - INTERVAL '5 minutes'
	// Then create alerts for each.
	log.Printf("[MONITOR] Running offline detection check at %s", time.Now().Format(time.RFC3339))
	_ = ctx
}

func (s *KioskMonitorService) telemetryCleanupLoop() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		ctx := context.Background()
		deleted, err := s.monitorRepo.CleanupOldTelemetry(ctx, 7*24*time.Hour)
		if err != nil {
			log.Printf("[MONITOR] telemetry cleanup error: %v", err)
		} else if deleted > 0 {
			log.Printf("[MONITOR] cleaned up %d old telemetry records", deleted)
		}
	}
}

// --- Helpers ---

func (s *KioskMonitorService) logEvent(ctx context.Context, kioskID, tenantID, eventType, severity, message, details string) {
	event := &domain.KioskEvent{
		ID:        uuid.New().String(),
		KioskID:   kioskID,
		TenantID:  tenantID,
		EventType: eventType,
		Severity:  severity,
		Message:   message,
		Details:   details,
		CreatedAt: time.Now(),
	}
	if err := s.monitorRepo.CreateEvent(ctx, event); err != nil {
		log.Printf("[MONITOR] failed to log event: %v", err)
	}
}

func (s *KioskMonitorService) classifyHealth(kiosk *domain.Kiosk, telemetry *domain.KioskTelemetry, alerts []domain.KioskAlert) string {
	if kiosk.Status == domain.KioskOffline {
		return "offline"
	}
	if kiosk.Status == domain.KioskMaintenance {
		return "maintenance"
	}

	// Check if heartbeat is stale (> 5 min)
	if time.Since(kiosk.LastHeartbeat) > 5*time.Minute {
		return "offline"
	}

	for _, a := range alerts {
		if a.Severity == "critical" {
			return "critical"
		}
	}

	if len(alerts) > 0 {
		return "degraded"
	}

	return "healthy"
}

func (s *KioskMonitorService) calculateHealthScore(kiosk *domain.Kiosk, telemetry *domain.KioskTelemetry, alerts []domain.KioskAlert, uptimePct float64) int {
	score := 100

	// Offline = 0
	if kiosk.Status == domain.KioskOffline || time.Since(kiosk.LastHeartbeat) > 5*time.Minute {
		return 0
	}

	// Alerts penalty
	for _, a := range alerts {
		if a.Severity == "critical" {
			score -= 30
		} else {
			score -= 10
		}
	}

	if telemetry != nil {
		// Paper level
		if telemetry.PaperLevel < 10 {
			score -= 15
		} else if telemetry.PaperLevel < 25 {
			score -= 5
		}

		// Temperature
		if telemetry.Temperature > 75 {
			score -= 20
		} else if telemetry.Temperature > 65 {
			score -= 10
		}

		// Peripherals
		if !telemetry.PrinterOK {
			score -= 25
		}
		if !telemetry.ScannerOK {
			score -= 15
		}

		// Disk
		if telemetry.DiskPercent > 95 {
			score -= 15
		} else if telemetry.DiskPercent > 90 {
			score -= 5
		}

		// Network
		if telemetry.NetworkMbps > 0 && telemetry.NetworkMbps < 1.0 {
			score -= 10
		}
	}

	// Uptime penalty
	if uptimePct < 90 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}
	return score
}
