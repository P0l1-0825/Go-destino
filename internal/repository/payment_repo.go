package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `INSERT INTO payments (id, tenant_id, kiosk_id, method, status, amount_cents, currency, reference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.TenantID, p.KioskID, p.Method, p.Status,
		p.AmountCents, p.Currency, p.Reference,
	)
	return err
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	p := &domain.Payment{}
	query := `SELECT id, tenant_id, kiosk_id, method, status, amount_cents, currency, reference, created_at, updated_at
		FROM payments WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.TenantID, &p.KioskID, &p.Method, &p.Status,
		&p.AmountCents, &p.Currency, &p.Reference, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
