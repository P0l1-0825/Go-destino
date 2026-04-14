---
allowed-tools: Read, Grep, Glob
description: Verifica que todas las queries SQL filtran por tenant_id
---

# /check-tenant-isolation — Verificacion multi-tenant GoDestino

Escanea todo el codigo para verificar aislamiento multi-tenant.

## Checks

### 1. Queries SQL en repositories
Busca TODAS las operaciones SQL (SELECT, INSERT, UPDATE, DELETE) en `internal/repository/` y verifica que incluyen `tenant_id`:

```bash
# Listar todos los archivos de repositorio
find internal/repository/ -name "*.go" -not -name "*_test.go"
```

Para cada archivo:
- Buscar lineas con SELECT, INSERT, UPDATE, DELETE
- Verificar que incluyen `tenant_id` como parametro
- Reportar lineas que NO incluyen tenant_id

### 2. Handler context extraction
Verificar que todos los handlers extraen tenant_id del context:
```bash
grep -rn "tenant_id" internal/handler/ | head -20
```

### 3. Middleware tenant injection
Verificar que el middleware de tenant inyecta tenant_id en el context desde JWT claims.

## Reporte

```
# Tenant Isolation Check — GoDestino

## Repositories escaneados: N
## Queries verificadas: N
## Queries con tenant_id: N
## Queries SIN tenant_id: N (ALERTA)

### Detalle de violaciones
- file:line — query sin tenant_id

### Veredicto: PASS | FAIL
```

Las unicas excepciones permitidas son:
- Tabla `tenants` (es la tabla raiz)
- Queries de migration/bootstrap
- Health check endpoint
