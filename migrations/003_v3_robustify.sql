-- GoDestino SaaS Transport Kiosk — V3: Robustification
-- Adds missing columns to bookings, payments, and shifts tables.

-- Bookings: add cancel_reason, cancelled_at
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancel_reason TEXT DEFAULT '';
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

-- Bookings: additional indexes
CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_driver ON bookings(driver_id);
CREATE INDEX IF NOT EXISTS idx_bookings_service_type ON bookings(service_type);

-- Payments: add booking_id, user_id, failure_reason, refunded_at
ALTER TABLE payments ADD COLUMN IF NOT EXISTS booking_id VARCHAR(255) DEFAULT '';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS user_id VARCHAR(255) DEFAULT '';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS failure_reason TEXT DEFAULT '';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMPTZ;

-- Payments: additional indexes
CREATE INDEX IF NOT EXISTS idx_payments_booking ON payments(booking_id);
CREATE INDEX IF NOT EXISTS idx_payments_kiosk ON payments(kiosk_id);
CREATE INDEX IF NOT EXISTS idx_payments_method ON payments(method);

-- Shifts: rename user_id to seller_id for clarity (the repo uses seller_id)
ALTER TABLE shifts RENAME COLUMN user_id TO seller_id;
ALTER TABLE shifts ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

-- Shifts: fix index
DROP INDEX IF EXISTS idx_shifts_user;
CREATE INDEX IF NOT EXISTS idx_shifts_seller ON shifts(seller_id);

-- Tickets: additional indexes
CREATE INDEX IF NOT EXISTS idx_tickets_route ON tickets(route_id);
CREATE INDEX IF NOT EXISTS idx_tickets_passenger ON tickets(passenger_id);
CREATE INDEX IF NOT EXISTS idx_tickets_valid_until ON tickets(valid_until);
