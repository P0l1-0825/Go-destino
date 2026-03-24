package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type VehicleRepository struct {
	db *sql.DB
}

func NewVehicleRepository(db *sql.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

// vehicleColumns are the DB columns we SELECT/INSERT.
// DB schema: id, tenant_id, driver_id, vehicle_type, make, model, year, plate, color, capacity, insurance_expiry, active, concesion_id, created_at, updated_at
const vehicleSelectCols = `id, tenant_id, driver_id, vehicle_type, make, model, year, plate, color, capacity, insurance_expiry, active, concesion_id, created_at, updated_at`

func scanVehicle(row interface{ Scan(dest ...any) error }) (*domain.Vehicle, error) {
	v := &domain.Vehicle{}
	var active bool
	var insuranceExp *time.Time
	var concesionID sql.NullString
	err := row.Scan(
		&v.ID, &v.TenantID, &v.DriverID, &v.Type, &v.Brand, &v.Model, &v.Year, &v.Plate, &v.Color,
		&v.Capacity, &insuranceExp, &active, &concesionID, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if active {
		v.Status = "active"
	} else {
		v.Status = "inactive"
	}
	v.InsuranceExp = insuranceExp
	if concesionID.Valid {
		v.ConcesionID = concesionID.String
	}
	return v, nil
}

func (r *VehicleRepository) Create(ctx context.Context, v *domain.Vehicle) error {
	query := `INSERT INTO vehicles (id, tenant_id, driver_id, vehicle_type, make, model, year, plate, color, capacity, insurance_expiry, active, concesion_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NULLIF($13,'')::uuid,NOW(),NOW())`
	active := v.Status != "inactive"
	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.TenantID, v.DriverID, v.Type, v.Brand, v.Model, v.Year, v.Plate, v.Color,
		v.Capacity, v.InsuranceExp, active, v.ConcesionID,
	)
	return err
}

func (r *VehicleRepository) GetByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	query := `SELECT ` + vehicleSelectCols + ` FROM vehicles WHERE id = $1`
	return scanVehicle(r.db.QueryRowContext(ctx, query, id))
}

func (r *VehicleRepository) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Vehicle, error) {
	query := `SELECT ` + vehicleSelectCols + ` FROM vehicles WHERE id = $1 AND tenant_id = $2`
	return scanVehicle(r.db.QueryRowContext(ctx, query, id, tenantID))
}

func (r *VehicleRepository) GetByDriverID(ctx context.Context, driverID string) (*domain.Vehicle, error) {
	query := `SELECT ` + vehicleSelectCols + ` FROM vehicles WHERE driver_id = $1 AND active = true`
	return scanVehicle(r.db.QueryRowContext(ctx, query, driverID))
}

func (r *VehicleRepository) GetByDriverIDTenant(ctx context.Context, driverID, tenantID string) (*domain.Vehicle, error) {
	query := `SELECT ` + vehicleSelectCols + ` FROM vehicles WHERE driver_id = $1 AND tenant_id = $2 AND active = true`
	return scanVehicle(r.db.QueryRowContext(ctx, query, driverID, tenantID))
}

func (r *VehicleRepository) UpdateStatus(ctx context.Context, id, tenantID, status string) error {
	active := status == "active"
	_, err := r.db.ExecContext(ctx, `UPDATE vehicles SET active=$1, updated_at=NOW() WHERE id=$2 AND tenant_id=$3`, active, id, tenantID)
	return err
}

func (r *VehicleRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Vehicle, error) {
	query := `SELECT ` + vehicleSelectCols + ` FROM vehicles WHERE tenant_id=$1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []domain.Vehicle
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, *v)
	}
	return vehicles, rows.Err()
}
