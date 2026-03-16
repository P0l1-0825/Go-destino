# GoDestino — Guía para Claude Code

## Proyecto

SaaS de transporte aeroportuario multi-tenant para LATAM.
Plataforma completa: kioscos de aeropuerto, apps móviles (turista + conductor), dashboard admin web.

## Stack

- **Backend**: Go 1.24, standard library HTTP router (no framework), PostgreSQL 17, Redis 7.4
- **Frontend Web**: Next.js 15 (App Router), TypeScript, Tailwind, shadcn/ui
- **Mobile**: React Native 0.84 + Expo SDK 54
- **Kiosk**: React Native + Android AOSP kiosk mode
- **AI/ML**: Python 3.13 + FastAPI (demand forecasting, dynamic pricing, fraud detection)

## Arquitectura Backend

```
cmd/api/main.go          → Bootstrap (repos → services → handlers → router)
internal/
├── config/              → Env config loading
├── domain/              → Entities, value objects, permissions (FUENTE DE VERDAD)
├── repository/          → PostgreSQL data access (14 repos)
├── service/             → Business logic (12 services)
├── handler/             → HTTP handlers (13 files, 60+ endpoints)
├── middleware/           → Auth JWT, CORS, logging, tenant, rate limit, recovery
└── router/              → Route wiring with RBAC permission guards
migrations/              → PostgreSQL DDL (001_init.sql, 002_v2_modules.sql)
pkg/response/            → Standard API response envelope
```

## Reglas de Código (SIEMPRE seguir)

### Go Backend
- Standard library HTTP. No Gin, no Fiber, no Echo.
- Domain-Driven Design: domain → repository → service → handler
- Multi-tenant: every query MUST filter by tenant_id
- RBAC: 10 roles, 77 permissions. Check via `domain.HasPermission(role, perm)`
- Permission middleware: `applyAuthPerm(authSvc, domain.PermXxx, handler)`
- Errors: use `response.Error(w, status, msg)` — structured JSON envelope
- Audit: `auditSvc.Log()` for all sensitive operations (fire-and-forget goroutine)
- No ORM: raw SQL via `database/sql` + `lib/pq`
- Naming: snake_case for DB columns, camelCase for JSON

### Frontend (TypeScript)
- Import tokens from `src/design-system/tokens.ts` — NEVER hardcode colors
- Functional components + hooks. No class components.
- Naming: PascalCase components, camelCase hooks/utils, UPPER_SNAKE constants
- Dark mode first — dashboard is always dark
- Accessibility: WCAG 2.1 AA (web/mobile), AAA (kiosk)

### React Native
- expo-router for navigation
- SafeAreaView on all screens
- Touch targets: minHeight 44pt iOS / 48dp Android
- Kiosk touch targets: minimum 160×60px
- Offline first: AsyncStorage for critical data, OfflineQueue for mutations

## Design Tokens (fuente de verdad)

### Colores
- navy: #0D1B5E (header, sidebar)
- blue: #2563EB (CTA primario, links activos)
- blueDark: #1D4ED8 (hover primario)
- sky: #38BDF8 (badges info, iconos)
- orange: #E87020 (CTA secundario)
- bg: #040C1F (fondo principal)
- s1: #070F28 (cards nivel 1)
- s2: #0B1535 (cards nivel 2, inputs)
- border: #1C2F62 (bordes, separadores)
- text: #CBD5E1 (párrafos)
- white: #F1F5F9 (títulos)
- success: #10B981 / warning: #F59E0B / error: #EF4444

### Tipografía
- Display: Syne (700/800) — títulos, headers
- Body: Plus Jakarta Sans (300-800) — texto, labels
- Mono: JetBrains Mono (400-700) — código, IDs, precios

## RBAC — 10 Roles

| Rol | Nivel | Scope |
|-----|-------|-------|
| SUPER_ADMIN | 0 | Global — todas las 77 permissions |
| ADMINISTRADOR | 1 | Tenant completo |
| CLIENTE_CONCESION | 2 | Empresa concesionaria |
| TESORERIA_CLIENTE | 2 | Finanzas empresa |
| MESA_CONTROL | 3 | Operaciones aeropuerto |
| OPERADOR | 3 | Operaciones campo |
| TAXISTA | 4 | Conductor individual |
| VENDEDOR | 4 | POS/kiosk seller |
| BROKER | 4 | Integrador API |
| USUARIO | 5 | Turista/pasajero |

## API Conventions

