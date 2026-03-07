package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type KioskSessionRepository struct {
	db *sql.DB
}

func NewKioskSessionRepository(db *sql.DB) *KioskSessionRepository {
	return &KioskSessionRepository{db: db}
}

func (r *KioskSessionRepository) Create(ctx context.Context, s *domain.KioskSession) error {
	query := `INSERT INTO kiosk_sessions (id, kiosk_id, tenant_id, lang, started_at, outcome, step_count, duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.KioskID, s.TenantID, s.Lang, s.StartedAt, s.Outcome, s.StepCount, s.DurationMs,
	)
	return err
}

func (r *KioskSessionRepository) End(ctx context.Context, id, outcome, bookingID string, steps int, endedAt *time.Time) error {
	durationMs := int64(0)
	if endedAt != nil {
		// Calculate duration from started_at
		var startedAt time.Time
		err := r.db.QueryRowContext(ctx, `SELECT started_at FROM kiosk_sessions WHERE id = $1`, id).Scan(&startedAt)
		if err == nil {
			durationMs = endedAt.Sub(startedAt).Milliseconds()
		}
	}

	query := `UPDATE kiosk_sessions SET ended_at = $1, outcome = $2, booking_id = $3, step_count = $4, duration_ms = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, endedAt, outcome, bookingID, steps, durationMs, id)
	return err
}

func (r *KioskSessionRepository) GetByID(ctx context.Context, id string) (*domain.KioskSession, error) {
	s := &domain.KioskSession{}
	query := `SELECT id, kiosk_id, tenant_id, lang, started_at, ended_at, booking_id, ticket_ids, outcome, step_count, duration_ms
		FROM kiosk_sessions WHERE id = $1`
	var endedAt sql.NullTime
	var bookingID, ticketIDs sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.KioskID, &s.TenantID, &s.Lang, &s.StartedAt, &endedAt,
		&bookingID, &ticketIDs, &s.Outcome, &s.StepCount, &s.DurationMs,
	)
	if err != nil {
		return nil, err
	}
	if endedAt.Valid {
		s.EndedAt = &endedAt.Time
	}
	s.BookingID = bookingID.String
	s.TicketIDs = ticketIDs.String
	return s, nil
}

func (r *KioskSessionRepository) ListByKiosk(ctx context.Context, kioskID string, limit int) ([]domain.KioskSession, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, kiosk_id, tenant_id, lang, started_at, ended_at, booking_id, ticket_ids, outcome, step_count, duration_ms
		FROM kiosk_sessions WHERE kiosk_id = $1 ORDER BY started_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, kioskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.KioskSession
	for rows.Next() {
		var s domain.KioskSession
		var endedAt sql.NullTime
		var bookingID, ticketIDs sql.NullString
		if err := rows.Scan(
			&s.ID, &s.KioskID, &s.TenantID, &s.Lang, &s.StartedAt, &endedAt,
			&bookingID, &ticketIDs, &s.Outcome, &s.StepCount, &s.DurationMs,
		); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			s.EndedAt = &endedAt.Time
		}
		s.BookingID = bookingID.String
		s.TicketIDs = ticketIDs.String
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// AvgSessionDuration returns the average session duration in ms for a kiosk.
func (r *KioskSessionRepository) AvgSessionDuration(ctx context.Context, kioskID string) (int64, error) {
	var avg sql.NullInt64
	query := `SELECT AVG(duration_ms) FROM kiosk_sessions WHERE kiosk_id = $1 AND outcome = 'completed' AND duration_ms > 0`
	err := r.db.QueryRowContext(ctx, query, kioskID).Scan(&avg)
	if err != nil {
		return 0, err
	}
	return avg.Int64, nil
}

// CountByOutcome returns session counts grouped by outcome for a kiosk.
func (r *KioskSessionRepository) CountByOutcome(ctx context.Context, kioskID string) (map[string]int, error) {
	query := `SELECT outcome, COUNT(*) FROM kiosk_sessions WHERE kiosk_id = $1 GROUP BY outcome`
	rows, err := r.db.QueryContext(ctx, query, kioskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var outcome string
		var count int
		if err := rows.Scan(&outcome, &count); err != nil {
			return nil, err
		}
		counts[outcome] = count
	}
	return counts, rows.Err()
}
