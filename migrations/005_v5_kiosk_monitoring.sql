-- v5: Kiosk Monitoring & Remote Support

-- Telemetry snapshots (hardware metrics over time)
CREATE TABLE IF NOT EXISTS kiosk_telemetry (
    id UUID PRIMARY KEY,
    kiosk_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    cpu_percent REAL NOT NULL DEFAULT 0,
    memory_percent REAL NOT NULL DEFAULT 0,
    disk_percent REAL NOT NULL DEFAULT 0,
    temperature REAL NOT NULL DEFAULT 0,
    paper_level INT NOT NULL DEFAULT 100,
    printer_ok BOOLEAN NOT NULL DEFAULT true,
    scanner_ok BOOLEAN NOT NULL DEFAULT true,
    network_type VARCHAR(20) NOT NULL DEFAULT 'ethernet',
    network_mbps REAL NOT NULL DEFAULT 0,
    uptime_sec BIGINT NOT NULL DEFAULT 0,
    app_version VARCHAR(50) NOT NULL DEFAULT '',
    os_version VARCHAR(100) NOT NULL DEFAULT '',
    screen_on BOOLEAN NOT NULL DEFAULT true,
    error_count INT NOT NULL DEFAULT 0,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kiosk_telemetry_kiosk ON kiosk_telemetry(kiosk_id);
CREATE INDEX IF NOT EXISTS idx_kiosk_telemetry_collected ON kiosk_telemetry(kiosk_id, collected_at DESC);
CREATE INDEX IF NOT EXISTS idx_kiosk_telemetry_cleanup ON kiosk_telemetry(collected_at);

-- Events log (significant kiosk events)
CREATE TABLE IF NOT EXISTS kiosk_events (
    id UUID PRIMARY KEY,
    kiosk_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    message TEXT NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kiosk_events_kiosk ON kiosk_events(kiosk_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_kiosk_events_tenant ON kiosk_events(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_kiosk_events_severity ON kiosk_events(tenant_id, severity);

-- Active alerts
CREATE TABLE IF NOT EXISTS kiosk_alerts (
    id UUID PRIMARY KEY,
    kiosk_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    message TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    acked_by VARCHAR(255),
    acked_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kiosk_alerts_active ON kiosk_alerts(kiosk_id, active);
CREATE INDEX IF NOT EXISTS idx_kiosk_alerts_tenant ON kiosk_alerts(tenant_id, active);
CREATE INDEX IF NOT EXISTS idx_kiosk_alerts_type ON kiosk_alerts(kiosk_id, alert_type, active);

-- Remote commands
CREATE TABLE IF NOT EXISTS kiosk_remote_commands (
    id UUID PRIMARY KEY,
    kiosk_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    command VARCHAR(50) NOT NULL,
    params TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    issued_by VARCHAR(255) NOT NULL,
    result TEXT,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_kiosk_commands_pending ON kiosk_remote_commands(kiosk_id, status);
CREATE INDEX IF NOT EXISTS idx_kiosk_commands_kiosk ON kiosk_remote_commands(kiosk_id, issued_at DESC);
