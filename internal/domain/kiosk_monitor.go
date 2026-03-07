package domain

import "time"

// --- Kiosk Monitoring & Remote Support Models ---

// KioskTelemetry stores a snapshot of kiosk hardware metrics.
type KioskTelemetry struct {
	ID            string    `json:"id" db:"id"`
	KioskID       string    `json:"kiosk_id" db:"kiosk_id"`
	TenantID      string    `json:"tenant_id" db:"tenant_id"`
	CPUPercent    float64   `json:"cpu_percent" db:"cpu_percent"`
	MemoryPercent float64   `json:"memory_percent" db:"memory_percent"`
	DiskPercent   float64   `json:"disk_percent" db:"disk_percent"`
	Temperature   float64   `json:"temperature" db:"temperature"`     // Celsius
	PaperLevel    int       `json:"paper_level" db:"paper_level"`     // 0-100%
	PrinterOK     bool      `json:"printer_ok" db:"printer_ok"`
	ScannerOK     bool      `json:"scanner_ok" db:"scanner_ok"`
	NetworkType   string    `json:"network_type" db:"network_type"`   // ethernet, wifi, 4g
	NetworkMbps   float64   `json:"network_mbps" db:"network_mbps"`
	UptimeSec     int64     `json:"uptime_seconds" db:"uptime_sec"`
	AppVersion    string    `json:"app_version" db:"app_version"`
	OSVersion     string    `json:"os_version" db:"os_version"`
	ScreenOn      bool      `json:"screen_on" db:"screen_on"`
	ErrorCount    int       `json:"error_count" db:"error_count"`     // Errors since last heartbeat
	CollectedAt   time.Time `json:"collected_at" db:"collected_at"`
}

