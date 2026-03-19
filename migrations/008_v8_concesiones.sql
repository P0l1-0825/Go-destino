-- Migration 008: Concesiones (Franchises)
-- Formalizes the concession model: each concession owns vehicles and has its own staff hierarchy.

-- ─── 1. Create concesiones table ───

CREATE TABLE IF NOT EXISTS concesiones (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50) NOT NULL,
    rfc         VARCHAR(20),                          -- Mexico tax ID
    type        VARCHAR(20) NOT NULL DEFAULT 'mixed',  -- taxi, van, shuttle, mixed
    status      VARCHAR(20) NOT NULL DEFAULT 'pending', -- active, inactive, suspended, pending
    manager_id  UUID,                                  -- FK to users (administrativo principal)
    phone       VARCHAR(20),
    email       VARCHAR(255),
    address     TEXT,
    max_vehicles INT NOT NULL DEFAULT 50,
    max_drivers  INT NOT NULL DEFAULT 50,
    logo_url    TEXT,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_concesiones_tenant ON concesiones(tenant_id);
CREATE INDEX IF NOT EXISTS idx_concesiones_status ON concesiones(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_concesiones_manager ON concesiones(manager_id);

-- ─── 2. Add concesion_id to users ───

ALTER TABLE users ADD COLUMN IF NOT EXISTS concesion_id UUID REFERENCES concesiones(id);
CREATE INDEX IF NOT EXISTS idx_users_concesion ON users(concesion_id);

-- ─── 3. Add concesion_id to drivers ───

ALTER TABLE drivers ADD COLUMN IF NOT EXISTS concesion_id UUID REFERENCES concesiones(id);
CREATE INDEX IF NOT EXISTS idx_drivers_concesion ON drivers(concesion_id);

-- ─── 4. Add concesion_id to vehicles ───

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS concesion_id UUID REFERENCES concesiones(id);
CREATE INDEX IF NOT EXISTS idx_vehicles_concesion ON vehicles(concesion_id);

-- ─── 5. Concesion ↔ Airport relationship (many-to-many) ───

CREATE TABLE IF NOT EXISTS concesion_airports (
    concesion_id UUID NOT NULL REFERENCES concesiones(id) ON DELETE CASCADE,
    airport_id   UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    PRIMARY KEY (concesion_id, airport_id)
);

-- ─── 6. Migrate existing company_id data to concesion_id ───
-- This is safe: if company_id is empty/null, concesion_id stays null.
-- After verifying data, company_id columns can be dropped in a future migration.

-- Note: Run these manually after creating concesion records for each unique company_id:
-- INSERT INTO concesiones (tenant_id, name, code, type, status)
--   SELECT DISTINCT tenant_id, company_id, company_id, 'mixed', 'active'
--   FROM users WHERE company_id IS NOT NULL AND company_id != '';
--
-- UPDATE users u SET concesion_id = c.id
--   FROM concesiones c WHERE c.code = u.company_id AND c.tenant_id = u.tenant_id;
-- UPDATE drivers d SET concesion_id = c.id
--   FROM concesiones c WHERE c.code = d.company_id AND c.tenant_id = d.tenant_id;
-- UPDATE vehicles v SET concesion_id = c.id
--   FROM concesiones c WHERE c.code = v.company_id AND c.tenant_id = v.tenant_id;

-- ─── 7. Add manager FK constraint (deferred — manager must exist first) ───

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_concesiones_manager'
    ) THEN
        ALTER TABLE concesiones
            ADD CONSTRAINT fk_concesiones_manager
            FOREIGN KEY (manager_id) REFERENCES users(id)
            ON DELETE SET NULL;
    END IF;
END $$;
