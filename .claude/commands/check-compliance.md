---
name: check-compliance
description: "Audita regulacion mexicana — CNBV, UIF, LFPDPPP, LFPIORPI, niveles KYC"
---

# /check-compliance — Mexican Regulatory Compliance Validator

## Contexto
Audita el cumplimiento regulatorio mexicano del proyecto actual según CLAUDE.md v2.2.0 §8.

## Instrucciones

Realiza una auditoría completa de cumplimiento regulatorio revisando:

### 1. Ley Fintech / CNBV / UIF
- [ ] Licencias y registros requeridos documentados
- [ ] Reportes regulatorios implementados (frecuencia y formato)
- [ ] Límites de operación por nivel KYC respetados:
  - Nivel 0: $1,000 MXN (sin verificación)
  - Nivel 1: $8,000 MXN (datos básicos)
  - Nivel 2: $50,000 MXN (INE + comprobante)
  - Nivel 3: Sin límite (verificación completa + videollamada)
- [ ] Validación de límites enforced en backend (NO solo frontend)
- [ ] Reportes de operaciones inusuales (UIF) implementados
- [ ] Retención de datos según requerimientos CNBV

### 2. LFPDPPP — Protección de Datos Personales
- [ ] Aviso de privacidad completo y accesible
- [ ] Derechos ARCO implementados:
  - **A**cceso: endpoint para consultar datos personales
  - **R**ectificación: endpoint para corregir datos
  - **C**ancelación: endpoint para solicitar eliminación
  - **O**posición: endpoint para limitar uso
- [ ] Consentimiento explícito recopilado y almacenado
- [ ] Responsable de datos personales designado
- [ ] Transferencias a terceros documentadas y consentidas
- [ ] Datos sensibles (biométricos, financieros) con protección adicional

### 3. LFPIORPI — Anti-Lavado (AML)
- [ ] Identificación del cliente (KYC) implementada
- [ ] Monitoreo de operaciones sospechosas
- [ ] Reportes a UIF configurados
- [ ] Listas negras / PEP verificadas
- [ ] Perfil transaccional del cliente establecido
- [ ] Alertas por operaciones fuera de perfil

### 4. Validadores mexicanos (@becm/mx-validators)
Verificar que existan y se usen correctamente:
- [ ] **CURP**: 18 caracteres, formato `[A-Z]{4}[0-9]{6}[HM][A-Z]{5}[A-Z0-9]{2}`
- [ ] **RFC**: 12-13 caracteres con dígito verificador
- [ ] **CLABE**: 18 dígitos con dígito verificador (algoritmo módulo 10)
- [ ] **Código Postal**: 5 dígitos, validar contra catálogo SEPOMEX
- [ ] **Número celular**: 10 dígitos, prefijo válido
- [ ] **INE/IFE**: formato de CIC y OCR

### 5. Niveles KYC — Implementación
```
Verificar flujo completo:
Nivel 0 → Registro mínimo (email + teléfono)
Nivel 1 → Datos básicos (nombre, CURP, RFC)
Nivel 2 → Documentos (INE + comprobante domicilio)
Nivel 3 → Verificación completa (videollamada + validación biométrica)
```
- [ ] Cada nivel tiene límites de operación enforced
- [ ] Upgrade de nivel requiere verificación adicional
- [ ] Downgrade no permitido sin intervención administrativa
- [ ] `@RequireKycLevel()` decorator usado en endpoints financieros

### 6. Formato y localización
- [ ] Montos en MXN con formato correcto (`$1,234.56 MXN`)
- [ ] Fechas en zona horaria `America/Mexico_City`
- [ ] Mensajes de error en español
- [ ] RFC en facturación con validación SAT
- [ ] CFDI 4.0 si aplica facturación electrónica

### Output
Genera un reporte con:
1. **Score de cumplimiento**: X/Y checks pasados
2. **Hallazgos críticos**: issues que deben resolverse inmediatamente
3. **Hallazgos menores**: mejoras recomendadas
4. **Archivos revisados**: lista de archivos analizados
5. **Recomendaciones**: pasos concretos para remediar cada hallazgo

Busca en el codebase actual los archivos relevantes. Si no hay proyecto activo, indica qué debería existir.
