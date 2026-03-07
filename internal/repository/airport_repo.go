package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type AirportRepository struct {
	db *sql.DB
}

func NewAirportRepository(db *sql.DB) *AirportRepository {
	return &AirportRepository{db: db}
}

func (r *AirportRepository) Create(ctx context.Context, a *domain.Airport) error {
	query := `INSERT INTO airports (id, tenant_id, code, name, city, country, country_code, lat, lng, timezone, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())`
	_, err := r.db.ExecContext(ctx, query, a.ID, a.TenantID, a.Code, a.Name, a.City, a.Country, a.CountryCode, a.Lat, a.Lng, a.Timezone, a.Active)
	return err
}

func (r *AirportRepository) GetByID(ctx context.Context, id string) (*domain.Airport, error) {
	a := &domain.Airport{}
	query := `SELECT id, tenant_id, code, name, city, country, country_code, lat, lng, timezone, active FROM airports WHERE id=$1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&a.ID, &a.TenantID, &a.Code, &a.Name, &a.City, &a.Country, &a.CountryCode, &a.Lat, &a.Lng, &a.Timezone, &a.Active)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AirportRepository) GetByCode(ctx context.Context, code string) (*domain.Airport, error) {
	a := &domain.Airport{}
	query := `SELECT id, tenant_id, code, name, city, country, country_code, lat, lng, timezone, active FROM airports WHERE code=$1`
	err := r.db.QueryRowContext(ctx, query, code).Scan(&a.ID, &a.TenantID, &a.Code, &a.Name, &a.City, &a.Country, &a.CountryCode, &a.Lat, &a.Lng, &a.Timezone, &a.Active)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AirportRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Airport, error) {
	query := `SELECT id, tenant_id, code, name, city, country, country_code, lat, lng, timezone, active FROM airports WHERE tenant_id=$1 AND active=true ORDER BY code`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var airports []domain.Airport
	for rows.Next() {
		var a domain.Airport
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Code, &a.Name, &a.City, &a.Country, &a.CountryCode, &a.Lat, &a.Lng, &a.Timezone, &a.Active); err != nil {
			return nil, err
		}
		airports = append(airports, a)
	}
	return airports, rows.Err()
}
