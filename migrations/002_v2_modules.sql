-- GoDestino SaaS Transport Kiosk — V2: Fleet, AI, Analytics, Notifications, Vouchers, Shifts, Safety
-- PostgreSQL 17

-- Airports
CREATE TABLE IF NOT EXISTS airports (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    code        VARCHAR(10) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    city        VARCHAR(255) DEFAULT '',
    country     VARCHAR(100) DEFAULT '',
    timezone    VARCHAR(100) DEFAULT 'America/Mexico_City',
    lat         DOUBLE PRECISION DEFAULT 0,
    lng         DOUBLE PRECISION DEFAULT 0,
    active      BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_airports_tenant ON airports(tenant_id);
CREATE INDEX idx_airports_code ON airports(code);

-- Airport terminals
CREATE TABLE IF NOT EXISTS airport_terminals (
    id          UUID PRIMARY KEY,
    airport_id  UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(20) DEFAULT '',
    active      BOOLEAN DEFAULT TRUE
);

-- Drivers
CREATE TABLE IF NOT EXISTS drivers (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    company_id      VARCHAR(255) DEFAULT '',
    license_number  VARCHAR(100) DEFAULT '',
    license_expiry  DATE,
    status          VARCHAR(50) DEFAULT 'offline',
    rating          DOUBLE PRECISION DEFAULT 5.0,
    total_trips     INT DEFAULT 0,
    total_ratings   INT DEFAULT 0,
    docs_verified   BOOLEAN DEFAULT FALSE,
    active          BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_drivers_tenant ON drivers(tenant_id);
CREATE INDEX idx_drivers_user ON drivers(user_id);
CREATE INDEX idx_drivers_status ON drivers(status);
CREATE INDEX idx_drivers_company ON drivers(company_id);

-- Vehicles
CREATE TABLE IF NOT EXISTS vehicles (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    driver_id       UUID NOT NULL REFERENCES drivers(id),
    vehicle_type    VARCHAR(50) NOT NULL,
    make            VARCHAR(100) DEFAULT '',
    model           VARCHAR(100) DEFAULT '',
    year            INT DEFAULT 0,
    plate           VARCHAR(20) NOT NULL,
    color           VARCHAR(50) DEFAULT '',
    capacity        INT DEFAULT 4,
    insurance_expiry DATE,
    active          BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vehicles_tenant ON vehicles(tenant_id);
CREATE INDEX idx_vehicles_driver ON vehicles(driver_id);
CREATE INDEX idx_vehicles_plate ON vehicles(plate);

-- Driver documents
CREATE TABLE IF NOT EXISTS driver_documents (
    id          UUID PRIMARY KEY,
    driver_id   UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    doc_type    VARCHAR(50) NOT NULL,
    url         TEXT NOT NULL,
    verified    BOOLEAN DEFAULT FALSE,
    expires_at  DATE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_driver_docs_driver ON driver_documents(driver_id);

-- Driver locations (latest position)
CREATE TABLE IF NOT EXISTS driver_locations (
    driver_id   UUID PRIMARY KEY REFERENCES drivers(id),
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    heading     DOUBLE PRECISION DEFAULT 0,
    speed       DOUBLE PRECISION DEFAULT 0,
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    user_id     VARCHAR(255) NOT NULL,
    channel     VARCHAR(50) NOT NULL,
    title       VARCHAR(500) NOT NULL,
    body        TEXT NOT NULL,
    data        JSONB DEFAULT '{}',
    status      VARCHAR(50) DEFAULT 'pending',
    sent_at     TIMESTAMPTZ,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_tenant ON notifications(tenant_id);
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);

-- Notification preferences
CREATE TABLE IF NOT EXISTS notification_preferences (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    channel     VARCHAR(50) NOT NULL,
    enabled     BOOLEAN DEFAULT TRUE,
    UNIQUE(user_id, channel)
);

-- Vouchers
CREATE TABLE IF NOT EXISTS vouchers (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    code            VARCHAR(50) UNIQUE NOT NULL,
    booking_id      VARCHAR(255) DEFAULT '',
    amount_cents    BIGINT NOT NULL,
    currency        VARCHAR(3) DEFAULT 'MXN',
    status          VARCHAR(50) DEFAULT 'active',
    created_by      VARCHAR(255) NOT NULL,
    redeemed_by     VARCHAR(255) DEFAULT '',
    payment_id      VARCHAR(255) DEFAULT '',
    expires_at      TIMESTAMPTZ NOT NULL,
    redeemed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vouchers_tenant ON vouchers(tenant_id);
CREATE INDEX idx_vouchers_code ON vouchers(code);
CREATE INDEX idx_vouchers_status ON vouchers(status);

-- Shifts (POS seller shifts)
CREATE TABLE IF NOT EXISTS shifts (
    id                  UUID PRIMARY KEY,
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    user_id             UUID NOT NULL REFERENCES users(id),
    airport_id          VARCHAR(255) DEFAULT '',
    terminal_id         VARCHAR(255) DEFAULT '',
    kiosk_id            VARCHAR(255) DEFAULT '',
    status              VARCHAR(50) DEFAULT 'open',
    opened_at           TIMESTAMPTZ DEFAULT NOW(),
    closed_at           TIMESTAMPTZ,
    total_sales_cents   BIGINT DEFAULT 0,
    cash_collected_cents BIGINT DEFAULT 0,
    card_collected_cents BIGINT DEFAULT 0,
    tickets_sold        INT DEFAULT 0,
    bookings_created    INT DEFAULT 0,
    commission_cents    BIGINT DEFAULT 0
);

CREATE INDEX idx_shifts_tenant ON shifts(tenant_id);
CREATE INDEX idx_shifts_user ON shifts(user_id);
CREATE INDEX idx_shifts_status ON shifts(status);

-- Audit log
CREATE TABLE IF NOT EXISTS audit_log (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    user_id     VARCHAR(255) NOT NULL,
    action      VARCHAR(100) NOT NULL,
    resource    VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) DEFAULT '',
    details     TEXT DEFAULT '',
    ip_address  VARCHAR(45) DEFAULT '',
    user_agent  TEXT DEFAULT '',
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_tenant ON audit_log(tenant_id);
CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_created ON audit_log(created_at DESC);

-- Safety incidents
CREATE TABLE IF NOT EXISTS safety_incidents (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    booking_id      VARCHAR(255) DEFAULT '',
    reporter_id     VARCHAR(255) NOT NULL,
    incident_type   VARCHAR(100) NOT NULL,
    severity        VARCHAR(50) NOT NULL DEFAULT 'medium',
    description     TEXT NOT NULL,
    lat             DOUBLE PRECISION DEFAULT 0,
    lng             DOUBLE PRECISION DEFAULT 0,
    status          VARCHAR(50) DEFAULT 'reported',
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_safety_tenant ON safety_incidents(tenant_id);
CREATE INDEX idx_safety_severity ON safety_incidents(severity);

-- SOS alerts
CREATE TABLE IF NOT EXISTS sos_alerts (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    user_id     UUID NOT NULL REFERENCES users(id),
    booking_id  VARCHAR(255) DEFAULT '',
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    status      VARCHAR(50) DEFAULT 'active',
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sos_tenant ON sos_alerts(tenant_id);
CREATE INDEX idx_sos_status ON sos_alerts(status);
