package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id, tenant_id, email, phone, password_hash, name, role, sub_role, company_id, lang, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.TenantID, u.Email, u.Phone, u.PasswordHash, u.Name, u.Role,
		u.SubRole, u.CompanyID, u.Lang, u.Active,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, tenant_id, email, phone, password_hash, name, role, sub_role, company_id, lang, active, mfa_enabled, created_at, updated_at, last_login
		FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Phone, &u.PasswordHash, &u.Name, &u.Role,
		&u.SubRole, &u.CompanyID, &u.Lang, &u.Active, &u.MFAEnabled,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLogin,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, tenant_id, email, phone, password_hash, name, role, sub_role, company_id, lang, active, mfa_enabled, created_at, updated_at, last_login
		FROM users WHERE tenant_id = $1 AND email = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Phone, &u.PasswordHash, &u.Name, &u.Role,
		&u.SubRole, &u.CompanyID, &u.Lang, &u.Active, &u.MFAEnabled,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLogin,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET name = $1, phone = $2, sub_role = $3, company_id = $4, lang = $5, updated_at = NOW()
		WHERE id = $6`
	res, err := r.db.ExecContext(ctx, query, u.Name, u.Phone, u.SubRole, u.CompanyID, u.Lang, u.ID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "user")
}

func (r *UserRepository) ChangePassword(ctx context.Context, userID, newHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, newHash, userID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "user")
}

func (r *UserRepository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE users SET active = false, updated_at = NOW() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "user")
}

func (r *UserRepository) Activate(ctx context.Context, id string) error {
	query := `UPDATE users SET active = true, updated_at = NOW() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "user")
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListByTenant returns all users without password_hash exposed.
func (r *UserRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.User, error) {
	query := `SELECT id, tenant_id, email, phone, name, role, sub_role, company_id, lang, active, mfa_enabled, created_at, updated_at, last_login
		FROM users WHERE tenant_id = $1 ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.Phone, &u.Name, &u.Role,
			&u.SubRole, &u.CompanyID, &u.Lang, &u.Active, &u.MFAEnabled,
			&u.CreatedAt, &u.UpdatedAt, &u.LastLogin,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) ListByTenantPaginated(ctx context.Context, tenantID string, limit, offset int) ([]domain.User, int, error) {
	var total int
	countQ := `SELECT COUNT(*) FROM users WHERE tenant_id = $1`
	if err := r.db.QueryRowContext(ctx, countQ, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, tenant_id, email, phone, name, role, sub_role, company_id, lang, active, mfa_enabled, created_at, updated_at, last_login
		FROM users WHERE tenant_id = $1 ORDER BY name LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.Phone, &u.Name, &u.Role,
			&u.SubRole, &u.CompanyID, &u.Lang, &u.Active, &u.MFAEnabled,
			&u.CreatedAt, &u.UpdatedAt, &u.LastLogin,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, tenantID, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE tenant_id = $1 AND email = $2)`
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) CountByRole(ctx context.Context, tenantID string) (map[domain.UserRole]int, error) {
	query := `SELECT role, COUNT(*) FROM users WHERE tenant_id = $1 GROUP BY role`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[domain.UserRole]int)
	for rows.Next() {
		var role domain.UserRole
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		result[role] = count
	}
	return result, rows.Err()
}
