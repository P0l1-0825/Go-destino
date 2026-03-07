package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type ShiftRepository struct {
	db *sql.DB
}

func NewShiftRepository(db *sql.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

func (r *ShiftRepository) Create(ctx context.Context, s *domain.ShiftRecord) error {
	query := `INSERT INTO shifts (id, tenant_id, seller_id, airport_id, terminal_id, kiosk_id, status, opened_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,'open',NOW(),NOW())`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.TenantID, s.SellerID, s.AirportID, s.TerminalID, s.KioskID)
	return err
}

func (r *ShiftRepository) Close(ctx context.Context, id string, totalSales, cashCollected, cardCollected, commissionCents int64, ticketsSold, bookingsCreated int) error {
	query := `UPDATE shifts SET status='closed', closed_at=NOW(), total_sales_cents=$1, cash_collected_cents=$2,
		card_collected_cents=$3, tickets_sold=$4, bookings_created=$5, commission_cents=$6 WHERE id=$7`
	_, err := r.db.ExecContext(ctx, query, totalSales, cashCollected, cardCollected, ticketsSold, bookingsCreated, commissionCents, id)
	return err
}

func (r *ShiftRepository) GetActive(ctx context.Context, sellerID string) (*domain.ShiftRecord, error) {
	s := &domain.ShiftRecord{}
	query := `SELECT id, tenant_id, seller_id, airport_id, terminal_id, kiosk_id, status, opened_at, closed_at,
		total_sales_cents, cash_collected_cents, card_collected_cents, tickets_sold, bookings_created, commission_cents
		FROM shifts WHERE seller_id=$1 AND status='open' ORDER BY opened_at DESC LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, sellerID).Scan(
		&s.ID, &s.TenantID, &s.SellerID, &s.AirportID, &s.TerminalID, &s.KioskID, &s.Status, &s.OpenedAt, &s.ClosedAt,
		&s.TotalSales, &s.CashCollected, &s.CardCollected, &s.TicketsSold, &s.BookingsCreated, &s.CommissionCents,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ShiftRepository) ListBySeller(ctx context.Context, sellerID string, limit int) ([]domain.ShiftRecord, error) {
	query := `SELECT id, tenant_id, seller_id, airport_id, terminal_id, kiosk_id, status, opened_at, closed_at,
		total_sales_cents, cash_collected_cents, card_collected_cents, tickets_sold, bookings_created, commission_cents
		FROM shifts WHERE seller_id=$1 ORDER BY opened_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, sellerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []domain.ShiftRecord
	for rows.Next() {
		var s domain.ShiftRecord
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.SellerID, &s.AirportID, &s.TerminalID, &s.KioskID, &s.Status, &s.OpenedAt, &s.ClosedAt,
			&s.TotalSales, &s.CashCollected, &s.CardCollected, &s.TicketsSold, &s.BookingsCreated, &s.CommissionCents,
		); err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}
	return shifts, rows.Err()
}
