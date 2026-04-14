---
name: check-pci
description: "Audita cumplimiento PCI DSS v4.0 — tokenizacion, cifrado, logs, webhooks. Score numerico"
---

# /check-pci — PCI DSS v4.0 Compliance Audit

## Contexto
Audita el cumplimiento PCI DSS v4.0 del proyecto actual según CLAUDE.md v2.2.0 §5.

## Instrucciones

Realiza una auditoría exhaustiva de PCI DSS v4.0 verificando todos los controles de seguridad:

### 1. Tokenización y Cifrado (§5.1)
- [ ] **TAPI Integration**: Datos de tarjeta enviados directo a TAPI, NUNCA tocan backend
- [ ] **Patrón KEK/DEK**:
  - KEK almacenado en Railway Secrets (variable de entorno)
  - DEK único por registro (generado con `crypto.randomBytes(32)`)
  - DEK cifrado con KEK antes de almacenar
  - AES-256-GCM usado (NO AES-CBC, NO AES-ECB)
  - IV único por operación (12 bytes random)
  - AuthTag almacenado junto al ciphertext
- [ ] **Rotación de KEK**: proceso documentado y probado
- [ ] **NO almacenar**: PAN completo, CVV/CVC, track data, PIN en ningún formato
- [ ] **Datos permitidos**: últimos 4 dígitos, BIN (primeros 6), fecha expiración tokenizada

### 2. Sanitización de Logs (§5.1.5)
- [ ] **LogSanitizingInterceptor** implementado como interceptor global
- [ ] Patrones detectados y sanitizados:
  - PAN: `\b\d{13,19}\b` con validación Luhn
  - CVV: `\b\d{3,4}\b` en contexto de tarjeta
  - Track data: `%B\d{13,19}` y `;?\d{13,19}=`
  - Tokens JWT en logs
  - API keys y secrets
- [ ] **Validación Luhn** antes de sanitizar (evitar falsos positivos)
- [ ] Sanitización en TODOS los loggers (NestJS Logger, Winston, etc.)
- [ ] Logs de acceso NO contienen datos de tarjeta
- [ ] Logs de error NO contienen stack traces con datos sensibles

### 3. Segregación CDE (§5.1)
- [ ] Servicios que manejan datos de tarjeta aislados
- [ ] Network segmentation documentada
- [ ] Acceso al CDE restringido y auditado
- [ ] Comunicación inter-servicio cifrada (TLS 1.2+)

### 4. Autenticación y Acceso (§5.2)
- [ ] **JWT RS256** (asimétrico, NO HS256)
- [ ] Access token TTL: 15 minutos máximo
- [ ] Refresh token: 7 días con rotación en Redis
- [ ] **JWT Blacklist**: verificación de `jti` en Redis en cada request
- [ ] **Argon2id** para passwords y PINs (NO bcrypt, NO SHA)
- [ ] MFA obligatorio para:
  - Roles administrativos (SUPER_ADMIN, TENANT_ADMIN)
  - Operaciones > $10,000 MXN
  - Acceso al CDE
- [ ] Lockout después de 5 intentos fallidos

### 5. Seguridad de API (§5.3)
- [ ] **Helmet** con CSP completo y HSTS
- [ ] **CORS**: origins explícitos (NUNCA `*` en producción)
- [ ] **ValidationPipe**: `whitelist: true`, `forbidNonWhitelisted: true`
- [ ] **Body size limits**: 256kb general, 512kb webhooks
- [ ] **Rate limiting**: configurado por endpoint según criticidad
- [ ] **Swagger**: deshabilitado en producción
- [ ] **RequestID**: middleware para correlación de logs

### 6. Webhooks (§5.4)
- [ ] HMAC con `crypto.timingSafeEqual()` (NO comparación directa)
- [ ] Timestamp validation (±5 min)
- [ ] Nonce anti-replay (Redis SET NX)
- [ ] Idempotencia verificada ANTES de procesar
- [ ] Validación ANTES de encolar a RabbitMQ
- [ ] DLQ configurada con alertas

### 7. Inter-service Security (§5.5)
- [ ] Service-to-service JWT (60s TTL) o shared secret
- [ ] `InternalServiceGuard` con `timingSafeEqual`
- [ ] Buffer padding a longitud igual antes de comparar
- [ ] Endpoints internos NO expuestos al público

### 8. Auditoría (§5.6)
- [ ] **AuditLog** modelo append-only (sin `updatedAt`)
- [ ] SHA-256 hash para detección de tampering
- [ ] Metadata sanitizada (sin PAN/CVV)
- [ ] Retención de logs según requisitos PCI (mínimo 1 año)

### 9. OWASP Top 10 (§5.7)
- [ ] A01 Broken Access Control: RBAC + tenant isolation
- [ ] A02 Cryptographic Failures: AES-256-GCM, RS256
- [ ] A03 Injection: ValidationPipe + Prisma parameterized
- [ ] A04 Insecure Design: threat modeling documentado
- [ ] A05 Security Misconfiguration: Helmet + CSP + HSTS
- [ ] A06 Vulnerable Components: npm audit + Snyk
- [ ] A07 Auth Failures: Argon2id + MFA + lockout
- [ ] A08 Data Integrity: webhook HMAC + audit hash
- [ ] A09 Logging Failures: log sanitization + audit trail
- [ ] A10 SSRF: URL validation + allowlist

### 10. Secrets Management (§5.8)
- [ ] Railway Secrets para variables sensibles
- [ ] `.env` en `.gitignore`
- [ ] NO secrets hardcodeados en código
- [ ] GitLeaks configurado en CI/CD
- [ ] Rotación de secrets documentada

### Output
Genera un reporte con:
1. **PCI Score**: X/Y controles verificados
2. **Hallazgos CRÍTICOS** (P0): Vulnerabilidades que rompen compliance
3. **Hallazgos ALTOS** (P1): Gaps que necesitan remediación
4. **Hallazgos MEDIOS** (P2): Mejoras recomendadas
5. **Controles OK**: Lista de controles que pasan
6. **Archivos revisados**: lista completa
7. **Plan de remediación**: pasos ordenados por prioridad

Busca en el codebase actual. Si no hay proyecto activo, genera el checklist como referencia.
