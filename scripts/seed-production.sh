#!/usr/bin/env bash
# seed-production.sh — Seeds Railway/production database via API only (no Docker)
# Usage: ./scripts/seed-production.sh <RAILWAY_API_URL>
# Example: ./scripts/seed-production.sh https://go-destino-production.up.railway.app

set -euo pipefail

if [ -z "${1:-}" ]; then
  echo "Usage: $0 <API_BASE_URL>"
  echo "Example: $0 https://go-destino-production.up.railway.app"
  exit 1
fi

API="${1}/api/v1"
# Default tenant created by migration 001_init.sql
TENANT_ID="00000000-0000-0000-0000-000000000001"

echo "=== GoDestino Production Seed Script ==="
echo "API: $API"
echo "Tenant: $TENANT_ID"
echo ""

# ─── Step 1: Health check ──────────────────────────────────────────────
echo "0/7 Health check..."
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$1/health" 2>/dev/null || echo "000")
if [ "$HEALTH" != "200" ]; then
  echo "   ❌ Backend not reachable (HTTP $HEALTH)"
  echo "   Make sure the backend is running at: $1"
  exit 1
fi
echo "   ✅ Backend healthy"

# ─── Step 2: Register admin (using known default tenant) ──────────────
echo "1/7 Registering admin user..."
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
  echo "   ✅ Admin created"
elif [ "$ADMIN_HTTP" = "409" ]; then
  echo "   ⚠️  Admin already exists, continuing..."
else
  echo "   ❌ Registration failed (HTTP $ADMIN_HTTP): $ADMIN_BODY"
  echo "   Trying to continue anyway..."
fi

# ─── Step 3: Login as admin ────────────────────────────────────────────
echo "2/7 Logging in as admin..."
LOGIN_RESP=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "admin@godestino.com",
    "password": "Admin2024x"
  }')

TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
if [ -z "$TOKEN" ]; then
  echo "   ❌ Login failed: $LOGIN_RESP"
  exit 1
fi
echo "   ✅ Logged in (token: ${TOKEN:0:20}...)"

# Helper functions
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

# ─── Step 4: Create Airport ───────────────────────────────────────────
echo "3/7 Creating airport (CUN)..."
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
  echo "   ⚠️  Airport: $AIRPORT_RESP"
  AIRPORTS=$(auth_get "/admin/airports")
  AIRPORT_ID=$(echo "$AIRPORTS" | python3 -c "import sys,json; airports=json.load(sys.stdin)['data']; print(airports[0]['id'] if airports else '')" 2>/dev/null || echo "")
  [ -n "$AIRPORT_ID" ] && echo "   ✅ Using existing airport (ID: $AIRPORT_ID)"
fi

# ─── Step 5: Create Kiosk service user ─────────────────────────────────
echo "4/7 Creating kiosk service user..."
KIOSK_USER_RESP=$(auth_post "/admin/users" '{
  "email": "kiosk@godestino.com",
  "password": "Kiosk2024x",
  "name": "Kiosk Service Account",
  "role": "VENDEDOR",
  "lang": "es"
}')
echo "   Response: $(echo "$KIOSK_USER_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print('✅ Created' if d.get('success') else '⚠️  ' + d.get('error','unknown'))" 2>/dev/null || echo "⚠️  $KIOSK_USER_RESP")"

# ─── Step 6: Create Routes ────────────────────────────────────────────
echo "5/7 Creating transport routes..."
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
  local success=$(echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin).get('success',False))" 2>/dev/null || echo "")
  if [ "$success" = "True" ]; then
    echo "   ✅ Route '$name'"
  else
    echo "   ⚠️  Route '$name': $(echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',''))" 2>/dev/null || echo "$resp")"
  fi
}

create_route "Aeropuerto → Zona Hotelera" "CUN-ZH" "shuttle" "Aeropuerto CUN T3" "Zona Hotelera Cancún" 35000
create_route "Aeropuerto → Centro Cancún" "CUN-CC" "shuttle" "Aeropuerto CUN T3" "Centro de Cancún" 25000
create_route "Aeropuerto → Playa del Carmen" "CUN-PDC" "shuttle" "Aeropuerto CUN T3" "Playa del Carmen" 75000
create_route "Aeropuerto → Tulum" "CUN-TUL" "shuttle" "Aeropuerto CUN T3" "Tulum" 120000

# ─── Step 7: Register Kiosk Device ────────────────────────────────────
echo "6/7 Registering kiosk device..."
KIOSK_RESP=$(auth_post "/kiosks/register" "{
  \"name\": \"Kiosk Terminal 3\",
  \"location\": \"Terminal 3, Puerta de Llegadas\",
  \"airport_id\": \"${AIRPORT_ID:-00000000-0000-0000-0000-000000000000}\",
  \"terminal_id\": \"T3\"
}")
KIOSK_ID=$(echo "$KIOSK_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null || echo "")
if [ -n "$KIOSK_ID" ]; then
  echo "   ✅ Kiosk registered (ID: $KIOSK_ID)"
else
  echo "   ⚠️  Kiosk: $KIOSK_RESP"
fi

# ─── Step 7: Verify kiosk login ───────────────────────────────────────
echo "7/7 Verifying kiosk login..."
KIOSK_LOGIN=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"email": "kiosk@godestino.com", "password": "Kiosk2024x"}')
KIOSK_TOKEN=$(echo "$KIOSK_LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
if [ -n "$KIOSK_TOKEN" ]; then
  echo "   ✅ Kiosk login successful"
else
  echo "   ❌ Kiosk login failed: $KIOSK_LOGIN"
fi

# ─── Summary ──────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════"
echo "  ✅ GoDestino Production Seed Complete"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "  Backend URL:  $1"
echo "  Tenant ID:    $TENANT_ID"
echo "  Kiosk ID:     ${KIOSK_ID:-unknown}"
echo ""
echo "  Admin:  admin@godestino.com / Admin2024x"
echo "  Kiosk:  kiosk@godestino.com / Kiosk2024x"
echo ""
echo "  🔧 To deploy kiosk, run:"
echo ""
echo "  cd frontend/apps/kiosk"
echo "  ./scripts/deploy-production.sh $1 $TENANT_ID ${KIOSK_ID:-<kiosk_id>}"
echo ""
echo "  Or manually set in wrangler.jsonc:"
echo "  NEXT_PUBLIC_DEMO_MODE=false"
echo "  NEXT_PUBLIC_API_URL=$1/api/v1"
echo "  NEXT_PUBLIC_TENANT_ID=$TENANT_ID"
echo "  NEXT_PUBLIC_KIOSK_ID=${KIOSK_ID:-<kiosk_id>}"
echo "  NEXT_PUBLIC_KIOSK_EMAIL=kiosk@godestino.com"
echo "  NEXT_PUBLIC_KIOSK_PASSWORD=Kiosk2024x"
echo ""
