package repository

import (
	"context"
	"database/sql"

	"github.com/P0l1-0825/Go-destino/internal/domain"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
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
	query := `SELECT id, booking_number, tenant_id, user_id, kiosk_id, route_id, driver_id, vehicle_id,
		status, service_type, pickup_address, dropoff_address, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng,
		passenger_count, price_cents, currency, payment_id, flight_number, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM bookings WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.BookingNumber, &b.TenantID, &b.UserID, &b.KioskID, &b.RouteID,
		&b.DriverID, &b.VehicleID, &b.Status, &b.ServiceType,
		&b.PickupAddress, &b.DropoffAddress, &b.PickupLat, &b.PickupLng,
		&b.DropoffLat, &b.DropoffLng, &b.PassengerCount,
		&b.PriceCents, &b.Currency, &b.PaymentID, &b.FlightNumber,
		&b.ScheduledAt, &b.StartedAt, &b.CompletedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookingRepository) GetByNumber(ctx context.Context, number string) (*domain.Booking, error) {
	b := &domain.Booking{}
	query := `SELECT id, booking_number, tenant_id, user_id, kiosk_id, route_id, driver_id, vehicle_id,
		status, service_type, pickup_address, dropoff_address, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng,
		passenger_count, price_cents, currency, payment_id, flight_number, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM bookings WHERE booking_number = $1`
	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&b.ID, &b.BookingNumber, &b.TenantID, &b.UserID, &b.KioskID, &b.RouteID,
		&b.DriverID, &b.VehicleID, &b.Status, &b.ServiceType,
		&b.PickupAddress, &b.DropoffAddress, &b.PickupLat, &b.PickupLng,
		&b.DropoffLat, &b.DropoffLng, &b.PassengerCount,
		&b.PriceCents, &b.Currency, &b.PaymentID, &b.FlightNumber,
		&b.ScheduledAt, &b.StartedAt, &b.CompletedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	query := `UPDATE bookings SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *BookingRepository) ListByTenant(ctx context.Context, tenantID string, limit int) ([]domain.Booking, error) {
	query := `SELECT id, booking_number, tenant_id, user_id, kiosk_id, route_id, driver_id, vehicle_id,
		status, service_type, pickup_address, dropoff_address, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng,
		passenger_count, price_cents, currency, payment_id, flight_number, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM bookings WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID, &b.BookingNumber, &b.TenantID, &b.UserID, &b.KioskID, &b.RouteID,
			&b.DriverID, &b.VehicleID, &b.Status, &b.ServiceType,
			&b.PickupAddress, &b.DropoffAddress, &b.PickupLat, &b.PickupLng,
			&b.DropoffLat, &b.DropoffLng, &b.PassengerCount,
			&b.PriceCents, &b.Currency, &b.PaymentID, &b.FlightNumber,
			&b.ScheduledAt, &b.StartedAt, &b.CompletedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}
