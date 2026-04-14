---
name: nano-banana
description: "Genera imágenes con Google Nano Banana 2 vía OpenRouter o Google AI Studio. Actívate para: generar imagen, crear asset, ilustración, ícono, mockup, thumbnail, hero image, imagen para [producto]. Usa la API key almacenada en APIs.env."
---

# Nano Banana 2 — Generación de imágenes · Grupo BECM / Polimentes

Eres el agente de generación de imágenes para el ecosistema BECM.
Generas assets visuales de alta calidad: íconos de app, imágenes hero, mockups de UI, ilustraciones para onboarding, thumbnails.

---

## Configuración del entorno

### API Key (configurar una sola vez)
```bash
# Opción A — OpenRouter (recomendado: $0.07 por imagen, un solo key para todos los modelos)
# Ir a openrouter.ai → API Keys → Create → copiar key

# Opción B — Google AI Studio (gratis hasta cierto límite)
# Ir a aistudio.google.com → API Keys → Create

# Almacenar el key en el proyecto
echo "OPENROUTER_API_KEY=sk-or-v1-{tu-key}" >> APIs.env
echo "GEMINI_API_KEY={tu-key}" >> APIs.env
# APIs.env está en .gitignore — nunca hacer commit de este archivo
```

### Script de generación
```bash
# Instalar dependencias
pip install requests pillow --break-system-packages

# El script generate_image.py ya está en .claude/skills/nano-banana/scripts/
```

---

## Cómo usar en Claude Code

```
# Generar una imagen nueva:
Genera una imagen de [descripción detallada] usando Nano Banana 2.
Estilo: [fintech dark | minimalista | vibrante | profesional]
Resolución: [1K | 2K | 4K]
Guardar en: assets/images/{nombre}.png

# Editar una imagen existente:
Edita assets/images/hero.png: [descripción del cambio]
Usando Nano Banana 2 con mi OpenRouter key.
```

---

## Prompts optimizados por producto BECM

### Polipay — dark fintech
```
Genera una imagen de fondo abstract para app fintech mexicana.
Estilo: dark gradiente con teal (#00BFA5) y navy (#1A237E).
Sin texto. Formas geométricas sutiles. Alta calidad.
Resolución: 4K. Guardar en: polipay-mobile/assets/hero-background.png
```

### Sayo — crédito
```
Ilustración minimalista de persona recibiendo crédito aprobado.
Estilo flat design. Colores: indigo y blanco. Sin texto.
Apropiado para pantalla de crédito aprobado en app móvil mexicana.
Resolución: 2K. Guardar en: sayo-mobile/assets/credit-approved.png
```

### Novek — eventos
```
Ilustración vibrante de crowd en evento con luz de escenario.
Estilo: bold, energético, púrpura y ámbar.
Para pantalla de selección de boletos.
Resolución: 4K. Guardar en: novek-mobile/assets/event-hero.png
```

### GoDestino — quiosco
```
Ícono flat de destino turístico mexicano (pirámide/playa).
Alto contraste, apropiado para pantalla de quiosco táctil.
Fondo transparente (PNG). Sin texto.
Resolución: 2K. Guardar en: godestino/assets/icons/destination.png
```

### PoliKYC — identidad
```
Ícono profesional de verificación de identidad (escudo + checkmark).
Estilo enterprise SaaS. Colores: slate y green.
Fondo transparente. Sin texto.
Resolución: 2K. Guardar en: polikyc/apps/web/public/icons/kyc-verified.png
```

---

## Reglas de generación

1. **Nunca generar rostros reconocibles** — usar personas abstractas o de espaldas
2. **Nunca incluir texto en las imágenes** — el texto lo agrega la UI
3. **Siempre verificar** que la imagen generada es apropiada para audiencia fintech/corporativa
4. **Optimizar para web/móvil** después de generar:
   ```bash
   # Comprimir para web (mantener calidad)
   python -c "
   from PIL import Image
   img = Image.open('output.png')
   img.save('output-optimized.png', optimize=True, quality=85)
   "
   ```
5. **Actualizar skill** cuando un resultado sea especialmente bueno:
   `Actualiza el skill nano-banana con este prompt exitoso para [producto]`

---

## Costos de referencia (OpenRouter)

| Resolución | Costo aprox. |
|-----------|-------------|
| 1K (~1024px) | ~$0.035 |
| 2K (~2048px) | ~$0.050 |
| 4K (~4096px) | ~$0.070 |

> Para mockups y assets de producción: usar 4K
> Para exploración y pruebas: usar 1K
