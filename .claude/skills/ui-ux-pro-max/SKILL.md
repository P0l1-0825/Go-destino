---
name: ui-ux-pro-max
description: "UI/UX design intelligence para Grupo BECM / Polimentes. 50+ estilos, 97 paletas, 99 UX guidelines, Flutter, Angular, React. Actívate para: plan, build, create, design, implement, review, fix, improve, optimize en screens, dashboards, landing pages, mobile apps. Productos: Polipay (fintech dark), Sayo (crédito moderno), Novek (eventos/cashless), PoliKYC (SaaS enterprise), Polipay IA (customer success teal), GoDestino (quioscos)"
---

# UI/UX Pro Max — Design Intelligence · Grupo BECM / Polimentes

Eres un experto en diseño de interfaces financieras y SaaS con enfoque en productos mexicanos.
Cuando se solicita cualquier tarea de UI/UX, primero planeas el sistema de diseño y luego generas el código.

---

## Identidad visual por producto

| Producto | Plataforma | Estilo | Paleta principal | Tipografía |
|----------|-----------|--------|-----------------|------------|
| **Polipay** | Flutter + Angular | Dark fintech, glassmorphism sutil | Teal `#00BFA5` + Navy `#1A237E` | Inter / Roboto |
| **Sayo** | Flutter | Moderno, minimalista, trust | Indigo `#3949AB` + White | Inter |
| **Vialpay** | Flutter | Clean, gubernamental accesible | Blue `#1976D2` + Gray | Roboto |
| **Novek** | Angular + Flutter | Bold, vibrante, energía de evento | Purple `#7B1FA2` + Amber `#FFB300` | Montserrat |
| **PoliKYC** | Angular | Enterprise SaaS, profesional | Slate `#37474F` + Green `#43A047` | Inter |
| **Polipay IA** | React | Customer success, cálido | Teal `#009688` + Emerald `#00BCD4` | Inter |
| **GoDestino** | React/HTML (quiosco) | Kiosk-first, táctil, alta legibilidad | Orange `#F4511E` + Dark `#212121` | Roboto |

---

## Proceso de diseño obligatorio (ejecuta en orden)

### Paso 1 — Identificar contexto
```
Producto: [Polipay | Sayo | Vialpay | Novek | PoliKYC | Polipay IA | GoDestino]
Plataforma: [Flutter | Angular | React | HTML kiosk]
Tipo: [screen móvil | dashboard web | landing | quiosco | componente]
Estado: [crear desde cero | mejorar existente | revisar accesibilidad]
```

### Paso 2 — Buscar en base de conocimiento de diseño
Antes de generar código, consulta:
1. **Estilo visual** apropiado para el producto (ver tabla arriba)
2. **Paleta** — usar tokens del producto, nunca colores hardcodeados
3. **Tipografía** — escala tipográfica consistente (h1 24px, h2 20px, body 14px, caption 12px)
4. **Layout pattern** — tarjeta / lista / form / dashboard / kiosk
5. **Accesibilidad** — contraste WCAG AA mínimo (ratio 4.5:1 para texto normal)

### Paso 3 — Generar design system snippet para este feature
```dart
// Flutter: tokens del producto
class PolipayTokens {
  static const primary = Color(0xFF00BFA5);
  static const onPrimary = Colors.white;
  static const surface = Color(0xFF1E1E2E);
  static const onSurface = Color(0xFFE0E0E0);
  static const error = Color(0xFFCF6679);
}
```

```typescript
// Angular/React: CSS custom properties
:root {
  --color-primary: #00BFA5;
  --color-surface: #1E1E2E;
  --color-on-surface: #E0E0E0;
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 20px;
}
```

### Paso 4 — Generar código con estas reglas

#### Flutter
- Usar `ThemeData` con `ColorScheme.dark()` para productos con dark mode
- `skeleton_loader` para estados de carga — nunca `CircularProgressIndicator` solo
- Responsive: `LayoutBuilder` + breakpoints 360 / 768 / 1024
- Animaciones: `AnimatedContainer`, `Hero`, `FadeTransition` — no packages externos
- Mock data SOLO en `lib/core/mock/` — nunca inline

#### Angular
- `Angular Material 3` con theming via `mat.define-theme()`
- Standalone components — nunca NgModules
- Signals para estado reactivo
- `@defer` para lazy loading de secciones pesadas
- ARIA labels en todos los elementos interactivos

#### React (Polipay IA)
- `shadcn/ui` + `TailwindCSS 3.x`
- Componentes en `src/components/ui/` (atómicos) y `src/components/` (moleculares)
- `Framer Motion` para animaciones (ya incluido en Polipay IA)
- Dark mode via `class="dark"` en el root

---

## 50+ Estilos disponibles — cuándo usarlos

| Estilo | Cuándo usar en BECM |
|--------|---------------------|
| **Glassmorphism** | Polipay dashboard — overlays de saldo, tarjetas de crédito |
| **Dark fintech** | Polipay móvil — pantallas de wallet, transferencias |
| **Minimalism** | Sayo — flujo de crédito, onboarding limpio |
| **Bento grid** | Dashboards PoliKYC — métricas multi-panel |
| **Enterprise SaaS** | PoliKYC web — tablas de tenants, configuración |
| **Kiosk / Large touch** | GoDestino — botones >60px, alta legibilidad, sin hover states |
| **Event / Bold** | Novek — sell-out urgency, countdown timers |
| **Customer success** | Polipay IA — chat UI, sentiment badges, warm palette |

---

## Checklist de calidad UI antes de generar

- [ ] Contraste texto/fondo ≥ 4.5:1 (texto normal) o 3:1 (texto grande)
- [ ] Toque mínimo 48x48dp en Flutter, 44px en web
- [ ] Skeleton loader en todos los estados de carga asíncrona
- [ ] Estado vacío diseñado (no solo "No hay datos")
- [ ] Error state con mensaje accionable
- [ ] Modo oscuro consistente (si el producto lo usa)
- [ ] `flutter analyze` o ESLint sin errores
- [ ] Sin colores hardcodeados — solo tokens o variables CSS

---

## Integración con otras herramientas del hub

- **Con Figma MCP:** primero lee el frame con `get_design_context`, luego consulta este skill para mapear al stack correcto
- **Con Nano Banana 2:** para generar imágenes/ilustraciones que completen la UI (íconos, hero images, placeholders realistas)
- **Con 21st.dev Magic:** `/ui` genera el componente React base, este skill le da el contexto de paleta y estilo correcto para Polipay IA
- **Con PoliCode:** después del diseño, PoliCode agrega la lógica de negocio (Riverpod, guards, tenantId)
