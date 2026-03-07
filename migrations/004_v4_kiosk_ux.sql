-- v4: Kiosk UX enhancements — transport cards, kiosk sessions, receipt tracking

-- Transport cards (reloadable passenger cards for kiosk payments)
CREATE TABLE IF NOT EXISTS transport_cards (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    card_number VARCHAR(25) NOT NULL,
    balance_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'MXN',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transport_cards_number ON transport_cards(tenant_id, card_number);
CREATE INDEX IF NOT EXISTS idx_transport_cards_tenant ON transport_cards(tenant_id);

-- Kiosk sessions (track user interactions for UX analytics)
CREATE TABLE IF NOT EXISTS kiosk_sessions (
    id UUID PRIMARY KEY,
    kiosk_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    lang VARCHAR(5) NOT NULL DEFAULT 'es',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    booking_id VARCHAR(255),
    ticket_ids TEXT,
    outcome VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    step_count INT NOT NULL DEFAULT 0,
    duration_ms BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_kiosk_sessions_kiosk ON kiosk_sessions(kiosk_id);
CREATE INDEX IF NOT EXISTS idx_kiosk_sessions_tenant ON kiosk_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_kiosk_sessions_outcome ON kiosk_sessions(outcome);
CREATE INDEX IF NOT EXISTS idx_kiosk_sessions_started ON kiosk_sessions(started_at DESC);

-- Add constraint to ensure balance never goes negative
ALTER TABLE transport_cards ADD CONSTRAINT chk_balance_nonnegative CHECK (balance_cents >= 0);
