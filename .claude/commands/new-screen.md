---
name: new-screen
description: "Genera screen Angular/Flutter/React con tokens de diseno BECM"
allowed-tools: ["Read", "Edit", "Write", "Glob", "Grep", "Bash"]
---

# /new-screen — Angular/React Screen Skeleton Generator

## Contexto
Genera una pantalla/componente siguiendo los estándares de UI/UX de Grupo BECM (CLAUDE.md v2.2.0 §9).

## Input requerido
- `$ARGUMENTS` — nombre de la pantalla y framework (ej: "dashboard angular", "user-profile react", "kyc-form flutter")

## Instrucciones

### Detectar framework del argumento:
- Si contiene "angular" → generar componente Angular 20 + Material 3
- Si contiene "react" → generar componente React 19 + shadcn/ui
- Si contiene "flutter" → generar widget Flutter 3.x + Riverpod
- Si no especifica → preguntar al usuario

---

### Angular 20 + Material 3 (B2B):

```
libs/ui/src/lib/<screen>/
├── <screen>.component.ts        (standalone component)
├── <screen>.component.html
├── <screen>.component.scss
├── <screen>.component.spec.ts
└── <screen>.routes.ts           (lazy loaded route)
```

Reglas:
- Standalone components (NO NgModules)
- Angular Material 3 components
- NgRx para state management si la pantalla maneja estado complejo
- Signals para estado local reactivo
- Design tokens: Primary `#00C9A7`, Secondary `#1A1A2E`, Font `Inter`
- WCAG 2.1 AA: contraste mínimo 4.5:1, alt text, aria-labels
- Responsive: mobile-first, breakpoints Material
- i18n ready con `@angular/localize`
- Lazy loading via route config

---

### React 19 + shadcn/ui (SaaS):

```
apps/<app>/src/features/<screen>/
├── <screen>.tsx                 (page component)
├── components/
│   └── <screen>-form.tsx        (si aplica)
├── hooks/
│   └── use-<screen>.ts
├── <screen>.test.tsx
└── index.ts
```

Reglas:
- Functional components con hooks
- shadcn/ui para componentes base
- TailwindCSS 3.x para estilos
- Zustand para state management si necesario
- React Hook Form + Zod para formularios
- Design tokens via CSS variables: `--color-primary: #00C9A7`
- WCAG 2.1 AA compliance
- Responsive con Tailwind breakpoints
- Error boundaries para manejo de errores

---

### Flutter 3.x (Mobile):

```
lib/features/<screen>/
├── presentation/
│   ├── pages/
│   │   └── <screen>_page.dart
│   ├── widgets/
│   │   └── <screen>_widget.dart
│   └── controllers/
│       └── <screen>_controller.dart
├── domain/
│   └── models/
│       └── <screen>_model.dart
└── data/
    └── repositories/
        └── <screen>_repository.dart
```

Reglas:
- Feature-based architecture bajo `lib/features/`
- Riverpod 2.x para state management
- Material 3 widgets
- Design tokens centralizados
- Responsive con `LayoutBuilder` / `MediaQuery`
- Accesibilidad: `Semantics` widgets

---

### Reglas UX comunes (§9):
- Fintech UX: confirmación explícita en operaciones financieras (2-step)
- Loading states: skeleton screens, no spinners genéricos
- Error states: mensajes claros en español, acción sugerida
- Empty states: ilustración + mensaje + CTA
- Montos MXN: formato `$1,234.56 MXN` con `Intl.NumberFormat('es-MX')`
- Inputs financieros: validación en tiempo real, formateo automático
- CLABE: formato `XXX XXX XXXXXXXXXXX X` con validación
- Animaciones: 200ms ease-in-out para transiciones, 300ms para modales

Genera todos los archivos completos y funcionales.
