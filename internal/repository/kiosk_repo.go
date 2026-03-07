package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type KioskRepository struct {
	db *sql.DB
}

func NewKioskRepository(db *sql.DB) *KioskRepository {
	return &KioskRepository{db: db}
}

func (r *KioskRepository) Create(ctx context.Context, k *domain.Kiosk) error {
	query := `INSERT INTO kiosks (id, tenant_id, name, location, airport_id, terminal_id, status, last_heartbeat, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query,
		k.ID, k.TenantID, k.Name, k.Location, k.AirportID, k.TerminalID, k.Status,
	)
	return err
}

func (r *KioskRepository) GetByID(ctx context.Context, id string) (*domain.Kiosk, error) {
	k := &domain.Kiosk{}
	query := `SELECT id, tenant_id, name, location, airport_id, terminal_id, status, last_heartbeat, created_at, updated_at
		FROM kiosks WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&k.ID, &k.TenantID, &k.Name, &k.Location, &k.AirportID, &k.TerminalID,
		&k.Status, &k.LastHeartbeat, &k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func (r *KioskRepository) UpdateHeartbeat(ctx context.Context, id string) error {
	query := `UPDATE kiosks SET last_heartbeat = NOW(), status = 'online', updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *KioskRepository) UpdateStatus(ctx context.Context, id string, status domain.KioskStatus) error {
	query := `UPDATE kiosks SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *KioskRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Kiosk, error) {
	query := `SELECT id, tenant_id, name, location, airport_id, terminal_id, status, last_heartbeat, created_at, updated_at
		FROM kiosks WHERE tenant_id = $1 ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kiosks []domain.Kiosk
	for rows.Next() {
		var k domain.Kiosk
		if err := rows.Scan(
			&k.ID, &k.TenantID, &k.Name, &k.Location, &k.AirportID, &k.TerminalID,
			&k.Status, &k.LastHeartbeat, &k.CreatedAt, &k.UpdatedAt,
		); err != nil {
			return nil, err
		}
		kiosks = append(kiosks, k)
	}
	return kiosks, rows.Err()
}
