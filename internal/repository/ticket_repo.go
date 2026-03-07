package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ctx context.Context, t *domain.Ticket) error {
	query := `INSERT INTO tickets (id, tenant_id, route_id, kiosk_id, payment_id, qr_code, status, price_cents, currency, passenger_id, valid_from, valid_until, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())`
	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.TenantID, t.RouteID, t.KioskID, t.PaymentID,
		t.QRCode, t.Status, t.PriceCents, t.Currency,
		t.PassengerID, t.ValidFrom, t.ValidUntil,
	)
	return err
}

func (r *TicketRepository) GetByID(ctx context.Context, id string) (*domain.Ticket, error) {
	t := &domain.Ticket{}
	query := `SELECT id, tenant_id, route_id, kiosk_id, payment_id, qr_code, status, price_cents, currency, passenger_id, valid_from, valid_until, created_at
		FROM tickets WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.TenantID, &t.RouteID, &t.KioskID, &t.PaymentID,
		&t.QRCode, &t.Status, &t.PriceCents, &t.Currency,
		&t.PassengerID, &t.ValidFrom, &t.ValidUntil, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) GetByQRCode(ctx context.Context, qrCode string) (*domain.Ticket, error) {
	t := &domain.Ticket{}
	query := `SELECT id, tenant_id, route_id, kiosk_id, payment_id, qr_code, status, price_cents, currency, passenger_id, valid_from, valid_until, created_at
		FROM tickets WHERE qr_code = $1`
	err := r.db.QueryRowContext(ctx, query, qrCode).Scan(
		&t.ID, &t.TenantID, &t.RouteID, &t.KioskID, &t.PaymentID,
		&t.QRCode, &t.Status, &t.PriceCents, &t.Currency,
		&t.PassengerID, &t.ValidFrom, &t.ValidUntil, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, id string, status domain.TicketStatus) error {
	query := `UPDATE tickets SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *TicketRepository) ListByKiosk(ctx context.Context, kioskID string, limit int) ([]domain.Ticket, error) {
	query := `SELECT id, tenant_id, route_id, kiosk_id, payment_id, qr_code, status, price_cents, currency, passenger_id, valid_from, valid_until, created_at
		FROM tickets WHERE kiosk_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, kioskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID, &t.TenantID, &t.RouteID, &t.KioskID, &t.PaymentID,
			&t.QRCode, &t.Status, &t.PriceCents, &t.Currency,
			&t.PassengerID, &t.ValidFrom, &t.ValidUntil, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}
