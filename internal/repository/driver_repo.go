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

func (r *DriverRepository) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Driver, error) {
	d := &domain.Driver{}
	query := `SELECT id, tenant_id, user_id, company_id, license_number, status, sub_role,
		rating, total_trips, docs_verified, biometric_verified,
		current_lat, current_lng, heading, speed, last_location_at, created_at, updated_at
		FROM drivers WHERE id = $1 AND tenant_id = $2`
	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
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

func (r *DriverRepository) UpdateLocation(ctx context.Context, id, tenantID string, lat, lng, heading, speed float64) error {
	// Dual-write: update drivers table and upsert driver_locations
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`UPDATE drivers SET current_lat=$1, current_lng=$2, heading=$3, speed=$4, last_location_at=NOW(), updated_at=NOW() WHERE id=$5 AND tenant_id=$6`,
		lat, lng, heading, speed, id, tenantID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO driver_locations (driver_id, lat, lng, heading, speed, updated_at)
		VALUES ($1,$2,$3,$4,$5,NOW())
		ON CONFLICT (driver_id) DO UPDATE SET lat=$2, lng=$3, heading=$4, speed=$5, updated_at=NOW()`,
		id, lat, lng, heading, speed)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *DriverRepository) UpdateStatus(ctx context.Context, id, tenantID string, status domain.DriverStatus) error {
	query := `UPDATE drivers SET status=$1, updated_at=NOW() WHERE id=$2 AND tenant_id=$3`
	_, err := r.db.ExecContext(ctx, query, status, id, tenantID)
	return err
}

func (r *DriverRepository) UpdateRating(ctx context.Context, id, tenantID string, rating float64, totalTrips int) error {
	query := `UPDATE drivers SET rating=$1, total_trips=$2, updated_at=NOW() WHERE id=$3 AND tenant_id=$4`
	_, err := r.db.ExecContext(ctx, query, rating, totalTrips, id, tenantID)
	return err
}

func (r *DriverRepository) SetDocsVerified(ctx context.Context, id, tenantID string, verified bool) error {
	query := `UPDATE drivers SET docs_verified=$1, updated_at=NOW() WHERE id=$2 AND tenant_id=$3`
	_, err := r.db.ExecContext(ctx, query, verified, id, tenantID)
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

func (r *DriverRepository) GetActiveLocations(ctx context.Context, tenantID string) ([]domain.DriverLocation, error) {
	query := `SELECT d.id, dl.lat, dl.lng, dl.heading, dl.speed, EXTRACT(EPOCH FROM dl.updated_at)::BIGINT
		FROM driver_locations dl
		JOIN drivers d ON d.id = dl.driver_id
		WHERE d.tenant_id = $1 AND d.status IN ('available','busy','on_trip')
		ORDER BY dl.updated_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locs []domain.DriverLocation
	for rows.Next() {
		var loc domain.DriverLocation
		if err := rows.Scan(&loc.DriverID, &loc.Lat, &loc.Lng, &loc.Heading, &loc.Speed, &loc.Timestamp); err != nil {
			return nil, err
		}
		locs = append(locs, loc)
	}
	return locs, rows.Err()
}

func (r *DriverRepository) ListByCompany(ctx context.Context, companyID, tenantID string) ([]domain.Driver, error) {
	query := `SELECT id, tenant_id, user_id, company_id, license_number, status, sub_role,
		rating, total_trips, docs_verified, biometric_verified,
		current_lat, current_lng, heading, speed, last_location_at, created_at, updated_at
		FROM drivers WHERE company_id=$1 AND tenant_id=$2 ORDER BY rating DESC`
	rows, err := r.db.QueryContext(ctx, query, companyID, tenantID)
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
