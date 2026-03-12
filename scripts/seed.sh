#!/usr/bin/env bash
# seed.sh — Seeds the GoDestino database with initial data for development/production
# Usage: ./scripts/seed.sh [API_BASE_URL]
# Default: http://localhost:8080

set -euo pipefail

API="${1:-http://localhost:8080}/api/v1"
TENANT_SLUG="godestino-cancun"

echo "=== GoDestino Seed Script ==="
echo "API: $API"
echo ""

# ─── Step 1: Create tenant directly in PostgreSQL ───────────────────────
echo "1/7 Creating tenant..."
# Insert only if not exists, then always select
docker exec go-destino-postgres-1 psql -U destino -d destino -tAc "
  INSERT INTO tenants (id, name, slug, active, plan, created_at, updated_at)
  VALUES (gen_random_uuid(), 'GoDestino Cancún', '${TENANT_SLUG}', true, 'pro', NOW(), NOW())
  ON CONFLICT (slug) DO NOTHING;
" > /dev/null 2>&1 || true
TENANT_ID=$(docker exec go-destino-postgres-1 psql -U destino -d destino -tAc "SELECT id FROM tenants WHERE slug = '${TENANT_SLUG}';")
TENANT_ID=$(echo "$TENANT_ID" | tr -d '[:space:]')
echo "   ✅ Tenant ID: $TENANT_ID"

# ─── Step 2: Register SUPER_ADMIN user ──────────────────────────────────
echo "2/7 Registering admin user..."
ADMIN_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API/auth/register" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "admin@godestino.com",
    "password": "Admin2024x",
    "name": "Admin GoDestino",
    "role": "SUPER_ADMIN",
    "lang": "es"
  }')

ADMIN_HTTP=$(echo "$ADMIN_RESP" | tail -1)
ADMIN_BODY=$(echo "$ADMIN_RESP" | sed '$d')

if [ "$ADMIN_HTTP" = "201" ]; then
  ADMIN_ID=$(echo "$ADMIN_BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "unknown")
  echo "   ✅ Admin created (ID: $ADMIN_ID)"
elif [ "$ADMIN_HTTP" = "409" ]; then
  echo "   ⚠️  Admin already exists (conflict), continuing..."
else
  echo "   ❌ Failed to register admin (HTTP $ADMIN_HTTP)"
  echo "   $ADMIN_BODY"
fi

# ─── Step 3: Login as admin ─────────────────────────────────────────────
echo "3/7 Logging in as admin..."
LOGIN_RESP=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "admin@godestino.com",
    "password": "Admin2024x"
  }')

TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
if [ -z "$TOKEN" ]; then
  echo "   ❌ Login failed"
  echo "   $LOGIN_RESP"
  exit 1
fi
echo "   ✅ Logged in (token: ${TOKEN:0:20}...)"

# Helper: authenticated POST
auth_post() {
  curl -s -X POST "$API$1" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -d "$2"
}

auth_get() {
  curl -s -X GET "$API$1" \
    -H "Authorization: Bearer $TOKEN" \
    -H "X-Tenant-ID: $TENANT_ID"
}

