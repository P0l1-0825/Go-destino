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
	query := `INSERT INTO users (id, tenant_id, email, password_hash, name, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.TenantID, u.Email, u.PasswordHash, u.Name, u.Role, u.Active)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, tenant_id, email, password_hash, name, role, active, created_at, updated_at
		FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, tenant_id, email, password_hash, name, role, active, created_at, updated_at
		FROM users WHERE tenant_id = $1 AND email = $2`
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.User, error) {
	query := `SELECT id, tenant_id, email, password_hash, name, role, active, created_at, updated_at
		FROM users WHERE tenant_id = $1 ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
