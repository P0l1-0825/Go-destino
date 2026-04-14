---
name: new-webhook
description: "Genera webhook seguro HMAC + anti-replay + RabbitMQ + DLQ"
allowed-tools: ["Read", "Edit", "Write", "Glob", "Grep", "Bash"]
---

# /new-webhook — Webhook Controller + Consumer Generator

## Contexto
Genera un endpoint de webhook seguro con consumer RabbitMQ siguiendo los estándares PCI DSS v4.0 de Grupo BECM (CLAUDE.md v2.2.0 §5.4).

## Input requerido
- `$ARGUMENTS` — nombre del proveedor/integración (ej: "tapi", "stp", "jaak", "twilio")

## Instrucciones

Genera la siguiente estructura:

```
apps/<app>/src/webhooks/<provider>/
├── <provider>-webhook.module.ts
├── <provider>-webhook.controller.ts
├── <provider>-webhook.service.ts
├── <provider>-webhook.consumer.ts
├── dto/
│   └── <provider>-webhook-payload.dto.ts
├── interfaces/
│   └── <provider>-webhook.interface.ts
└── __tests__/
    ├── <provider>-webhook.controller.spec.ts
    ├── <provider>-webhook.service.spec.ts
    └── <provider>-webhook.consumer.spec.ts
```

### Patrón de seguridad obligatorio (§5.4):

1. **Controller** — Validación ANTES de encolar:
   ```
   Recibir request → Validar timestamp (±5min) → Validar nonce (Redis SET NX 10min)
   → Verificar HMAC (timingSafeEqual) → Verificar idempotencia → Encolar a RabbitMQ → Responder 200
   ```

2. **HMAC Signature**:
   - `crypto.timingSafeEqual()` — NUNCA comparación directa de strings
   - Construir payload: `${timestamp}.${JSON.stringify(body)}`
   - Algorithm configurable (SHA-256 por defecto)
   - Secret desde Railway Secrets / env vars

3. **Anti-replay**:
   - Timestamp: rechazar si `|now - timestamp| > 300000` (5 min)
   - Nonce: `Redis SET NX` con TTL 600s, rechazar si ya existe
   - Idempotencia: verificar `idempotencyKey` en Redis antes de procesar

4. **Consumer RabbitMQ**:
   - Queue dedicada: `webhook.<provider>.process`
   - Validar schema del payload (class-validator)
   - Procesar → Marcar como procesado → ACK
   - Retry con backoff exponencial: 5s, 10s, 20s
   - Máximo 3 reintentos → DLQ: `webhook.<provider>.dlq`
   - Alertar en DLQ (log critical + métricas)

5. **IP Allowlist**:
   - Documentar IPs del proveedor
   - Implementar via Cloudflare Workers o middleware

6. **Rate Limiting**:
   - Endpoint webhook: 100 req/min por IP
   - Body size limit: 512kb (webhooks permiten más que API general)

7. **Logging y Auditoría**:
   - Log de recepción (sin payload sensible)
   - Log de resultado de procesamiento
   - AuditService para cambios de estado
   - NUNCA loggear secrets, tokens, PAN

### Tests:
- Test de validación HMAC (firma válida e inválida)
- Test de anti-replay (timestamp expirado, nonce duplicado)
- Test de idempotencia (request duplicado)
- Test de consumer (proceso exitoso, retry, DLQ)
- Test de rate limiting
- Mock de Redis, RabbitMQ, PrismaService

Genera todos los archivos completos. Pregunta al usuario la configuración del proveedor si no se conoce.