- Base: `/api/v1/`
- Auth: Bearer JWT in Authorization header
- Tenant: X-Tenant-ID header (fallback to JWT claim)
- Response envelope: `{ success: bool, data: T, error?: string, meta?: { page, per_page, total_count } }`
- Error format: `{ success: false, error: "message" }`
- Health: GET /health → `{ status, service, version }`

## Security

- JWT HS256 with role + permissions in claims
- JWT_SECRET validated at startup (min 32 chars, no known defaults → fatalf)
- bcrypt password hashing (cost 12)
- Rate limiting per user_id + IP
- Token blacklist, login limiter, password reset: **Redis-backed** (graceful in-memory fallback)
- Security interfaces: `TokenBlacklistStore`, `LoginLimiterStore`, `PasswordResetTokenStore`
- Redis implementations: `redis_token_blacklist.go`, `redis_login_limiter.go`, `redis_password_reset.go`
- Input validation before any DB query
- Never log PII (emails, phones, full names) — tokens truncated to 8 chars in logs
- Audit log on: payments, cancellations, role changes, tenant operations
- Multi-tenant isolation: ALL SQL queries include `AND tenant_id = $N`

## Environment Variables

See docker-compose.yml. Key vars:
- SERVER_PORT, APP_ENV
- DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
- REDIS_HOST, REDIS_PORT
- JWT_SECRET, JWT_EXPIRE_HOURS

## PoliAgents — Sistema de agentes

GoDestino tiene 14 agentes especializados + 7 slash commands + 3 hooks de seguridad.

### 8 PoliAgents (Opus 4.6)

| Agente | Archivo | Funcion |
|--------|---------|---------|
| PoliOrch | `.claude/agents/agent-poliorch.md` | Orquestador: descompone epics, coordina agentes |
| PoliCode | `.claude/agents/agent-policode.md` | Genera codigo Go, Next.js, React Native |
| PoliSec | `.claude/agents/agent-polisec.md` | Auditoria OWASP, JWT, tenant isolation |
| PoliTest | `.claude/agents/agent-politest.md` | Tests Go table-driven, Jest frontend |
| PoliDeploy | `.claude/agents/agent-polideploy.md` | Railway deploy, Docker, Cloudflare Workers |
| PoliMonitor | `.claude/agents/agent-polimonitor.md` | Health checks, logs, alertas |
| PoliDocs | `.claude/agents/agent-polidocs.md` | CLAUDE.md, README, API docs |
| PoliEdge | `.claude/agents/agent-poliedge.md` | Cloudflare Workers, R2, KV, WAF |

### 6 Agentes especialistas

| Agente | Archivo | Funcion |
|--------|---------|---------|
| Go Developer | `.claude/agents/go-developer.md` | Go 1.24 standard lib, raw SQL, multi-tenant |
| Redis Specialist | `.claude/agents/redis-specialist.md` | go-redis/v9, blacklist, rate limiting |
| Security Auditor | `.claude/agents/security-auditor.md` | OWASP Top 10 para Go backend |
| Code Reviewer | `.claude/agents/code-reviewer.md` | Review PRs, convenciones, seguridad |
| Debugger | `.claude/agents/debugger.md` | Root cause analysis, fix bugs |
| Test Generator | `.claude/agents/test-generator.md` | Tests Go table-driven, httptest |

### 7 Slash Commands

| Comando | Descripcion |
|---------|-------------|
| `/security-audit` | Auditoria de seguridad OWASP completa |
| `/generate-tests` | Genera suite de tests Go para modulo |
| `/new-module` | Genera modulo Go completo (domain→handler) |
| `/check-deploy` | Verifica salud de todos los servicios |
| `/check-tenant-isolation` | Verifica tenant_id en todas las queries |
| `/secrets-scanner` | Escanea codebase por secretos expuestos |
| `/test-infrastructure` | Prueba integral de toda la infraestructura |

### 3 Hooks de seguridad

| Hook | Tipo | Trigger | Funcion |
|------|------|---------|---------|
| `secret-scanner.py` | BLOCKING | PreToolUse:Bash (git commit) | Detecta 50+ tipos de secrets |
| `dangerous-command-blocker.py` | BLOCKING | PreToolUse:Bash | Bloquea rm -rf, dd, mkfs, etc. |
| `tenant-guard.sh` | WARNING | PreToolUse:Edit/Write | Alerta si SQL en repo sin tenant_id |

### Deployment URLs

| Servicio | URL |
|----------|-----|
| API Backend | https://godestino-api-production.up.railway.app |
| Admin Dashboard | https://godestino-admin.direccion-2ac.workers.dev |
| Kiosk | https://godestino-kiosk.direccion-2ac.workers.dev |
| Driver | https://godestino-driver.direccion-2ac.workers.dev |