// KioskHeartbeatFull is the extended heartbeat payload sent by kiosks.
type KioskHeartbeatFull struct {
	KioskID       string  `json:"kiosk_id"`
	Status        string  `json:"status"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskPercent   float64 `json:"disk_percent"`
	Temperature   float64 `json:"temperature"`
	PaperLevel    int     `json:"paper_level"`
	PrinterOK     bool    `json:"printer_ok"`
	ScannerOK     bool    `json:"scanner_ok"`
	NetworkType   string  `json:"network_type"`
	NetworkMbps   float64 `json:"network_mbps"`
	UptimeSec     int64   `json:"uptime_seconds"`
	AppVersion    string  `json:"app_version"`
	OSVersion     string  `json:"os_version"`
	ScreenOn      bool    `json:"screen_on"`
	ErrorCount    int     `json:"error_count"`
}

// KioskEvent represents a significant event from a kiosk.
type KioskEvent struct {
	ID        string    `json:"id" db:"id"`
	KioskID   string    `json:"kiosk_id" db:"kiosk_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	EventType string    `json:"event_type" db:"event_type"` // boot, shutdown, error, paper_low, paper_replaced, offline, online, maintenance_start, maintenance_end, command_executed
	Severity  string    `json:"severity" db:"severity"`     // info, warning, critical
	Message   string    `json:"message" db:"message"`
	Details   string    `json:"details,omitempty" db:"details"` // JSON extra data
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// KioskAlert represents an active monitoring alert.
type KioskAlert struct {
	ID          string     `json:"id" db:"id"`
	KioskID     string     `json:"kiosk_id" db:"kiosk_id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	AlertType   string     `json:"alert_type" db:"alert_type"`     // offline, paper_low, high_temp, printer_error, scanner_error, disk_full, high_error_rate, slow_network
	Severity    string     `json:"severity" db:"severity"`         // warning, critical
	Message     string     `json:"message" db:"message"`
	Active      bool       `json:"active" db:"active"`
	AckedBy     string     `json:"acked_by,omitempty" db:"acked_by"`
	AckedAt     *time.Time `json:"acked_at,omitempty" db:"acked_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// KioskRemoteCommand represents a command sent to a kiosk for remote support.
type KioskRemoteCommand struct {
	ID         string     `json:"id" db:"id"`
	KioskID    string     `json:"kiosk_id" db:"kiosk_id"`
	TenantID   string     `json:"tenant_id" db:"tenant_id"`
	Command    string     `json:"command" db:"command"`       // reboot, restart_app, update_config, set_maintenance, clear_maintenance, screenshot, run_diagnostic, clear_cache
	Params     string     `json:"params,omitempty" db:"params"` // JSON params
	Status     string     `json:"status" db:"status"`         // pending, sent, executed, failed
	IssuedBy   string     `json:"issued_by" db:"issued_by"`   // User who issued it
	Result     string     `json:"result,omitempty" db:"result"` // Execution result
	IssuedAt   time.Time  `json:"issued_at" db:"issued_at"`
	ExecutedAt *time.Time `json:"executed_at,omitempty" db:"executed_at"`
}

// KioskDiagnostics is a comprehensive diagnostic report for a single kiosk.
type KioskDiagnostics struct {
	Kiosk           Kiosk               `json:"kiosk"`
	Status          string              `json:"status"`           // healthy, degraded, critical, offline
	LastTelemetry   *KioskTelemetry     `json:"last_telemetry"`
	ActiveAlerts    []KioskAlert        `json:"active_alerts"`
	RecentEvents    []KioskEvent        `json:"recent_events"`
	PendingCommands []KioskRemoteCommand `json:"pending_commands"`
	SessionStats    KioskSessionStats   `json:"session_stats"`
	UptimePercent   float64             `json:"uptime_percent"`   // Last 24h
	HealthScore     int                 `json:"health_score"`     // 0-100
}

// KioskSessionStats aggregates session data for a kiosk.
type KioskSessionStats struct {
	TotalSessions    int     `json:"total_sessions"`
	CompletedCount   int     `json:"completed"`
	AbandonedCount   int     `json:"abandoned"`
	TimeoutCount     int     `json:"timeout"`
	CompletionRate   float64 `json:"completion_rate"`   // 0-100
	AvgDurationMs    int64   `json:"avg_duration_ms"`
}

// KioskFleetDashboard is the monitoring overview for all kiosks.
type KioskFleetDashboard struct {
	TotalKiosks      int              `json:"total_kiosks"`
	OnlineCount      int              `json:"online"`
	OfflineCount     int              `json:"offline"`
	MaintenanceCount int              `json:"maintenance"`
	DegradedCount    int              `json:"degraded"`
	CriticalAlerts   int              `json:"critical_alerts"`
	WarningAlerts    int              `json:"warning_alerts"`
	AvgHealthScore   int              `json:"avg_health_score"`
	Kiosks           []KioskSummary   `json:"kiosks"`
	GeneratedAt      time.Time        `json:"generated_at"`
}

// KioskSummary is a compact status for listing kiosks in the dashboard.
type KioskSummary struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	AirportID      string    `json:"airport_id"`
	TerminalID     string    `json:"terminal_id"`
	Status         string    `json:"status"`
	HealthScore    int       `json:"health_score"`
	PaperLevel     int       `json:"paper_level"`
	Temperature    float64   `json:"temperature"`
	AppVersion     string    `json:"app_version"`
	UptimeSec      int64     `json:"uptime_seconds"`
	ActiveAlerts   int       `json:"active_alerts"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	SessionsToday  int       `json:"sessions_today"`
}

// KioskMonitorEvent is published via SSE for real-time monitoring.
type KioskMonitorEvent struct {
	EventType string      `json:"event_type"` // heartbeat, alert, event, command_result
	KioskID   string      `json:"kiosk_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// SendRemoteCommandRequest is the API payload for sending a command.
type SendRemoteCommandRequest struct {
	Command string `json:"command"`
	Params  string `json:"params,omitempty"`
}

// Valid remote commands
var ValidRemoteCommands = map[string]bool{
	"reboot":            true,
	"restart_app":       true,
	"update_config":     true,
	"set_maintenance":   true,
	"clear_maintenance": true,
	"screenshot":        true,
	"run_diagnostic":    true,
	"clear_cache":       true,
	"update_app":        true,
	"print_test":        true,
}
