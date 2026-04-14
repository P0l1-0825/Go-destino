# /test-infrastructure — Prueba Integral de Infraestructura Claude Code

Ejecuta una prueba completa de todos los agentes, comandos y configuraciones instalados en este proyecto.
Valida que cada componente funcione correctamente contra el código real de GoDestino.

## Instrucciones

Ejecuta las siguientes 10 fases en orden. Para cada fase, reporta PASS/FAIL con detalle.
NO modifiques ningún archivo del proyecto. Solo lee, analiza y reporta.

---

### FASE 1: Health Check del Proyecto
- Verifica que CLAUDE.md existe y tiene las secciones: Stack, Arquitectura Backend, Reglas de Código, RBAC, API Conventions, Security
- Verifica que `go.mod` declara module `github.com/P0l1-0825/Go-destino`
- Cuenta: archivos .go en internal/, migraciones SQL, handlers, repos, services
- Reporta la estructura completa del proyecto

### FASE 2: Auditoría de Seguridad (@security-auditor)
Actúa como el agente `@security-auditor` y analiza:
- `internal/middleware/auth.go` — ¿JWT valida correctamente? ¿Inyecta claims en context?
- `internal/middleware/ratelimit.go` — ¿Implementa rate limiting por IP/user?
- `internal/security/password.go` — ¿Usa bcrypt con cost >= 12?
- `internal/security/login_limiter.go` — ¿Protege contra brute force?
- `internal/security/token_blacklist.go` — ¿Soporta logout/revocación?
- `internal/middleware/security_headers.go` — ¿Headers OWASP (X-Frame-Options, CSP, HSTS)?
- Busca SQL injection patterns: strings concatenadas en queries SQL en todos los repos
- Busca PII leaks: ¿Se logean emails, phones o passwords en algún handler/service?

### FASE 3: Revisión de Permisos RBAC (@compliance-specialist)
Actúa como `@compliance-specialist` y valida:
- `internal/domain/permissions.go` — Cuenta las 77 permissions declaradas, verifica que `AllPermissions()` las incluye todas
- Verifica que los 10 roles en `RolePermissions` mapean correctamente
- Verifica que `HasPermission()` funciona con lookup O(n)
- En `internal/router/router.go`, verifica que TODOS los endpoints sensibles usan `applyAuthPerm()`
- Lista cualquier endpoint que NO tenga guard de permisos y debería tenerlo

### FASE 4: Revisión de Código Backend (@code-reviewer)
Actúa como `@code-reviewer` y revisa:
- `cmd/api/main.go` — ¿Bootstrap correcto? repos → services → handlers → router
- Un handler representativo (payment_handler.go) — ¿Sigue el patrón DDD? ¿Multi-tenant filter?
- Un repository representativo (booking_repo.go) — ¿Raw SQL? ¿Filtra por tenant_id? ¿Prepared statements?
- Un service representativo (booking_service.go) — ¿Lógica de negocio separada del handler?
- `pkg/response/` — ¿Envelope estándar {success, data, error}?

### FASE 5: Arquitectura y Patrones (@code-architect)
Actúa como `@code-architect` y evalúa:
- ¿La estructura sigue Domain-Driven Design? domain → repo → service → handler
- ¿Hay dependencias circulares entre packages?
- ¿El router usa standard library http (no Gin/Echo/Fiber)?
- ¿Middleware chain es correcto? (recovery → requestid → logging → cors → security_headers → ratelimit → auth → tenant)
- Evalúa la calidad de las 7 migraciones SQL (001 a 007)

### FASE 6: Base de Datos y Migraciones (@postgres-pro)
Actúa como `@postgres-pro` y analiza:
- Lee todas las migraciones (001_init.sql hasta 007_v7_production_bootstrap.sql)
- ¿Todas las tablas tienen tenant_id? ¿Hay índices compuestos (tenant_id, ...)?
- ¿Hay foreign keys correctas?
- ¿Los tipos de datos son apropiados? (UUID, TIMESTAMPTZ, NUMERIC para dinero, etc.)
- ¿Hay constraints CHECK donde deberían estar? (status enums, amounts > 0)

### FASE 7: API y Endpoints (@api-architect)
Actúa como `@api-architect` y valida:
- Lee `internal/router/router.go` completo
- Lista TODOS los endpoints con su método HTTP, path, permission guard
- ¿Sigue REST conventions? (GET list, GET :id, POST, PUT :id, DELETE :id)
- ¿Health check en GET /health?
- ¿Versionamiento /api/v1/?
- ¿WebSocket endpoints están protegidos?

### FASE 8: Análisis de Performance (@performance-engineer)
Actúa como `@performance-engineer`:
- ¿Los repos usan connection pooling? (Revisa db.go)
- ¿Hay queries N+1 en algún handler que haga loops con queries individuales?
- ¿Rate limiter usa Redis o es in-memory?
- ¿Hay goroutine leaks potenciales? (goroutines sin context/timeout)
- ¿Los handlers paginan resultados? (limit/offset)

### FASE 9: Testing y QA (@test-engineer)
Actúa como `@test-engineer`:
- ¿Existen archivos _test.go?
- Si no existen, genera un PLAN de tests prioritarios:
  1. Tests unitarios para `HasPermission()` (domain layer)
  2. Tests unitarios para `auth middleware` (mock authSvc)
  3. Tests de integración para un handler (payment_handler con httptest)
  4. Tests de seguridad (SQL injection, XSS en inputs, RBAC bypass)
- NO escribas los tests, solo el plan con archivos y funciones específicas

### FASE 10: Reporte Ejecutivo Final
Genera una tabla resumen:

| Fase | Agente | Estado | Issues Críticos | Issues Menores |
|------|--------|--------|-----------------|----------------|

Seguido de:
- **Top 5 Issues Críticos** que deben resolverse antes de producción
- **Top 5 Mejoras Recomendadas** para la siguiente iteración
- **Score de Producción-Readiness** (0-100%)
- **Siguiente paso recomendado** como acción concreta

$ARGUMENTS
