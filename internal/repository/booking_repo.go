package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

const bookingColumns = `id, booking_number, tenant_id, user_id, kiosk_id, route_id, driver_id, vehicle_id,
	status, service_type, pickup_address, dropoff_address, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng,
	passenger_count, price_cents, currency, payment_id, flight_number, cancel_reason,
	scheduled_at, started_at, completed_at, cancelled_at, created_at, updated_at`

func scanBooking(row interface{ Scan(...interface{}) error }, b *domain.Booking) error {
	return row.Scan(
		&b.ID, &b.BookingNumber, &b.TenantID, &b.UserID, &b.KioskID, &b.RouteID,
		&b.DriverID, &b.VehicleID, &b.Status, &b.ServiceType,
		&b.PickupAddress, &b.DropoffAddress, &b.PickupLat, &b.PickupLng,
		&b.DropoffLat, &b.DropoffLng, &b.PassengerCount,
		&b.PriceCents, &b.Currency, &b.PaymentID, &b.FlightNumber, &b.CancelReason,
		&b.ScheduledAt, &b.StartedAt, &b.CompletedAt, &b.CancelledAt, &b.CreatedAt, &b.UpdatedAt,
	)
}

func (r *BookingRepository) Create(ctx context.Context, b *domain.Booking) error {
	query := `INSERT INTO bookings (id, booking_number, tenant_id, user_id, kiosk_id, route_id, driver_id, vehicle_id,
		status, service_type, pickup_address, dropoff_address, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng,
		passenger_count, price_cents, currency, payment_id, flight_number, scheduled_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,NOW(),NOW())`
	_, err := r.db.ExecContext(ctx, query,
		b.ID, b.BookingNumber, b.TenantID, b.UserID, b.KioskID, b.RouteID,
		b.DriverID, b.VehicleID, b.Status, b.ServiceType,
		b.PickupAddress, b.DropoffAddress, b.PickupLat, b.PickupLng,
		b.DropoffLat, b.DropoffLng, b.PassengerCount,
		b.PriceCents, b.Currency, b.PaymentID, b.FlightNumber, b.ScheduledAt,
	)
	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	b := &domain.Booking{}
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE id = $1`
	if err := scanBooking(r.db.QueryRowContext(ctx, query, id), b); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookingRepository) GetByIDTenant(ctx context.Context, id, tenantID string) (*domain.Booking, error) {
	b := &domain.Booking{}
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE id = $1 AND tenant_id = $2`
	if err := scanBooking(r.db.QueryRowContext(ctx, query, id, tenantID), b); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookingRepository) GetByNumber(ctx context.Context, number string) (*domain.Booking, error) {
	b := &domain.Booking{}
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE booking_number = $1`
	if err := scanBooking(r.db.QueryRowContext(ctx, query, number), b); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookingRepository) GetByNumberTenant(ctx context.Context, number, tenantID string) (*domain.Booking, error) {
	b := &domain.Booking{}
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE booking_number = $1 AND tenant_id = $2`
	if err := scanBooking(r.db.QueryRowContext(ctx, query, number, tenantID), b); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookingRepository) GetByUserID(ctx context.Context, tenantID, userID string, limit, offset int) ([]domain.Booking, error) {
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE tenant_id = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	return r.queryBookings(ctx, query, tenantID, userID, limit, offset)
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id, tenantID string, status domain.BookingStatus) error {
	query := `UPDATE bookings SET status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`
	res, err := r.db.ExecContext(ctx, query, status, id, tenantID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "booking")
}

func (r *BookingRepository) AssignDriver(ctx context.Context, id, tenantID, driverID, vehicleID string) error {
	query := `UPDATE bookings SET driver_id = $1, vehicle_id = $2, status = 'assigned', updated_at = NOW() WHERE id = $3 AND tenant_id = $4`
	res, err := r.db.ExecContext(ctx, query, driverID, vehicleID, id, tenantID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "booking")
}

func (r *BookingRepository) SetStarted(ctx context.Context, id, tenantID string) error {
	query := `UPDATE bookings SET status = 'started', started_at = NOW(), updated_at = NOW() WHERE id = $1 AND tenant_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "booking")
}

func (r *BookingRepository) SetCompleted(ctx context.Context, id, tenantID string) error {
	query := `UPDATE bookings SET status = 'completed', completed_at = NOW(), updated_at = NOW() WHERE id = $1 AND tenant_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "booking")
}

func (r *BookingRepository) SetCancelled(ctx context.Context, id, tenantID, reason string) error {
	query := `UPDATE bookings SET status = 'cancelled', cancel_reason = $1, cancelled_at = NOW(), updated_at = NOW() WHERE id = $2 AND tenant_id = $3`
	res, err := r.db.ExecContext(ctx, query, reason, id, tenantID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res, "booking")
}

func (r *BookingRepository) SetPayment(ctx context.Context, id, tenantID, paymentID string) error {
	query := `UPDATE bookings SET payment_id = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`
	_, err := r.db.ExecContext(ctx, query, paymentID, id, tenantID)
	return err
}

func (r *BookingRepository) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.Booking, error) {
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`
	return r.queryBookings(ctx, query, tenantID, limit)
}

func (r *BookingRepository) ListFiltered(ctx context.Context, f domain.ListBookingsFilter) ([]domain.Booking, int, error) {
	where := []string{"tenant_id = $1"}
	args := []interface{}{f.TenantID}
	idx := 2

	if f.UserID != "" {
		where = append(where, fmt.Sprintf("user_id = $%d", idx))
		args = append(args, f.UserID)
		idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.ServiceType != "" {
		where = append(where, fmt.Sprintf("service_type = $%d", idx))
		args = append(args, f.ServiceType)
		idx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQ := `SELECT COUNT(*) FROM bookings WHERE ` + whereClause
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	dataQ := fmt.Sprintf(`SELECT `+bookingColumns+` FROM bookings WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, idx, idx+1)
	args = append(args, limit, offset)

	bookings, err := r.queryBookings(ctx, dataQ, args...)
	return bookings, total, err
}

func (r *BookingRepository) CountByStatus(ctx context.Context, tenantID string) (map[domain.BookingStatus]int, error) {
	query := `SELECT status, COUNT(*) FROM bookings WHERE tenant_id = $1 GROUP BY status`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[domain.BookingStatus]int)
	for rows.Next() {
		var status domain.BookingStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}
	return result, rows.Err()
}

func (r *BookingRepository) queryBookings(ctx context.Context, query string, args ...interface{}) ([]domain.Booking, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0, 32)
	for rows.Next() {
		var b domain.Booking
		if err := scanBooking(rows, &b); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

func checkRowsAffected(res sql.Result, entity string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s not found", entity)
	}
	return nil
}
