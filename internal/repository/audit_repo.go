package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, entry *domain.AuditLogEntry) error {
	query := `INSERT INTO audit_log (id, tenant_id, user_id, action, resource, resource_id, details, ip_address, user_agent, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.TenantID, entry.UserID, entry.Action, entry.Resource,
		entry.ResourceID, entry.Details, entry.IPAddress, entry.UserAgent,
	)
	return err
}

func (r *AuditRepository) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.AuditLogEntry, error) {
	query := `SELECT id, tenant_id, user_id, action, resource, resource_id, details, ip_address, user_agent, created_at
		FROM audit_log WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.AuditLogEntry
	for rows.Next() {
		var e domain.AuditLogEntry
		if err := rows.Scan(&e.ID, &e.TenantID, &e.UserID, &e.Action, &e.Resource, &e.ResourceID, &e.Details, &e.IPAddress, &e.UserAgent, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *AuditRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLogEntry, error) {
	query := `SELECT id, tenant_id, user_id, action, resource, resource_id, details, ip_address, user_agent, created_at
		FROM audit_log WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.AuditLogEntry
	for rows.Next() {
		var e domain.AuditLogEntry
		if err := rows.Scan(&e.ID, &e.TenantID, &e.UserID, &e.Action, &e.Resource, &e.ResourceID, &e.Details, &e.IPAddress, &e.UserAgent, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
