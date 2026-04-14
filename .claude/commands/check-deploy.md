---
allowed-tools: Read, Bash, Grep
argument-hint: [--staging | --production]
description: Verifica salud de todos los servicios GoDestino desplegados
---

# /check-deploy — Verificacion de deploy GoDestino

Verifica estado de todos los servicios: $ARGUMENTS

## Checks

### 1. Build local
```bash
cd Go-destino && go build ./... && go test ./... -count=1 && go vet ./...
```

### 2. Health checks produccion
```bash
echo "=== API Backend ==="
curl -sf https://godestino-api-production.up.railway.app/health | python3 -m json.tool

echo "=== Admin Dashboard ==="
curl -sf -o /dev/null -w "Status: %{http_code}\n" https://godestino-admin.direccion-2ac.workers.dev

echo "=== Kiosk ==="
curl -sf -o /dev/null -w "Status: %{http_code}\n" https://godestino-kiosk.direccion-2ac.workers.dev

echo "=== Driver ==="
curl -sf -o /dev/null -w "Status: %{http_code}\n" https://godestino-driver.direccion-2ac.workers.dev
```

### 3. Redis status
```bash
railway logs --service godestino-api --tail 10 2>/dev/null | grep -i "redis"
```

### 4. Test login
```bash
curl -s -X POST https://godestino-api-production.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{"email":"admin@godestino.com","password":"Admin2024x"}' | python3 -c "
import json,sys
d=json.load(sys.stdin)
print('Login:', 'OK' if d.get('success') else 'FAIL')
print('Role:', d.get('data',{}).get('user',{}).get('role','?'))
"
```

### 5. Recent logs (errores)
```bash
railway logs --service godestino-api --since 1h 2>/dev/null | grep -iE "error|panic|fatal" | tail -10
```

## Reporte

```
# Deploy Check — GoDestino ({fecha})

| Servicio | Status |
|----------|--------|
| Build local | OK/FAIL |
| Tests | OK/FAIL |
| API Health | 200/DOWN |
| Admin | 200/DOWN |
| Kiosk | 200/DOWN |
| Driver | 200/DOWN |
| Redis | Connected/Fallback |
| Login | OK/FAIL |
| Errors (1h) | N errores |
```
