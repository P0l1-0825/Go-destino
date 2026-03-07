package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type DriverRepository struct {
	db *sql.DB
}

func NewDriverRepository(db *sql.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (r *DriverRepository) Create(ctx context.Context, d *domain.Driver) error {
	query := `INSERT INTO drivers (id, tenant_id, user_id, company_id, license_number, status, sub_role,
		rating, total_trips, docs_verified, biometric_verified, current_lat, current_lng, heading, speed, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW(),NOW())`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.TenantID, d.UserID, d.CompanyID, d.LicenseNumber, d.Status, d.SubRole,
		d.Rating, d.TotalTrips, d.DocsVerified, d.BiometricVerified,
		d.CurrentLat, d.CurrentLng, d.Heading, d.Speed,
	)
	return err
}

func (r *DriverRepository) GetByID(ctx context.Context, id string) (*domain.Driver, error) {
	d := &domain.Driver{}
	query := `SELECT id, tenant_id, user_id, company_id, license_number, status, sub_role,
		rating, total_trips, docs_verified, biometric_verified,
		current_lat, current_lng, heading, speed, last_location_at, created_at, updated_at
		FROM drivers WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.TenantID, &d.UserID, &d.CompanyID, &d.LicenseNumber, &d.Status, &d.SubRole,
		&d.Rating, &d.TotalTrips, &d.DocsVerified, &d.BiometricVerified,
		&d.CurrentLat, &d.CurrentLng, &d.Heading, &d.Speed, &d.LastLocationAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DriverRepository) GetByUserID(ctx context.Context, userID string) (*domain.Driver, error) {
	d := &domain.Driver{}
	query := `SELECT id, tenant_id, user_id, company_id, license_number, status, sub_role,
		rating, total_trips, docs_verified, biometric_verified,
		current_lat, current_lng, heading, speed, last_location_at, created_at, updated_at
		FROM drivers WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&d.ID, &d.TenantID, &d.UserID, &d.CompanyID, &d.LicenseNumber, &d.Status, &d.SubRole,
		&d.Rating, &d.TotalTrips, &d.DocsVerified, &d.BiometricVerified,
		&d.CurrentLat, &d.CurrentLng, &d.Heading, &d.Speed, &d.LastLocationAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DriverRepository) UpdateLocation(ctx context.Context, id string, lat, lng, heading, speed float64) error {
	query := `UPDATE drivers SET current_lat=$1, current_lng=$2, heading=$3, speed=$4, last_location_at=NOW(), updated_at=NOW() WHERE id=$5`
	_, err := r.db.ExecContext(ctx, query, lat, lng, heading, speed, id)
	return err
}

func (r *DriverRepository) UpdateStatus(ctx context.Context, id string, status domain.DriverStatus) error {
	query := `UPDATE drivers SET status=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *DriverRepository) UpdateRating(ctx context.Context, id string, rating float64, totalTrips int) error {
	query := `UPDATE drivers SET rating=$1, total_trips=$2, updated_at=NOW() WHERE id=$3`
	_, err := r.db.ExecContext(ctx, query, rating, totalTrips, id)
	return err
}

func (r *DriverRepository) SetDocsVerified(ctx context.Context, id string, verified bool) error {
	query := `UPDATE drivers SET docs_verified=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, verified, id)
	return err
}

func (r *DriverRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Driver, error) {
	query := `SELECT id, tenant_id, user_id, company_id, license_number, status, sub_role,
		rating, total_trips, docs_verified, biometric_verified,
		current_lat, current_lng, heading, speed, last_location_at, created_at, updated_at
		FROM drivers WHERE tenant_id=$1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []domain.Driver
	for rows.Next() {
		var d domain.Driver
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.UserID, &d.CompanyID, &d.LicenseNumber, &d.Status, &d.SubRole,
			&d.Rating, &d.TotalTrips, &d.DocsVerified, &d.BiometricVerified,
			&d.CurrentLat, &d.CurrentLng, &d.Heading, &d.Speed, &d.LastLocationAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		drivers = append(drivers, d)
	}
	return drivers, rows.Err()
}

func (r *DriverRepository) ListByCompany(ctx context.Context, companyID string) ([]domain.Driver, error) {
	query := `SELECT id, tenant_id, user_id, company_id, license_number, status, sub_role,
		rating, total_trips, docs_verified, biometric_verified,
		current_lat, current_lng, heading, speed, last_location_at, created_at, updated_at
		FROM drivers WHERE company_id=$1 ORDER BY rating DESC`
	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []domain.Driver
	for rows.Next() {
		var d domain.Driver
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.UserID, &d.CompanyID, &d.LicenseNumber, &d.Status, &d.SubRole,
			&d.Rating, &d.TotalTrips, &d.DocsVerified, &d.BiometricVerified,
			&d.CurrentLat, &d.CurrentLng, &d.Heading, &d.Speed, &d.LastLocationAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		drivers = append(drivers, d)
	}
	return drivers, rows.Err()
}
