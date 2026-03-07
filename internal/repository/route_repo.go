package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type RouteRepository struct {
	db *sql.DB
}

func NewRouteRepository(db *sql.DB) *RouteRepository {
	return &RouteRepository{db: db}
}

func (r *RouteRepository) Create(ctx context.Context, route *domain.Route) error {
	query := `INSERT INTO routes (id, tenant_id, name, code, transport_type, origin, destination, price_cents, currency, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query,
		route.ID, route.TenantID, route.Name, route.Code, route.TransportType,
		route.Origin, route.Destination, route.PriceCents, route.Currency, route.Active,
	)
	return err
}

func (r *RouteRepository) GetByID(ctx context.Context, id string) (*domain.Route, error) {
	route := &domain.Route{}
	query := `SELECT id, tenant_id, name, code, transport_type, origin, destination, price_cents, currency, active, created_at, updated_at
		FROM routes WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&route.ID, &route.TenantID, &route.Name, &route.Code, &route.TransportType,
		&route.Origin, &route.Destination, &route.PriceCents, &route.Currency, &route.Active,
		&route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (r *RouteRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Route, error) {
	query := `SELECT id, tenant_id, name, code, transport_type, origin, destination, price_cents, currency, active, created_at, updated_at
		FROM routes WHERE tenant_id = $1 AND active = true ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []domain.Route
	for rows.Next() {
		var rt domain.Route
		if err := rows.Scan(
			&rt.ID, &rt.TenantID, &rt.Name, &rt.Code, &rt.TransportType,
			&rt.Origin, &rt.Destination, &rt.PriceCents, &rt.Currency, &rt.Active,
			&rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		routes = append(routes, rt)
	}
	return routes, rows.Err()
}

func (r *RouteRepository) ListByTransportType(ctx context.Context, tenantID string, transportType domain.TransportType) ([]domain.Route, error) {
	query := `SELECT id, tenant_id, name, code, transport_type, origin, destination, price_cents, currency, active, created_at, updated_at
		FROM routes WHERE tenant_id = $1 AND transport_type = $2 AND active = true ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, tenantID, transportType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []domain.Route
	for rows.Next() {
		var rt domain.Route
		if err := rows.Scan(
			&rt.ID, &rt.TenantID, &rt.Name, &rt.Code, &rt.TransportType,
			&rt.Origin, &rt.Destination, &rt.PriceCents, &rt.Currency, &rt.Active,
			&rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		routes = append(routes, rt)
	}
	return routes, rows.Err()
}
