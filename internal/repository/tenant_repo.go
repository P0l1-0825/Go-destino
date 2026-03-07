package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type TenantRepository struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, t *domain.Tenant) error {
	query := `INSERT INTO tenants (id, name, slug, logo, active, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Slug, t.Logo, t.Active, t.Plan)
	return err
}

func (r *TenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t := &domain.Tenant{}
	query := `SELECT id, name, slug, logo, active, plan, created_at, updated_at FROM tenants WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Slug, &t.Logo, &t.Active, &t.Plan, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	t := &domain.Tenant{}
	query := `SELECT id, name, slug, logo, active, plan, created_at, updated_at FROM tenants WHERE slug = $1`
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&t.ID, &t.Name, &t.Slug, &t.Logo, &t.Active, &t.Plan, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TenantRepository) List(ctx context.Context) ([]domain.Tenant, error) {
	query := `SELECT id, name, slug, logo, active, plan, created_at, updated_at FROM tenants ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Logo, &t.Active, &t.Plan, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}
