package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/google/uuid"
)

type ConcesionRepository struct {
	db *sql.DB
}

func NewConcesionRepository(db *sql.DB) *ConcesionRepository {
	return &ConcesionRepository{db: db}
}

func (r *ConcesionRepository) Create(ctx context.Context, c *domain.Concesion) error {
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO concesiones (id, tenant_id, name, code, rfc, type, status, manager_id, phone, email, address, max_vehicles, max_drivers, logo_url, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		c.ID, c.TenantID, c.Name, c.Code, c.RFC, c.Type, c.Status,
		nullStr(c.ManagerID), c.Phone, c.Email, c.Address,
		c.MaxVehicles, c.MaxDrivers, c.LogoURL, c.Notes,
		c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *ConcesionRepository) GetByID(ctx context.Context, id, tenantID string) (*domain.Concesion, error) {
	c := &domain.Concesion{}
	var managerID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, code, rfc, type, status, manager_id, phone, email, address,
		       max_vehicles, max_drivers, logo_url, notes, created_at, updated_at
		FROM concesiones WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.Code, &c.RFC, &c.Type, &c.Status,
		&managerID, &c.Phone, &c.Email, &c.Address,
		&c.MaxVehicles, &c.MaxDrivers, &c.LogoURL, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.ManagerID = managerID.String
	return c, nil
}

func (r *ConcesionRepository) Update(ctx context.Context, c *domain.Concesion) error {
	c.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE concesiones SET name=$1, phone=$2, email=$3, address=$4, status=$5,
		       max_vehicles=$6, max_drivers=$7, logo_url=$8, notes=$9, updated_at=$10
		WHERE id=$11 AND tenant_id=$12`,
		c.Name, c.Phone, c.Email, c.Address, c.Status,
		c.MaxVehicles, c.MaxDrivers, c.LogoURL, c.Notes, c.UpdatedAt,
		c.ID, c.TenantID,
	)
	return err
}

func (r *ConcesionRepository) Delete(ctx context.Context, id, tenantID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM concesiones WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *ConcesionRepository) List(ctx context.Context, f domain.ListConcesionesFilter) ([]domain.Concesion, int, error) {
	where := []string{"tenant_id = $1"}
	args := []interface{}{f.TenantID}
	n := 2

	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", n))
		args = append(args, f.Status)
		n++
	}
	if f.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", n))
		args = append(args, f.Type)
		n++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", n, n))
		args = append(args, "%"+f.Search+"%")
		n++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM concesiones WHERE "+whereClause, args...).Scan(&total)

	// Paginated query
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, code, rfc, type, status, manager_id, phone, email, address,
		       max_vehicles, max_drivers, logo_url, notes, created_at, updated_at
		FROM concesiones WHERE %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		whereClause, n, n+1)
	args = append(args, limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []domain.Concesion
	for rows.Next() {
		var c domain.Concesion
		var managerID sql.NullString
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.Code, &c.RFC, &c.Type, &c.Status,
			&managerID, &c.Phone, &c.Email, &c.Address,
			&c.MaxVehicles, &c.MaxDrivers, &c.LogoURL, &c.Notes,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		c.ManagerID = managerID.String
		result = append(result, c)
	}

	return result, total, rows.Err()
}

// CountDrivers returns the number of drivers assigned to a concesion.
func (r *ConcesionRepository) CountDrivers(ctx context.Context, concesionID, tenantID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM drivers WHERE concesion_id = $1 AND tenant_id = $2`,
		concesionID, tenantID,
	).Scan(&count)
	return count, err
}

// CountVehicles returns the number of vehicles assigned to a concesion.
func (r *ConcesionRepository) CountVehicles(ctx context.Context, concesionID, tenantID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM vehicles WHERE concesion_id = $1 AND tenant_id = $2`,
		concesionID, tenantID,
	).Scan(&count)
	return count, err
}

// CountStaff returns the number of users assigned to a concesion.
func (r *ConcesionRepository) CountStaff(ctx context.Context, concesionID, tenantID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE concesion_id = $1 AND tenant_id = $2`,
		concesionID, tenantID,
	).Scan(&count)
	return count, err
}

// ListStaff returns all users assigned to a concesion.
func (r *ConcesionRepository) ListStaff(ctx context.Context, concesionID, tenantID string) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, email, phone, name, role, sub_role, concesion_id, lang, active, created_at
		FROM users WHERE concesion_id = $1 AND tenant_id = $2 ORDER BY role, name`,
		concesionID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var concID, subRole, phone sql.NullString
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &phone, &u.Name, &u.Role, &subRole, &concID, &u.Lang, &u.Active, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.Phone = phone.String
		u.SubRole = subRole.String
		u.ConcesionID = concID.String
		users = append(users, u)
	}
	return users, rows.Err()
}

// AssignStaff sets a user's concesion_id.
func (r *ConcesionRepository) AssignStaff(ctx context.Context, userID, concesionID, tenantID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET concesion_id = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`,
		concesionID, userID, tenantID,
	)
	return err
}

// RemoveStaff clears a user's concesion_id.
func (r *ConcesionRepository) RemoveStaff(ctx context.Context, userID, tenantID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET concesion_id = NULL, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		userID, tenantID,
	)
	return err
}

// SetManager updates the concesion's manager.
func (r *ConcesionRepository) SetManager(ctx context.Context, concesionID, managerID, tenantID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE concesiones SET manager_id = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`,
		managerID, concesionID, tenantID,
	)
	return err
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
