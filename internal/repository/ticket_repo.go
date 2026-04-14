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

const ticketColumns = `id, tenant_id, route_id, kiosk_id, payment_id, qr_code, status, price_cents, currency, passenger_id, valid_from, valid_until, created_at`

func scanTicket(row interface{ Scan(...interface{}) error }, t *domain.Ticket) error {
	return row.Scan(
		&t.ID, &t.TenantID, &t.RouteID, &t.KioskID, &t.PaymentID,
		&t.QRCode, &t.Status, &t.PriceCents, &t.Currency,
		&t.PassengerID, &t.ValidFrom, &t.ValidUntil, &t.CreatedAt,
	)
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
	query := `SELECT ` + ticketColumns + ` FROM tickets WHERE id = $1`
	if err := scanTicket(r.db.QueryRowContext(ctx, query, id), t); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) GetByQRCode(ctx context.Context, qrCode string) (*domain.Ticket, error) {
	t := &domain.Ticket{}
	query := `SELECT ` + ticketColumns + ` FROM tickets WHERE qr_code = $1`
	if err := scanTicket(r.db.QueryRowContext(ctx, query, qrCode), t); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, id string, status domain.TicketStatus) error {
	query := `UPDATE tickets SET status = $1 WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "ticket")
}

func (r *TicketRepository) ListByKiosk(ctx context.Context, kioskID string, limit int) ([]domain.Ticket, error) {
	query := `SELECT ` + ticketColumns + ` FROM tickets WHERE kiosk_id = $1 ORDER BY created_at DESC LIMIT $2`
	return r.queryTickets(ctx, query, kioskID, limit)
}

func (r *TicketRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]domain.Ticket, error) {
	query := `SELECT ` + ticketColumns + ` FROM tickets WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return r.queryTickets(ctx, query, tenantID, limit, offset)
}

func (r *TicketRepository) ListByRoute(ctx context.Context, routeID string, limit int) ([]domain.Ticket, error) {
	query := `SELECT ` + ticketColumns + ` FROM tickets WHERE route_id = $1 ORDER BY created_at DESC LIMIT $2`
	return r.queryTickets(ctx, query, routeID, limit)
}

func (r *TicketRepository) ListByPassenger(ctx context.Context, passengerID string, limit int) ([]domain.Ticket, error) {
	query := `SELECT ` + ticketColumns + ` FROM tickets WHERE passenger_id = $1 ORDER BY created_at DESC LIMIT $2`
	return r.queryTickets(ctx, query, passengerID, limit)
}

func (r *TicketRepository) CountByRoute(ctx context.Context, routeID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tickets WHERE route_id = $1 AND status = 'active'`
	err := r.db.QueryRowContext(ctx, query, routeID).Scan(&count)
	return count, err
}

func (r *TicketRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tickets WHERE tenant_id = $1`
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	return count, err
}

func (r *TicketRepository) ExpireOld(ctx context.Context) (int64, error) {
	query := `UPDATE tickets SET status = 'expired' WHERE status = 'active' AND valid_until < NOW()`
	res, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *TicketRepository) queryTickets(ctx context.Context, query string, args ...interface{}) ([]domain.Ticket, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate with a reasonable initial capacity to avoid repeated
	// backing-array copies on the first ~8 rows (the most common page size).
	tickets := make([]domain.Ticket, 0, 16)
	for rows.Next() {
		var t domain.Ticket
		if err := scanTicket(rows, &t); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}
