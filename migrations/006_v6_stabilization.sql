-- GoDestino SaaS Transport Kiosk — V6: Stabilization
-- Adds missing columns required by repositories and fixes schema gaps.

-- ============================================================
-- Drivers: add columns used by driver_repo.go
-- ============================================================
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS sub_role VARCHAR(50) DEFAULT '';
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS biometric_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS current_lat DOUBLE PRECISION DEFAULT 0;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS current_lng DOUBLE PRECISION DEFAULT 0;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS heading DOUBLE PRECISION DEFAULT 0;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS speed DOUBLE PRECISION DEFAULT 0;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS last_location_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_drivers_biometric ON drivers(biometric_verified);
CREATE INDEX IF NOT EXISTS idx_drivers_location_at ON drivers(last_location_at);

-- ============================================================
-- Notifications: add booking_id used by notification_repo.go
-- ============================================================
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS booking_id VARCHAR(255) DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_notifications_booking ON notifications(booking_id);

-- ============================================================
-- Airports: add country_code used by airport_repo.go
-- ============================================================
ALTER TABLE airports ADD COLUMN IF NOT EXISTS country_code VARCHAR(10) DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_airports_country_code ON airports(country_code);

-- ============================================================
-- Transport cards: fix CHECK constraint (idempotent re-apply)
-- ============================================================
ALTER TABLE transport_cards DROP CONSTRAINT IF EXISTS chk_balance_nonnegative;
ALTER TABLE transport_cards ADD CONSTRAINT chk_balance_nonnegative CHECK (balance_cents >= 0);
