#!/bin/bash
# GoDestino — Cloudflare Tunnel Deploy Script
# Ejecutar desde la raíz del proyecto
# Prerequisitos: cloudflared, docker compose, o PostgreSQL + Redis local

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}═══════════════════════════════════════════${NC}"
echo -e "${BLUE}  GoDestino API — Cloudflare Deploy${NC}"
echo -e "${BLUE}═══════════════════════════════════════════${NC}"

# 1. Build
echo -e "\n${GREEN}[1/4] Building Go binary...${NC}"
CGO_ENABLED=0 go build -ldflags="-s -w" -o ./godestino ./cmd/api
echo "  ✓ Binary built: ./godestino"

# 2. Start dependencies (docker or local)
echo -e "\n${GREEN}[2/4] Starting dependencies...${NC}"
if command -v docker &> /dev/null && docker info &> /dev/null; then
    echo "  Using Docker Compose..."
    docker compose up -d postgres redis
    echo "  Waiting for services to be healthy..."
    sleep 5
    # Run migrations
    docker compose exec -T postgres psql -U destino -d destino -f /docker-entrypoint-initdb.d/001_init.sql 2>/dev/null || true
    docker compose exec -T postgres psql -U destino -d destino -f /docker-entrypoint-initdb.d/002_v2_modules.sql 2>/dev/null || true
    export DB_HOST=localhost DB_PORT=5432
else
    echo "  Docker not available. Using local PostgreSQL/Redis..."
    export DB_HOST=localhost DB_PORT=5432
fi

# 3. Start API
echo -e "\n${GREEN}[3/4] Starting GoDestino API...${NC}"
export DB_USER=${DB_USER:-destino}
export DB_PASSWORD=${DB_PASSWORD:-destino}
export DB_NAME=${DB_NAME:-destino}
export DB_SSLMODE=${DB_SSLMODE:-disable}
export REDIS_HOST=${REDIS_HOST:-localhost}
export REDIS_PORT=${REDIS_PORT:-6379}
export JWT_SECRET=${JWT_SECRET:-change-me-in-production}
export JWT_EXPIRE_HOURS=${JWT_EXPIRE_HOURS:-24}
export SERVER_PORT=${SERVER_PORT:-8080}
export APP_ENV=${APP_ENV:-production}

./godestino &
API_PID=$!
sleep 2

# Verify health
if curl -sf http://localhost:${SERVER_PORT}/health > /dev/null; then
    echo "  ✓ API healthy on port ${SERVER_PORT}"
else
    echo "  ✗ API failed to start"
    kill $API_PID 2>/dev/null
    exit 1
fi

# 4. Start Cloudflare Tunnel
echo -e "\n${GREEN}[4/4] Starting Cloudflare Tunnel...${NC}"
if ! command -v cloudflared &> /dev/null; then
    echo "  Installing cloudflared..."
    curl -sL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o /usr/local/bin/cloudflared
    chmod +x /usr/local/bin/cloudflared
fi

echo -e "  ${BLUE}Starting tunnel to http://localhost:${SERVER_PORT}...${NC}"
echo -e "  ${BLUE}Your public URL will appear below:${NC}\n"

cloudflared tunnel --url http://localhost:${SERVER_PORT}

# Cleanup on exit
trap "kill $API_PID 2>/dev/null; echo 'Stopped.'" EXIT
