#!/usr/bin/env bash
# deploy-railway.sh — Complete Railway deployment guide and automation
# Run this AFTER authenticating: railway login

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "═══════════════════════════════════════════════════════"
echo "  🚂 GoDestino → Railway Deployment"
echo "═══════════════════════════════════════════════════════"
echo ""

# Check Railway CLI
if ! command -v railway &> /dev/null; then
  echo "❌ Railway CLI not installed."
  echo "   Install: brew install railway"
  exit 1
fi

# Check auth
if ! railway whoami &> /dev/null 2>&1; then
  echo "❌ Not logged in to Railway."
  echo "   Run: railway login"
  exit 1
fi

echo "✅ Railway CLI authenticated"
echo ""

cd "$PROJECT_DIR"

# ─── Step 1: Initialize project ───────────────────────────────────────
echo "Step 1/5: Initializing Railway project..."
if [ ! -f ".railway/config.json" ] && [ -z "${RAILWAY_PROJECT_ID:-}" ]; then
  railway init --name godestino-api
  echo "   ✅ Project initialized"
else
  echo "   ⚠️  Project already linked"
fi

# ─── Step 2: Add PostgreSQL ───────────────────────────────────────────
echo ""
echo "Step 2/5: Adding PostgreSQL..."
echo "   ⚠️  MANUAL STEP REQUIRED:"
echo "   1. Go to your Railway dashboard: https://railway.app/dashboard"
echo "   2. Open the 'godestino-api' project"
echo "   3. Click '+ New Service' → 'Database' → 'PostgreSQL'"
echo "   4. Wait for it to provision"
echo ""
read -p "   Press Enter when PostgreSQL is ready..."

# ─── Step 3: Add Redis ────────────────────────────────────────────────
echo ""
echo "Step 3/5: Adding Redis..."
echo "   ⚠️  MANUAL STEP REQUIRED:"
echo "   1. In the same project, click '+ New Service' → 'Database' → 'Redis'"
echo "   2. Wait for it to provision"
echo ""
read -p "   Press Enter when Redis is ready..."

# ─── Step 4: Configure env vars ──────────────────────────────────────
echo ""
echo "Step 4/5: Configuring environment variables..."

# Generate a secure JWT secret
JWT_SECRET=$(openssl rand -hex 32)

railway variables set \
  APP_ENV=production \
  JWT_SECRET="$JWT_SECRET" \
  JWT_EXPIRE_HOURS=720 \
  CORS_ORIGINS="https://godestino-kiosk.direccion-2ac.workers.dev,http://localhost:3000" \
  2>/dev/null || {
    echo "   ⚠️  Could not set variables via CLI. Set them manually:"
    echo "   APP_ENV=production"
    echo "   JWT_SECRET=$JWT_SECRET"
    echo "   JWT_EXPIRE_HOURS=720"
    echo "   CORS_ORIGINS=https://godestino-kiosk.direccion-2ac.workers.dev,http://localhost:3000"
    echo ""
    echo "   NOTE: Railway automatically provides DATABASE_URL and REDIS_URL"
    echo "         when you add PostgreSQL and Redis services."
    read -p "   Press Enter when env vars are set..."
  }

echo "   ✅ Environment configured"
echo "   🔑 JWT Secret: $JWT_SECRET"
echo "   (Save this somewhere safe!)"

# ─── Step 5: Deploy ──────────────────────────────────────────────────
echo ""
echo "Step 5/5: Deploying..."
railway up --detach 2>/dev/null || {
  echo "   ⚠️  'railway up' failed. Try deploying via GitHub integration instead:"
  echo ""
  echo "   1. Go to https://railway.app/dashboard"
  echo "   2. Open the 'godestino-api' project"
  echo "   3. Click '+ New Service' → 'GitHub Repo'"
  echo "   4. Select 'P0l1-0825/Go-destino'"
  echo "   5. Branch: claude/saas-transport-kiosk-app-vfNMZ"
  echo "   6. Railway will auto-detect the Dockerfile and build"
  echo ""
  read -p "   Press Enter when deployed..."
}

# ─── Get deployment URL ──────────────────────────────────────────────
echo ""
echo "Getting deployment URL..."
RAILWAY_URL=$(railway domain 2>/dev/null || echo "")
if [ -z "$RAILWAY_URL" ]; then
  echo "   ⚠️  Could not detect URL automatically."
  echo "   Go to Railway dashboard → Settings → Generate Domain"
  echo ""
  read -p "   Enter your Railway URL (e.g., godestino-api-production.up.railway.app): " RAILWAY_URL
fi

if [[ ! "$RAILWAY_URL" =~ ^https?:// ]]; then
  RAILWAY_URL="https://$RAILWAY_URL"
fi

echo "   ✅ Backend URL: $RAILWAY_URL"

# ─── Health Check ────────────────────────────────────────────────────
echo ""
echo "Checking backend health..."
for i in {1..12}; do
  HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "$RAILWAY_URL/health" 2>/dev/null || echo "000")
  if [ "$HEALTH" = "200" ]; then
    echo "   ✅ Backend is healthy!"
    break
  fi
  echo "   Attempt $i/12: HTTP $HEALTH (waiting 10s...)"
  sleep 10
done

if [ "$HEALTH" != "200" ]; then
  echo "   ❌ Backend not responding. Check Railway logs."
  exit 1
fi

# ─── Seed Production ─────────────────────────────────────────────────
echo ""
echo "Seeding production database..."
bash "$SCRIPT_DIR/seed-production.sh" "$RAILWAY_URL"

echo ""
echo "═══════════════════════════════════════════════════════"
echo "  🎉 Railway Deployment Complete!"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "  Backend: $RAILWAY_URL"
echo "  Health:  $RAILWAY_URL/health"
echo ""
echo "  Next: Update wrangler.jsonc and redeploy kiosk:"
echo "  cd frontend/apps/kiosk"
echo "  # Update wrangler.jsonc env vars (see seed output above)"
echo "  npx turbo build --filter=@godestino/kiosk"
echo "  npx @opennextjs/cloudflare build && npx wrangler deploy"
echo ""
