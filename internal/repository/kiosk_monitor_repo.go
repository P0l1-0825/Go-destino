package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type KioskMonitorRepository struct {
	db *sql.DB
}

func NewKioskMonitorRepository(db *sql.DB) *KioskMonitorRepository {
	return &KioskMonitorRepository{db: db}
}

// --- Telemetry ---

func (r *KioskMonitorRepository) SaveTelemetry(ctx context.Context, t *domain.KioskTelemetry) error {
	query := `INSERT INTO kiosk_telemetry
		(id, kiosk_id, tenant_id, cpu_percent, memory_percent, disk_percent, temperature,
		paper_level, printer_ok, scanner_ok, network_type, network_mbps, uptime_sec,
		app_version, os_version, screen_on, error_count, collected_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`
	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.KioskID, t.TenantID, t.CPUPercent, t.MemoryPercent, t.DiskPercent,
		t.Temperature, t.PaperLevel, t.PrinterOK, t.ScannerOK, t.NetworkType,
		t.NetworkMbps, t.UptimeSec, t.AppVersion, t.OSVersion, t.ScreenOn,
		t.ErrorCount, t.CollectedAt,
	)
	return err
}

func (r *KioskMonitorRepository) GetLatestTelemetry(ctx context.Context, kioskID string) (*domain.KioskTelemetry, error) {
	t := &domain.KioskTelemetry{}
	query := `SELECT id, kiosk_id, tenant_id, cpu_percent, memory_percent, disk_percent,
		temperature, paper_level, printer_ok, scanner_ok, network_type, network_mbps,
		uptime_sec, app_version, os_version, screen_on, error_count, collected_at
		FROM kiosk_telemetry WHERE kiosk_id = $1 ORDER BY collected_at DESC LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, kioskID).Scan(
		&t.ID, &t.KioskID, &t.TenantID, &t.CPUPercent, &t.MemoryPercent, &t.DiskPercent,
		&t.Temperature, &t.PaperLevel, &t.PrinterOK, &t.ScannerOK, &t.NetworkType,
		&t.NetworkMbps, &t.UptimeSec, &t.AppVersion, &t.OSVersion, &t.ScreenOn,
		&t.ErrorCount, &t.CollectedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *KioskMonitorRepository) GetTelemetryHistory(ctx context.Context, kioskID string, since time.Time, limit int) ([]domain.KioskTelemetry, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, kiosk_id, tenant_id, cpu_percent, memory_percent, disk_percent,
		temperature, paper_level, printer_ok, scanner_ok, network_type, network_mbps,
		uptime_sec, app_version, os_version, screen_on, error_count, collected_at
		FROM kiosk_telemetry WHERE kiosk_id = $1 AND collected_at >= $2
		ORDER BY collected_at DESC LIMIT $3`
	rows, err := r.db.QueryContext(ctx, query, kioskID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.KioskTelemetry
	for rows.Next() {
		var t domain.KioskTelemetry
		if err := rows.Scan(
			&t.ID, &t.KioskID, &t.TenantID, &t.CPUPercent, &t.MemoryPercent, &t.DiskPercent,
			&t.Temperature, &t.PaperLevel, &t.PrinterOK, &t.ScannerOK, &t.NetworkType,
			&t.NetworkMbps, &t.UptimeSec, &t.AppVersion, &t.OSVersion, &t.ScreenOn,
			&t.ErrorCount, &t.CollectedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// CleanupOldTelemetry removes telemetry older than the given duration.
func (r *KioskMonitorRepository) CleanupOldTelemetry(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := r.db.ExecContext(ctx, `DELETE FROM kiosk_telemetry WHERE collected_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// --- Events ---

func (r *KioskMonitorRepository) CreateEvent(ctx context.Context, e *domain.KioskEvent) error {
	query := `INSERT INTO kiosk_events (id, kiosk_id, tenant_id, event_type, severity, message, details, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.ExecContext(ctx, query,
		e.ID, e.KioskID, e.TenantID, e.EventType, e.Severity, e.Message, e.Details, e.CreatedAt,
	)
	return err
}

func (r *KioskMonitorRepository) ListEvents(ctx context.Context, kioskID string, limit int) ([]domain.KioskEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, kiosk_id, tenant_id, event_type, severity, message, details, created_at
		FROM kiosk_events WHERE kiosk_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, kioskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.KioskEvent
	for rows.Next() {
		var e domain.KioskEvent
		var details sql.NullString
		if err := rows.Scan(&e.ID, &e.KioskID, &e.TenantID, &e.EventType, &e.Severity, &e.Message, &details, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Details = details.String
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *KioskMonitorRepository) ListEventsByTenant(ctx context.Context, tenantID string, severity string, limit int) ([]domain.KioskEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows *sql.Rows
	var err error
	if severity != "" {
		query := `SELECT id, kiosk_id, tenant_id, event_type, severity, message, details, created_at
			FROM kiosk_events WHERE tenant_id = $1 AND severity = $2 ORDER BY created_at DESC LIMIT $3`
		rows, err = r.db.QueryContext(ctx, query, tenantID, severity, limit)
	} else {
		query := `SELECT id, kiosk_id, tenant_id, event_type, severity, message, details, created_at
			FROM kiosk_events WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`
		rows, err = r.db.QueryContext(ctx, query, tenantID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.KioskEvent
	for rows.Next() {
		var e domain.KioskEvent
		var details sql.NullString
		if err := rows.Scan(&e.ID, &e.KioskID, &e.TenantID, &e.EventType, &e.Severity, &e.Message, &details, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Details = details.String
		events = append(events, e)
	}
	return events, rows.Err()
}

// --- Alerts ---

func (r *KioskMonitorRepository) CreateAlert(ctx context.Context, a *domain.KioskAlert) error {
	query := `INSERT INTO kiosk_alerts (id, kiosk_id, tenant_id, alert_type, severity, message, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.KioskID, a.TenantID, a.AlertType, a.Severity, a.Message, a.Active, a.CreatedAt,
	)
	return err
}

func (r *KioskMonitorRepository) GetActiveAlerts(ctx context.Context, kioskID string) ([]domain.KioskAlert, error) {
	query := `SELECT id, kiosk_id, tenant_id, alert_type, severity, message, active, acked_by, acked_at, resolved_at, created_at
		FROM kiosk_alerts WHERE kiosk_id = $1 AND active = true ORDER BY created_at DESC`
	return r.scanAlerts(ctx, query, kioskID)
}

func (r *KioskMonitorRepository) GetActiveAlertsByTenant(ctx context.Context, tenantID string) ([]domain.KioskAlert, error) {
	query := `SELECT id, kiosk_id, tenant_id, alert_type, severity, message, active, acked_by, acked_at, resolved_at, created_at
		FROM kiosk_alerts WHERE tenant_id = $1 AND active = true ORDER BY severity, created_at DESC`
	return r.scanAlerts(ctx, query, tenantID)
}

func (r *KioskMonitorRepository) AckAlert(ctx context.Context, alertID, userID string) error {
	now := time.Now()
	query := `UPDATE kiosk_alerts SET acked_by = $1, acked_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, userID, now, alertID)
	return err
}

func (r *KioskMonitorRepository) ResolveAlert(ctx context.Context, alertID string) error {
	now := time.Now()
	query := `UPDATE kiosk_alerts SET active = false, resolved_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, now, alertID)
	return err
}

func (r *KioskMonitorRepository) ResolveAlertsByType(ctx context.Context, kioskID, alertType string) error {
	now := time.Now()
	query := `UPDATE kiosk_alerts SET active = false, resolved_at = $1 WHERE kiosk_id = $2 AND alert_type = $3 AND active = true`
	_, err := r.db.ExecContext(ctx, query, now, kioskID, alertType)
	return err
}

func (r *KioskMonitorRepository) HasActiveAlert(ctx context.Context, kioskID, alertType string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM kiosk_alerts WHERE kiosk_id = $1 AND alert_type = $2 AND active = true`
	err := r.db.QueryRowContext(ctx, query, kioskID, alertType).Scan(&count)
	return count > 0, err
}

func (r *KioskMonitorRepository) CountActiveAlerts(ctx context.Context, tenantID string) (critical int, warning int, err error) {
	query := `SELECT severity, COUNT(*) FROM kiosk_alerts WHERE tenant_id = $1 AND active = true GROUP BY severity`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var sev string
		var cnt int
		if err := rows.Scan(&sev, &cnt); err != nil {
			return 0, 0, err
		}
		switch sev {
		case "critical":
			critical = cnt
		case "warning":
			warning = cnt
		}
	}
	return critical, warning, rows.Err()
}

func (r *KioskMonitorRepository) scanAlerts(ctx context.Context, query string, arg string) ([]domain.KioskAlert, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []domain.KioskAlert
	for rows.Next() {
		var a domain.KioskAlert
		var ackedBy sql.NullString
		var ackedAt, resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.KioskID, &a.TenantID, &a.AlertType, &a.Severity,
			&a.Message, &a.Active, &ackedBy, &ackedAt, &resolvedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.AckedBy = ackedBy.String
		if ackedAt.Valid {
			a.AckedAt = &ackedAt.Time
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// --- Remote Commands ---

func (r *KioskMonitorRepository) CreateCommand(ctx context.Context, c *domain.KioskRemoteCommand) error {
	query := `INSERT INTO kiosk_remote_commands (id, kiosk_id, tenant_id, command, params, status, issued_by, issued_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.KioskID, c.TenantID, c.Command, c.Params, c.Status, c.IssuedBy, c.IssuedAt,
	)
	return err
}

func (r *KioskMonitorRepository) GetPendingCommands(ctx context.Context, kioskID string) ([]domain.KioskRemoteCommand, error) {
	query := `SELECT id, kiosk_id, tenant_id, command, params, status, issued_by, result, issued_at, executed_at
		FROM kiosk_remote_commands WHERE kiosk_id = $1 AND status IN ('pending','sent') ORDER BY issued_at`
	return r.scanCommands(ctx, query, kioskID)
}

func (r *KioskMonitorRepository) ListCommands(ctx context.Context, kioskID string, limit int) ([]domain.KioskRemoteCommand, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, kiosk_id, tenant_id, command, params, status, issued_by, result, issued_at, executed_at
		FROM kiosk_remote_commands WHERE kiosk_id = $1 ORDER BY issued_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, kioskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanCommandRows(rows)
}

func (r *KioskMonitorRepository) UpdateCommandStatus(ctx context.Context, cmdID, status, result string) error {
	now := time.Now()
	query := `UPDATE kiosk_remote_commands SET status = $1, result = $2, executed_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, status, result, now, cmdID)
	return err
}

func (r *KioskMonitorRepository) scanCommands(ctx context.Context, query string, arg string) ([]domain.KioskRemoteCommand, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanCommandRows(rows)
}

func (r *KioskMonitorRepository) scanCommandRows(rows *sql.Rows) ([]domain.KioskRemoteCommand, error) {
	var cmds []domain.KioskRemoteCommand
	for rows.Next() {
		var c domain.KioskRemoteCommand
		var params, result sql.NullString
		var executedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.KioskID, &c.TenantID, &c.Command, &params, &c.Status,
			&c.IssuedBy, &result, &c.IssuedAt, &executedAt); err != nil {
			return nil, err
		}
		c.Params = params.String
		c.Result = result.String
		if executedAt.Valid {
			c.ExecutedAt = &executedAt.Time
		}
		cmds = append(cmds, c)
	}
	return cmds, rows.Err()
}

// --- Uptime Calculation ---

// CalcUptimePercent calculates the percentage of time a kiosk was online in the last N hours.
func (r *KioskMonitorRepository) CalcUptimePercent(ctx context.Context, kioskID string, hours int) (float64, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	var total, online int
	query := `SELECT COUNT(*),
		COUNT(*) FILTER (WHERE cpu_percent IS NOT NULL)
		FROM kiosk_telemetry WHERE kiosk_id = $1 AND collected_at >= $2`
	err := r.db.QueryRowContext(ctx, query, kioskID, since).Scan(&total, &online)
	if err != nil || total == 0 {
		return 0, err
	}
	return float64(online) / float64(total) * 100, nil
}

// SessionStatsForKiosk returns aggregated session stats.
func (r *KioskMonitorRepository) SessionStatsForKiosk(ctx context.Context, kioskID string) (*domain.KioskSessionStats, error) {
	stats := &domain.KioskSessionStats{}

	query := `SELECT
		COUNT(*),
		COUNT(*) FILTER (WHERE outcome = 'completed'),
		COUNT(*) FILTER (WHERE outcome = 'abandoned'),
		COUNT(*) FILTER (WHERE outcome = 'timeout'),
		COALESCE(AVG(duration_ms) FILTER (WHERE outcome = 'completed' AND duration_ms > 0), 0)
		FROM kiosk_sessions WHERE kiosk_id = $1`

	err := r.db.QueryRowContext(ctx, query, kioskID).Scan(
		&stats.TotalSessions, &stats.CompletedCount, &stats.AbandonedCount,
		&stats.TimeoutCount, &stats.AvgDurationMs,
	)
	if err != nil {
		return nil, err
	}

	if stats.TotalSessions > 0 {
		stats.CompletionRate = float64(stats.CompletedCount) / float64(stats.TotalSessions) * 100
	}

	return stats, nil
}

// CountSessionsToday counts sessions started today for a kiosk.
func (r *KioskMonitorRepository) CountSessionsToday(ctx context.Context, kioskID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM kiosk_sessions WHERE kiosk_id = $1 AND started_at >= CURRENT_DATE`
	err := r.db.QueryRowContext(ctx, query, kioskID).Scan(&count)
	return count, err
}
