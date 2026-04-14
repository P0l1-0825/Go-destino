---
name: new-module
description: "Genera modulo NestJS CRUD completo con multi-tenancy y guards"
allowed-tools: ["Read", "Edit", "Write", "Glob", "Grep", "Bash"]
---

# /new-module — NestJS CRUD Module Generator

## Contexto
Genera un módulo NestJS completo siguiendo los estándares de Grupo BECM (CLAUDE.md v2.2.0 §4).

## Input requerido
- `$ARGUMENTS` — nombre del módulo en singular (ej: "transaction", "user", "kyc-verification")

## Instrucciones

Genera la siguiente estructura para el módulo `$ARGUMENTS`:

```
apps/<app>/src/<module>/
├── <module>.module.ts
├── <module>.controller.ts
├── <module>.service.ts
├── dto/
│   ├── create-<module>.dto.ts
│   └── update-<module>.dto.ts
├── entities/
│   └── <module>.entity.ts
├── interfaces/
│   └── <module>.interface.ts
└── __tests__/
    ├── <module>.controller.spec.ts
    └── <module>.service.spec.ts
```

### Reglas obligatorias (§4 CLAUDE.md):

1. **Controller**:
   - `@Controller('api/v1/<module-plural>')` con versionado
   - `@UseGuards(JwtAuthGuard, RolesGuard, TenantGuard)` en el controller
   - `@Roles()` por endpoint según criticidad
   - `@ApiTags()`, `@ApiOperation()`, `@ApiResponse()` decoradores Swagger
   - Inyectar `requestId` desde headers para correlación
   - Endpoints: GET (list con paginación), GET :id, POST, PATCH :id, DELETE :id (soft delete)

2. **Service**:
   - Inyectar `PrismaService` y `AuditService`
   - Todas las operaciones financieras con idempotencia (Redis SET NX)
   - Soft delete: `where: { deletedAt: null }` en todas las queries
   - Multi-tenant: filtrar siempre por `tenantId` del JWT payload
   - Logging con `Logger` de NestJS, NUNCA console.log
   - Emitir eventos de auditoría en create/update/delete

3. **DTOs**:
   - `class-validator` con `whitelist: true`
   - `@IsUUID()`, `@IsString()`, `@IsNumber({maxDecimalPlaces: 2})` según tipo
   - `@Matches()` para CURP, RFC, CLABE cuando aplique
   - Separar CreateDto y UpdateDto (UpdateDto extends PartialType(CreateDto))

4. **Tests**:
   - Mock de PrismaService, AuditService, Redis
   - Cobertura mínima 80%
   - Test de guards (auth, roles, tenant)
   - Test de validación de DTOs

5. **Prisma model** (agregar a schema.prisma):
   - `id String @id @default(cuid())`
   - `tenantId String` con índice
   - `createdAt DateTime @default(now())`
   - `updatedAt DateTime @updatedAt`
   - `deletedAt DateTime?` (soft delete)
   - `@@map("snake_case_plural")`

### Seguridad (§5):
- NO exponer IDs internos en responses — usar DTOs de respuesta
- Sanitizar logs (no loggear PAN, CVV, tokens)
- Rate limiting según criticidad del endpoint
- Body size limit 256kb

Genera todos los archivos completos y funcionales. Pregunta al usuario el nombre de la app Nx si hay múltiples apps.
