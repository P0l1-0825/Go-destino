package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

const paymentColumns = `id, tenant_id, booking_id, kiosk_id, user_id, method, status, amount_cents, currency, reference, failure_reason, refunded_at, created_at, updated_at`

func scanPayment(row interface{ Scan(...interface{}) error }, p *domain.Payment) error {
	return row.Scan(
		&p.ID, &p.TenantID, &p.BookingID, &p.KioskID, &p.UserID,
		&p.Method, &p.Status, &p.AmountCents, &p.Currency,
		&p.Reference, &p.FailureReason, &p.RefundedAt,
		&p.CreatedAt, &p.UpdatedAt,
	)
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `INSERT INTO payments (id, tenant_id, booking_id, kiosk_id, user_id, method, status, amount_cents, currency, reference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.TenantID, p.BookingID, p.KioskID, p.UserID,
		p.Method, p.Status, p.AmountCents, p.Currency, p.Reference,
	)
	return err
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	p := &domain.Payment{}
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE id = $1`
	if err := scanPayment(r.db.QueryRowContext(ctx, query, id), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Payment, error) {
	p := &domain.Payment{}
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE id = $1 AND tenant_id = $2`
	if err := scanPayment(r.db.QueryRowContext(ctx, query, id, tenantID), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) GetByReference(ctx context.Context, reference string) (*domain.Payment, error) {
	p := &domain.Payment{}
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE reference = $1`
	if err := scanPayment(r.db.QueryRowContext(ctx, query, reference), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) GetByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error) {
	p := &domain.Payment{}
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE booking_id = $1 ORDER BY created_at DESC LIMIT 1`
	if err := scanPayment(r.db.QueryRowContext(ctx, query, bookingID), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) GetByBookingIDTenant(ctx context.Context, bookingID, tenantID string) (*domain.Payment, error) {
	p := &domain.Payment{}
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE booking_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT 1`
	if err := scanPayment(r.db.QueryRowContext(ctx, query, bookingID, tenantID), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, tenantID string, status domain.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`
	res, err := r.db.ExecContext(ctx, query, status, id, tenantID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "payment")
}

func (r *PaymentRepository) MarkFailed(ctx context.Context, id, tenantID, reason string) error {
	query := `UPDATE payments SET status = 'failed', failure_reason = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`
	_, err := r.db.ExecContext(ctx, query, reason, id, tenantID)
	return err
}

func (r *PaymentRepository) MarkRefunded(ctx context.Context, id, tenantID string) error {
	query := `UPDATE payments SET status = 'refunded', refunded_at = NOW(), updated_at = NOW() WHERE id = $1 AND tenant_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "payment")
}

func (r *PaymentRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]domain.Payment, error) {
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return r.queryPayments(ctx, query, tenantID, limit, offset)
}

func (r *PaymentRepository) ListByKiosk(ctx context.Context, kioskID string, limit int) ([]domain.Payment, error) {
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE kiosk_id = $1 ORDER BY created_at DESC LIMIT $2`
	return r.queryPayments(ctx, query, kioskID, limit)
}

func (r *PaymentRepository) SumByTenant(ctx context.Context, tenantID string) (int64, error) {
	var total sql.NullInt64
	query := `SELECT SUM(amount_cents) FROM payments WHERE tenant_id = $1 AND status = 'completed'`
	if err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&total); err != nil {
		return 0, err
	}
	return total.Int64, nil
}

func (r *PaymentRepository) Refund(ctx context.Context, originalID, tenantID string, refundPayment *domain.Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	updateQ := `UPDATE payments SET status = 'refunded', refunded_at = NOW(), updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND status = 'completed'`
	res, err := tx.ExecContext(ctx, updateQ, originalID, tenantID)
	if err != nil {
		return fmt.Errorf("marking refunded: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("payment %s not found or not in completed status", originalID)
	}

	insertQ := `INSERT INTO payments (id, tenant_id, booking_id, kiosk_id, user_id, method, status, amount_cents, currency, reference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'completed', $7, $8, $9, NOW(), NOW())`
	_, err = tx.ExecContext(ctx, insertQ,
		refundPayment.ID, refundPayment.TenantID, refundPayment.BookingID, refundPayment.KioskID, refundPayment.UserID,
		refundPayment.Method, -refundPayment.AmountCents, refundPayment.Currency, refundPayment.Reference,
	)
	if err != nil {
		return fmt.Errorf("inserting refund: %w", err)
	}

	return tx.Commit()
}

func (r *PaymentRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *PaymentRepository) queryPayments(ctx context.Context, query string, args ...interface{}) ([]domain.Payment, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := scanPayment(rows, &p); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
