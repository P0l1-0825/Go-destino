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

const shiftColumns = `id, tenant_id, seller_id, airport_id, terminal_id, kiosk_id, status, opened_at, closed_at,
	total_sales_cents, cash_collected_cents, card_collected_cents, tickets_sold, bookings_created, commission_cents`

func scanShift(row interface{ Scan(...interface{}) error }, s *domain.ShiftRecord) error {
	return row.Scan(
		&s.ID, &s.TenantID, &s.SellerID, &s.AirportID, &s.TerminalID, &s.KioskID,
		&s.Status, &s.OpenedAt, &s.ClosedAt,
		&s.TotalSales, &s.CashCollected, &s.CardCollected,
		&s.TicketsSold, &s.BookingsCreated, &s.CommissionCents,
	)
}

func (r *ShiftRepository) Create(ctx context.Context, s *domain.ShiftRecord) error {
	query := `INSERT INTO shifts (id, tenant_id, seller_id, airport_id, terminal_id, kiosk_id, status, opened_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,'open',NOW(),NOW())`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.TenantID, s.SellerID, s.AirportID, s.TerminalID, s.KioskID)
	return err
}

// Close updates the shift with totals. Fixed parameter ordering to match SQL placeholders.
func (r *ShiftRepository) Close(ctx context.Context, id string, totalSales, cashCollected, cardCollected, commissionCents int64, ticketsSold, bookingsCreated int) error {
	query := `UPDATE shifts SET status='closed', closed_at=NOW(),
		total_sales_cents=$1, cash_collected_cents=$2, card_collected_cents=$3,
		tickets_sold=$4, bookings_created=$5, commission_cents=$6
		WHERE id=$7 AND status='open'`
	res, err := r.db.ExecContext(ctx, query, totalSales, cashCollected, cardCollected, ticketsSold, bookingsCreated, commissionCents, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "shift")
}

func (r *ShiftRepository) GetByID(ctx context.Context, id string) (*domain.ShiftRecord, error) {
	s := &domain.ShiftRecord{}
	query := `SELECT ` + shiftColumns + ` FROM shifts WHERE id=$1`
	if err := scanShift(r.db.QueryRowContext(ctx, query, id), s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ShiftRepository) GetActive(ctx context.Context, sellerID string) (*domain.ShiftRecord, error) {
	s := &domain.ShiftRecord{}
	query := `SELECT ` + shiftColumns + ` FROM shifts WHERE seller_id=$1 AND status='open' ORDER BY opened_at DESC LIMIT 1`
	if err := scanShift(r.db.QueryRowContext(ctx, query, sellerID), s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ShiftRepository) ListBySeller(ctx context.Context, sellerID string, limit int) ([]domain.ShiftRecord, error) {
	query := `SELECT ` + shiftColumns + ` FROM shifts WHERE seller_id=$1 ORDER BY opened_at DESC LIMIT $2`
	return r.queryShifts(ctx, query, sellerID, limit)
}

func (r *ShiftRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]domain.ShiftRecord, error) {
	query := `SELECT ` + shiftColumns + ` FROM shifts WHERE tenant_id=$1 ORDER BY opened_at DESC LIMIT $2 OFFSET $3`
	return r.queryShifts(ctx, query, tenantID, limit, offset)
}

func (r *ShiftRepository) ListByKiosk(ctx context.Context, kioskID string, limit int) ([]domain.ShiftRecord, error) {
	query := `SELECT ` + shiftColumns + ` FROM shifts WHERE kiosk_id=$1 ORDER BY opened_at DESC LIMIT $2`
	return r.queryShifts(ctx, query, kioskID, limit)
}

func (r *ShiftRepository) queryShifts(ctx context.Context, query string, args ...interface{}) ([]domain.ShiftRecord, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shifts := make([]domain.ShiftRecord, 0, 16)
	for rows.Next() {
		var s domain.ShiftRecord
		if err := scanShift(rows, &s); err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}
	return shifts, rows.Err()
}
