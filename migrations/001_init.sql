-- GoDestino SaaS Transport Kiosk — Initial Schema
-- PostgreSQL 17

-- Tenants (SaaS multi-tenancy)
CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    logo        TEXT DEFAULT '',
    active      BOOLEAN DEFAULT TRUE,
    plan        VARCHAR(50) DEFAULT 'free',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Users (all roles: SUPER_ADMIN, ADMINISTRADOR, CLIENTE_CONCESION, etc.)
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    email           VARCHAR(255) NOT NULL,
    phone           VARCHAR(30) DEFAULT '',
    password_hash   TEXT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    role            VARCHAR(50) NOT NULL,
    sub_role        VARCHAR(50) DEFAULT '',
    company_id      VARCHAR(255) DEFAULT '',
    lang            VARCHAR(10) DEFAULT 'es',
    active          BOOLEAN DEFAULT TRUE,
    mfa_enabled     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    last_login      TIMESTAMPTZ,
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_email ON users(email);

-- Routes
CREATE TABLE IF NOT EXISTS routes (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(50) NOT NULL,
    transport_type  VARCHAR(50) NOT NULL,
    origin          VARCHAR(255) NOT NULL,
    destination     VARCHAR(255) NOT NULL,
    price_cents     BIGINT NOT NULL DEFAULT 0,
    currency        VARCHAR(3) DEFAULT 'MXN',
    active          BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_routes_tenant ON routes(tenant_id);
CREATE INDEX idx_routes_type ON routes(transport_type);

-- Stops
CREATE TABLE IF NOT EXISTS stops (
    id          UUID PRIMARY KEY,
    route_id    UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    sequence    INT NOT NULL
);

-- Schedules
CREATE TABLE IF NOT EXISTS schedules (
    id          UUID PRIMARY KEY,
    route_id    UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    departure   VARCHAR(5) NOT NULL,
    active      BOOLEAN DEFAULT TRUE
);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    kiosk_id        VARCHAR(255) DEFAULT '',
    method          VARCHAR(50) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',
    amount_cents    BIGINT NOT NULL,
    currency        VARCHAR(3) DEFAULT 'MXN',
    reference       TEXT DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payments_tenant ON payments(tenant_id);
CREATE INDEX idx_payments_status ON payments(status);

-- Tickets
CREATE TABLE IF NOT EXISTS tickets (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    route_id        UUID NOT NULL REFERENCES routes(id),
    kiosk_id        VARCHAR(255) DEFAULT '',
    payment_id      UUID REFERENCES payments(id),
    qr_code         VARCHAR(255) UNIQUE NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    price_cents     BIGINT NOT NULL,
    currency        VARCHAR(3) DEFAULT 'MXN',
    passenger_id    VARCHAR(255) DEFAULT '',
    valid_from      TIMESTAMPTZ NOT NULL,
    valid_until     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tickets_tenant ON tickets(tenant_id);
CREATE INDEX idx_tickets_qr ON tickets(qr_code);
CREATE INDEX idx_tickets_kiosk ON tickets(kiosk_id);
CREATE INDEX idx_tickets_status ON tickets(status);

-- Kiosks
CREATE TABLE IF NOT EXISTS kiosks (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    name            VARCHAR(255) NOT NULL,
    location        TEXT DEFAULT '',
    airport_id      VARCHAR(255) DEFAULT '',
    terminal_id     VARCHAR(255) DEFAULT '',
    status          VARCHAR(50) DEFAULT 'offline',
    last_heartbeat  TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_kiosks_tenant ON kiosks(tenant_id);
CREATE INDEX idx_kiosks_status ON kiosks(status);

-- Bookings
CREATE TABLE IF NOT EXISTS bookings (
    id                UUID PRIMARY KEY,
    booking_number    VARCHAR(20) UNIQUE NOT NULL,
    tenant_id         UUID NOT NULL REFERENCES tenants(id),
    user_id           VARCHAR(255) DEFAULT '',
    kiosk_id          VARCHAR(255) DEFAULT '',
    route_id          VARCHAR(255) DEFAULT '',
    driver_id         VARCHAR(255) DEFAULT '',
    vehicle_id        VARCHAR(255) DEFAULT '',
    status            VARCHAR(50) NOT NULL DEFAULT 'pending',
    service_type      VARCHAR(50) NOT NULL,
    pickup_address    TEXT DEFAULT '',
    dropoff_address   TEXT DEFAULT '',
    pickup_lat        DOUBLE PRECISION DEFAULT 0,
    pickup_lng        DOUBLE PRECISION DEFAULT 0,
    dropoff_lat       DOUBLE PRECISION DEFAULT 0,
    dropoff_lng       DOUBLE PRECISION DEFAULT 0,
    passenger_count   INT DEFAULT 1,
    price_cents       BIGINT DEFAULT 0,
    currency          VARCHAR(3) DEFAULT 'MXN',
    payment_id        VARCHAR(255) DEFAULT '',
    flight_number     VARCHAR(20) DEFAULT '',
    scheduled_at      TIMESTAMPTZ,
    started_at        TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_bookings_tenant ON bookings(tenant_id);
CREATE INDEX idx_bookings_number ON bookings(booking_number);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_kiosk ON bookings(kiosk_id);

-- Transport Cards (reloadable)
CREATE TABLE IF NOT EXISTS transport_cards (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    card_number     VARCHAR(50) UNIQUE NOT NULL,
    balance_cents   BIGINT DEFAULT 0,
    currency        VARCHAR(3) DEFAULT 'MXN',
    active          BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_cards_tenant ON transport_cards(tenant_id);
CREATE INDEX idx_cards_number ON transport_cards(card_number);

-- Seed default tenant
INSERT INTO tenants (id, name, slug, active, plan) VALUES
    ('00000000-0000-0000-0000-000000000001', 'GoDestino Demo', 'demo', true, 'enterprise')
ON CONFLICT (id) DO NOTHING;