# ─── Step 4: Create Airport ─────────────────────────────────────────────
echo "4/7 Creating airport (CUN)..."
AIRPORT_RESP=$(auth_post "/admin/airports" '{
  "code": "CUN",
  "name": "Aeropuerto Internacional de Cancún",
  "city": "Cancún",
  "country": "México",
  "country_code": "MX",
  "lat": 21.0365,
  "lng": -86.8769,
  "timezone": "America/Cancun"
}')
AIRPORT_ID=$(echo "$AIRPORT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
if [ -n "$AIRPORT_ID" ]; then
  echo "   ✅ Airport CUN created (ID: $AIRPORT_ID)"
else
  echo "   ⚠️  Airport response: $AIRPORT_RESP"
  # Try to get existing airport
  AIRPORTS=$(auth_get "/admin/airports")
  AIRPORT_ID=$(echo "$AIRPORTS" | python3 -c "import sys,json; airports=json.load(sys.stdin)['data']; print(airports[0]['id'] if airports else '')" 2>/dev/null || echo "")
  if [ -n "$AIRPORT_ID" ]; then
    echo "   ✅ Using existing airport (ID: $AIRPORT_ID)"
  fi
fi

# ─── Step 5: Create Kiosk service user (VENDEDOR) ───────────────────────
echo "5/7 Creating kiosk service user..."
KIOSK_USER_RESP=$(auth_post "/admin/users" '{
  "email": "kiosk@godestino.com",
  "password": "Kiosk2024x",
  "name": "Kiosk Service Account",
  "role": "VENDEDOR",
  "lang": "es"
}')
KIOSK_USER_HTTP=$(echo "$KIOSK_USER_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null || echo "")
if [ -n "$KIOSK_USER_HTTP" ]; then
  echo "   ✅ Kiosk user created (ID: $KIOSK_USER_HTTP)"
else
  echo "   ⚠️  Kiosk user response: $KIOSK_USER_RESP"
  echo "   (May already exist, continuing...)"
fi

# ─── Step 6: Create Transport Routes ────────────────────────────────────
echo "6/7 Creating transport routes..."

create_route() {
  local name="$1" code="$2" type="$3" origin="$4" dest="$5" price="$6"
  local resp=$(auth_post "/routes" "{
    \"name\": \"$name\",
    \"code\": \"$code\",
    \"transport_type\": \"$type\",
    \"origin\": \"$origin\",
    \"destination\": \"$dest\",
    \"price_cents\": $price,
    \"currency\": \"MXN\"
  }")
  local rid=$(echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
  if [ -n "$rid" ]; then
    echo "   ✅ Route '$name' (ID: $rid)"
  else
    echo "   ⚠️  Route '$name': $resp"
  fi
}

create_route "Aeropuerto → Zona Hotelera" "CUN-ZH" "shuttle" "Aeropuerto CUN T3" "Zona Hotelera Cancún" 35000
create_route "Aeropuerto → Centro Cancún" "CUN-CC" "shuttle" "Aeropuerto CUN T3" "Centro de Cancún" 25000
create_route "Aeropuerto → Playa del Carmen" "CUN-PDC" "shuttle" "Aeropuerto CUN T3" "Playa del Carmen" 75000
create_route "Aeropuerto → Tulum" "CUN-TUL" "shuttle" "Aeropuerto CUN T3" "Tulum" 120000

# ─── Step 7: Register Kiosk Device ──────────────────────────────────────
echo "7/7 Registering kiosk device..."
KIOSK_RESP=$(auth_post "/kiosks/register" "{
  \"name\": \"Kiosk Terminal 3\",
  \"location\": \"Terminal 3, Puerta de Llegadas\",
  \"airport_id\": \"$AIRPORT_ID\",
  \"terminal_id\": \"T3\"
}")
KIOSK_ID=$(echo "$KIOSK_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
if [ -n "$KIOSK_ID" ]; then
  echo "   ✅ Kiosk registered (ID: $KIOSK_ID)"
else
  echo "   ⚠️  Kiosk response: $KIOSK_RESP"
fi

# ─── Summary ────────────────────────────────────────────────────────────
echo ""
echo "=== Seed Complete ==="
echo "Tenant ID:    $TENANT_ID"
echo "Airport ID:   ${AIRPORT_ID:-unknown}"
echo "Kiosk ID:     ${KIOSK_ID:-unknown}"
echo ""
echo "Admin login:  admin@godestino.com / Admin2024x"
echo "Kiosk login:  kiosk@godestino.com / Kiosk2024x"
echo ""
echo "Next steps:"
echo "  1. Set NEXT_PUBLIC_TENANT_ID=$TENANT_ID in .env.local"
echo "  2. Set NEXT_PUBLIC_KIOSK_ID=${KIOSK_ID:-<kiosk_id>} in .env.local"
echo "  3. Set NEXT_PUBLIC_DEMO_MODE=false"
echo "  4. Set NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1"
