-- Migration 007: Production bootstrap
-- Ensures default tenant is properly configured for production.
-- Safe to run on both fresh and existing databases.

-- If the default tenant still has slug 'demo', update it to production values.
-- If slug 'godestino-cancun' already exists (from seed script), leave it alone.
UPDATE tenants SET
    name = 'GoDestino Cancún',
    slug = 'godestino-cancun',
    plan = 'pro'
WHERE id = '00000000-0000-0000-0000-000000000001'
  AND slug = 'demo'
  AND NOT EXISTS (SELECT 1 FROM tenants WHERE slug = 'godestino-cancun');
