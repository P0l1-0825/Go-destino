# Setup de herramientas de diseño — PoliAgents Hub

## Las 4 herramientas y qué necesitas para cada una

| # | Herramienta | Costo | Setup |
|---|-------------|-------|-------|
| 1 | Google Drive MCP | Gratis | Ya conectado en Cowork |
| 2 | UI UX Pro Max | Gratis | Copiar skill al repo |
| 3 | Nano Banana 2 | ~$0.07/imagen | API key OpenRouter o Google AI Studio |
| 4 | 21st.dev Magic | Plan gratuito/pago | API key en 21st.dev + configurar MCP |

---

## Herramienta 1 — Google Drive MCP

Ya está conectado en Cowork. Para Claude Code, asegúrate de que el conector `google-drive` esté activo en Claude Desktop → Settings → Integrations.

**Cómo usarlo en sesiones de diseño:**
```
Lee el archivo "Design Guidelines Polimentes" en mi Google Drive
```

---

## Herramienta 2 — UI UX Pro Max Skill

Es un skill de Claude Code (no necesita API key).

### Instalar en cada repo que haga UI
```bash
# Opción A — CLI (recomendado)
npm install -g uipro-cli
cd ~/Dev/P0l1-0825-001-MX/{tu-repo}
uipro init --ai claude

# Opción B — Copiar desde el hub (ya tienes el skill aquí)
mkdir -p .claude/skills/ui-ux-pro-max
cp ~/Dev/P0l1-0825-001-MX/poliagents-hub/.claude/skills/ui-ux-pro-max/SKILL.md \
   .claude/skills/ui-ux-pro-max/SKILL.md

# Actualizar el skill cuando salga nueva versión
uipro update  # o repetir el cp anterior
```

### Activación automática
Claude Code lo lee automáticamente cuando detecta trabajo de UI.
También puedes invocarlo explícito:
```
Usando el skill ui-ux-pro-max, planea el design system para la pantalla de [descripción]
del producto Sayo en Flutter.
```

---

## Herramienta 3 — Nano Banana 2

### Paso 1 — Obtener API key

**Opción A: OpenRouter** (recomendado — $0.07/imagen, un key para todos los modelos)
1. Ir a [openrouter.ai](https://openrouter.ai) → Sign up
2. Agregar $5-10 USD en créditos (duran mucho)
3. API Keys → Create → nombrarlo "becm-claude-code"
4. Copiar el key: `sk-or-v1-...`

**Opción B: Google AI Studio** (gratis con límites)
1. Ir a [aistudio.google.com](https://aistudio.google.com)
2. API Keys → Create API Key → copiar

### Paso 2 — Almacenar el key

```bash
# En cada repo donde vas a generar imágenes:
echo "OPENROUTER_API_KEY=sk-or-v1-{tu-key}" >> APIs.env

# O configurar globalmente:
echo 'export OPENROUTER_API_KEY="sk-or-v1-{tu-key}"' >> ~/.zshrc
source ~/.zshrc

# Verificar que APIs.env está en .gitignore (nunca hacer commit del key)
grep "APIs.env" .gitignore || echo "APIs.env" >> .gitignore
```

### Paso 3 — Instalar dependencias del script
```bash
pip install requests pillow --break-system-packages
# o con uv (más rápido):
uv pip install requests pillow
```

### Paso 4 — Copiar el script al repo
```bash
mkdir -p .claude/skills/nano-banana/scripts
cp ~/Dev/P0l1-0825-001-MX/poliagents-hub/.claude/skills/nano-banana/SKILL.md \
   .claude/skills/nano-banana/SKILL.md
cp ~/Dev/P0l1-0825-001-MX/poliagents-hub/.claude/skills/nano-banana/scripts/generate_image.py \
   .claude/skills/nano-banana/scripts/generate_image.py
```

### Paso 5 — Probar
```
En Claude Code:
Genera una imagen de prueba de un escudo de seguridad para fintech.
Estilo minimalista, color teal, fondo transparente.
Guardar en: test-output.png. Resolución 1K.
```

---

## Herramienta 4 — 21st.dev Magic

### Paso 1 — Obtener API key
1. Ir a [21st.dev/magic](https://21st.dev/magic)
2. Sign up (plan gratuito disponible)
3. Console → API Keys → Create
4. Copiar el key

### Paso 2 — Instalar via CLI (recomendado)
```bash
# Instala y configura automáticamente en ~/.claude/mcp.json
npx @21st-dev/cli@latest install --api-key TU_API_KEY
```

### Paso 3 — O configurar manualmente en ~/.claude/mcp.json
```json
{
  "mcpServers": {
    "@21st-dev/magic": {
      "command": "npx",
      "args": ["-y", "@21st-dev/magic@latest", "API_KEY=\"TU_API_KEY\""]
    }
  }
}
```

### Paso 4 — Configurar variable de entorno (opcional, más seguro)
```bash
echo 'export TWENTY_FIRST_API_KEY="tu-key"' >> ~/.zshrc
source ~/.zshrc
```

### Paso 5 — Probar en Claude Code
```
/ui Crea un componente Card para mostrar saldo de cuenta.
    Teal accent, dark background.
    Muestra: nombre del usuario, saldo, número de cuenta (enmascarado).
    React + shadcn/ui + TypeScript + TailwindCSS.
```

---

## Framelink MCP (extra — Figma sin Dev seat)

Si Figma oficial requiere Dev seat y no lo tienes, Framelink es gratuito.

```bash
# Agregar a ~/.claude/mcp.json
# (el hub ya tiene la config en .claude/mcp.json)

# Solo necesitas un Figma Personal Access Token:
# figma.com → Account Settings → Personal access tokens → Create
# Copiar el token

echo 'export FIGMA_API_KEY="tu-token"' >> ~/.zshrc
source ~/.zshrc
```

---

## Verificar que todo funciona

```bash
# En Claude Code, abre cualquier repo de frontend y prueba:

# 1. UI UX Pro Max (debe activarse automáticamente)
"Dime qué estilo visual debo usar para una pantalla de login de Polipay"

# 2. Nano Banana 2
"Genera un ícono abstracto de seguridad. 1K. Guardar en test.png"

# 3. 21st.dev Magic
"/ui Un botón de pago con animación de carga"

# 4. Google Drive
"Busca en mi Google Drive algún doc sobre guidelines de diseño"
```

---

## Resumen de archivos que debes copiar a cada repo de frontend

```bash
REPO="sayo-mobile"  # cambiar por el repo que quieras

cd ~/Dev/P0l1-0825-001-MX/$REPO

# UI UX Pro Max skill
mkdir -p .claude/skills/ui-ux-pro-max
cp ~/Dev/P0l1-0825-001-MX/poliagents-hub/.claude/skills/ui-ux-pro-max/SKILL.md \
   .claude/skills/ui-ux-pro-max/

# Nano Banana 2 skill
mkdir -p .claude/skills/nano-banana/scripts
cp ~/Dev/P0l1-0825-001-MX/poliagents-hub/.claude/skills/nano-banana/SKILL.md \
   .claude/skills/nano-banana/
cp ~/Dev/P0l1-0825-001-MX/poliagents-hub/.claude/skills/nano-banana/scripts/generate_image.py \
   .claude/skills/nano-banana/scripts/

# Los MCPs (21st.dev + Framelink) van en ~/.claude/mcp.json → global para todos los repos
```
