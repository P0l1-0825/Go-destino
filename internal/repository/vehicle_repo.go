package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type VehicleRepository struct {
	db *sql.DB
}

func NewVehicleRepository(db *sql.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) Create(ctx context.Context, v *domain.Vehicle) error {
	query := `INSERT INTO vehicles (id, tenant_id, driver_id, company_id, plate, brand, model, year, color, type, capacity, status, insurance_id, insurance_exp, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW(),NOW())`
	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.TenantID, v.DriverID, v.CompanyID, v.Plate, v.Brand, v.Model, v.Year, v.Color,
		v.Type, v.Capacity, v.Status, v.InsuranceID, v.InsuranceExp,
	)
	return err
}

func (r *VehicleRepository) GetByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	v := &domain.Vehicle{}
	query := `SELECT id, tenant_id, driver_id, company_id, plate, brand, model, year, color, type, capacity, status, insurance_id, insurance_exp, created_at, updated_at
		FROM vehicles WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.TenantID, &v.DriverID, &v.CompanyID, &v.Plate, &v.Brand, &v.Model, &v.Year, &v.Color,
		&v.Type, &v.Capacity, &v.Status, &v.InsuranceID, &v.InsuranceExp, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *VehicleRepository) GetByDriverID(ctx context.Context, driverID string) (*domain.Vehicle, error) {
	v := &domain.Vehicle{}
	query := `SELECT id, tenant_id, driver_id, company_id, plate, brand, model, year, color, type, capacity, status, insurance_id, insurance_exp, created_at, updated_at
		FROM vehicles WHERE driver_id = $1 AND status = 'active'`
	err := r.db.QueryRowContext(ctx, query, driverID).Scan(
		&v.ID, &v.TenantID, &v.DriverID, &v.CompanyID, &v.Plate, &v.Brand, &v.Model, &v.Year, &v.Color,
		&v.Type, &v.Capacity, &v.Status, &v.InsuranceID, &v.InsuranceExp, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *VehicleRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE vehicles SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

func (r *VehicleRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Vehicle, error) {
	query := `SELECT id, tenant_id, driver_id, company_id, plate, brand, model, year, color, type, capacity, status, insurance_id, insurance_exp, created_at, updated_at
		FROM vehicles WHERE tenant_id=$1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []domain.Vehicle
	for rows.Next() {
		var v domain.Vehicle
		if err := rows.Scan(
			&v.ID, &v.TenantID, &v.DriverID, &v.CompanyID, &v.Plate, &v.Brand, &v.Model, &v.Year, &v.Color,
			&v.Type, &v.Capacity, &v.Status, &v.InsuranceID, &v.InsuranceExp, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, rows.Err()
}
