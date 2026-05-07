# Diego Noticias — Plan de implementación maestro

**Versión**: 1.0
**Fecha**: 2026-05-06
**Autor del plan**: definido en colaboración con el dueño del proyecto
**Audiencia**: agente de IA (Cursor Auto) y desarrolladores humanos

---

## ÍNDICE

1. [Instrucciones obligatorias para el agente](#1-instrucciones-obligatorias-para-el-agente)
2. [Resumen del proyecto](#2-resumen-del-proyecto)
3. [Decisiones cerradas (no negociar)](#3-decisiones-cerradas-no-negociar)
4. [Arquitectura general](#4-arquitectura-general)
5. [Estructura del repositorio](#5-estructura-del-repositorio)
6. [Dependencias y versiones](#6-dependencias-y-versiones)
7. [Modelo de datos](#7-modelo-de-datos)
8. [Sitio público (Hugo)](#8-sitio-público-hugo)
9. [Admin SPA (Vue 3 + TypeScript)](#9-admin-spa-vue-3--typescript)
10. [Backend Go](#10-backend-go)
11. [Integración con Groq](#11-integración-con-groq)
12. [Pipeline de imágenes](#12-pipeline-de-imágenes)
13. [Sistema de publicidad](#13-sistema-de-publicidad)
14. [Autenticación y seguridad](#14-autenticación-y-seguridad)
15. [SEO](#15-seo)
16. [Despliegue](#16-despliegue)
17. [Variables de entorno](#17-variables-de-entorno)
18. [Convenciones de código](#18-convenciones-de-código)
19. [Fases de implementación](#19-fases-de-implementación)
20. [Checklist de open source](#20-checklist-de-open-source)
21. [Glosario](#21-glosario)

---

## 1. Instrucciones obligatorias para el agente

> **Lee esta sección completa antes de tocar código. Vuelve a ella cuando dudes.**

### 1.1 Cómo proceder

1. **Trabaja por fases** (sección 19). No empieces la fase N+1 hasta que la fase N tenga todos los checks ✅. Cada fase tiene su propio "Definition of Done".
2. **Lee primero la fase completa**, identifica archivos a tocar, y solo después escribe código.
3. **Antes de implementar algo no especificado aquí, detente y pregunta al dueño**. No improvises features, dependencias ni rutas.
4. **Cuando termines una sub-tarea**, ejecuta su verificación (tests / curl / Lighthouse / etc.) y marca el check.
5. Si **una verificación falla**, depura — no avances ni "deshabilites" la verificación.
6. Cuando termines una fase, **resume al dueño en 2–3 líneas qué quedó hecho y qué sigue**, no sueltes diffs largos.

### 1.2 Reglas duras (NO violar)

- ❌ **No introduzcas dependencias no listadas en sección 6** sin pedir permiso. Si crees que falta una, propón antes.
- ❌ **No metas Node.js como dependencia de runtime en el VPS**. Node solo durante build del admin SPA y debe quedar fuera del binario final.
- ❌ **No uses MySQL ni ninguna base de datos**. La fuente de verdad es el filesystem (Markdown + JSON).
- ❌ **No agregues tests inflados**. Tests sí, pero solo donde aporten (lógica de negocio: slugs, parsing frontmatter, validaciones de banners, prompt builder, image pipeline). UI tests no.
- ❌ **No comentes código obvio**. Solo comenta el *porqué* cuando no es deducible del código.
- ❌ **No metas features fuera del spec**: nada de "tags", "comentarios", "versiones", "drafts publicables a futuro", "multi-idioma", "multi-usuario", "favoritos", etc.
- ❌ **No hagas `git push`, `git commit --amend`, `git reset --hard` ni borres ramas** sin que el dueño lo autorice expresamente para cada acción.
- ❌ **No subas secretos al repo jamás**. Si dudas si un valor es secreto, asume que sí.
- ❌ **No uses `panic` en código que sirve HTTP**. Devuelve errores con códigos correctos.
- ❌ **No silencies errores con `_`** salvo que la intención sea explícita y obvia (ej. `defer f.Close()`).
- ❌ **No uses CSS-in-JS**. Tailwind utilitario en plantillas Hugo y en componentes Vue.
- ❌ **No uses framework UI de Vue (Vuetify, PrimeVue, etc.)**. Tailwind + componentes propios.

### 1.3 Reglas blandas (preferencias fuertes)

- ✅ Errores en español cuando van al usuario; logs internos en inglés.
- ✅ Idempotencia: cualquier `POST /publicar` debe poder reintentarse sin romper estado.
- ✅ "Atomic write" para archivos críticos: escribir a `.tmp`, luego `rename`.
- ✅ Logs estructurados (`log/slog`) con nivel.
- ✅ Confías en framework guarantees. No agregues validaciones redundantes en cada capa.
- ✅ Mobile-first en CSS. Desktop = breakpoints `md:` y `lg:`.
- ✅ Accesibilidad: roles ARIA, `alt` en imágenes, contraste AA, navegación por teclado.

### 1.4 Si te bloqueas

Detente. Resume el bloqueo en 5 líneas: qué intentaste, qué falló, hipótesis del problema, qué necesitas. No "rodees" un problema con hacks.

---

## 2. Resumen del proyecto

**Diego Noticias** es un blog/sitio de noticias minimalista con dos componentes:

- **Sitio público**: estático, ultra-rápido, SEO óptimo, generado con Hugo, servido por un binario Go vía nginx.
- **Admin**: SPA mínima (CRUD de artículos + CRUD de publicidad + ajustes) protegida por contraseña, embebida en el mismo binario Go.

El flujo central de creación de artículo:
1. El usuario admin pega un texto crudo (~40 palabras) en un textarea, escoge un tono y opcionalmente sube una imagen y/o título.
2. El backend llama a Groq con un prompt estructurado y devuelve **título + cuerpo (170–230 palabras) + meta description + categoría + alt** en una sola petición JSON con streaming.
3. El admin ve el resultado aparecer en vivo, puede editar y luego publicar.
4. Al publicar, el backend escribe un Markdown en `site/content/articulos/{slug}.md`, dispara `hugo` y `pagefind`, y el sitio público se actualiza.

**Prioridades en orden**: SEO > velocidad > responsive > UI bonita y minimalista > accesibilidad.

---

## 3. Decisiones cerradas (no negociar)

| Área | Decisión |
|---|---|
| Generador estático | **Hugo extended (último estable)** |
| CSS | **Tailwind CSS v4 vía Hugo Pipes (binario standalone)** + `@apply` selectivo en `assets/css/main.css` |
| JS público | **Alpine.js** para interactividad puntual; **Pagefind** para búsqueda |
| Admin SPA | **Vue 3 + TypeScript + Vite + Pinia + Vue Router**, embebida con `go:embed` |
| Backend | **Go** (último estable). Binario único. |
| Almacenamiento | **Filesystem**: Markdown + JSON. **Sin base de datos.** |
| IA | **Groq**, modelo `llama-3.3-70b-versatile` |
| Streaming IA | **SSE** del backend al admin con JSON parcial parseado en cliente |
| Imágenes | **govips** (libvips) → AVIF + WebP en 4 tamaños (320/640/1024/1600) |
| Reverse proxy / TLS | **nginx** existente del VPS |
| Auth | **Argon2id** + cookie HTTP-only de sesión + token CSRF |
| Idioma | `es-CO`. Mensajes UI en español. |
| Permalinks | `/:slug/` con trailing slash |
| Paginación | 12 artículos por página |
| Licencia | **MIT** |
| Analíticas | **GoAccess** parseando logs de nginx |
| Cap IA | **100 generaciones/día** (env var) |
| Username admin | `diego` |
| AdSense | **deshabilitado por defecto** (toggle en ajustes) |
| Banners propios | máx **2 activos simultáneos**, máx **5 inactivos** (7 totales) |
| Bottom nav móvil | **Inicio · Buscar · Admin** (3 items) |

### 3.1 Lista cerrada de tonos

```
informativo, profesional, institucional, academico, cronica,
editorial, conversacional, pedagogico, dramatico, sensacionalista
```

Cada tono tiene una descripción corta (sección 11.2) que se inyecta en el prompt.

### 3.2 Categorías

**No hay lista cerrada**. La IA devuelve **una palabra** que describa el tema del artículo. Se guarda en frontmatter como string libre. La categoría se muestra como label en la UI pero no tiene página índice ni lógica.

---

## 4. Arquitectura general

```
                        ┌─────────────────────────────────────────────┐
   Internet ── nginx ───┤   Go binary (127.0.0.1:8080)                │
   (TLS, gzip, brotli)  │                                             │
                        │   ├─ GET /              → site/public/*     │
                        │   ├─ GET /admin/*       → SPA Vue (embed)   │
                        │   ├─ * /admin/api/*     → handlers JSON     │
                        │   ├─ GET /images/*      → static-uploads/*  │
                        │   ├─ GET /ads/*         → data/ads/*        │
                        │   └─ GET /pagefind/*    → site/public/pf/*  │
                        └────────────┬────────────────────────────────┘
                                     │
        ┌────────────────────────────┼─────────────────────────────────┐
        │                            │                                 │
   site/content/              data/ (no-repo)                    Groq API
   articulos/*.md             ├─ admin.json   (hash password)    (HTTPS)
   (source of truth)          ├─ settings.json
                              ├─ ads.json
                              └─ ads/{id}/*.{avif,webp}
                                     │
                                     ▼
                           hugo build ──→ site/public/
                           pagefind   ──→ site/public/pagefind/
```

Un único proceso Go.
- **Lectura pública**: nginx → Go → archivos en `site/public/`.
- **Escritura admin**: SPA → API JSON Go → escribe `.md` y `data/*.json` → invoca `hugo` + `pagefind`.

---

## 5. Estructura del repositorio

```
diegonoticias/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/               # handlers HTTP del admin
│   │   ├── articles.go
│   │   ├── ads.go
│   │   ├── auth.go
│   │   ├── images.go
│   │   ├── settings.go
│   │   ├── ai.go
│   │   ├── build.go
│   │   ├── middleware.go
│   │   └── router.go
│   ├── articles/          # leer/escribir/listar .md
│   │   ├── article.go
│   │   ├── frontmatter.go
│   │   ├── slug.go
│   │   ├── store.go
│   │   └── slug_test.go
│   ├── ai/                # cliente Groq + prompt
│   │   ├── client.go
│   │   ├── prompt.go
│   │   ├── stream.go
│   │   └── prompt_test.go
│   ├── auth/              # sesiones, login, CSRF
│   │   ├── auth.go
│   │   ├── session.go
│   │   ├── password.go
│   │   └── csrf.go
│   ├── builder/           # invoca hugo + pagefind
│   │   └── builder.go
│   ├── images/            # libvips, srcsets, compresión
│   │   ├── pipeline.go
│   │   └── pipeline_test.go
│   ├── ads/               # CRUD de banners
│   │   ├── store.go
│   │   ├── validate.go
│   │   └── validate_test.go
│   ├── settings/
│   │   └── store.go
│   ├── seo/               # helpers de JSON-LD, RSS extra
│   │   └── jsonld.go
│   ├── ratelimit/
│   │   └── ratelimit.go
│   └── config/
│       └── config.go
├── site/                          # proyecto Hugo
│   ├── hugo.toml
│   ├── archetypes/
│   │   └── articulos.md
│   ├── content/
│   │   └── articulos/             # generados por el admin
│   ├── layouts/
│   │   ├── _default/
│   │   │   ├── baseof.html
│   │   │   ├── single.html
│   │   │   └── list.html
│   │   ├── articulos/
│   │   │   └── single.html
│   │   ├── partials/
│   │   │   ├── head/
│   │   │   │   ├── meta.html
│   │   │   │   ├── og.html
│   │   │   │   └── jsonld.html
│   │   │   ├── header.html
│   │   │   ├── footer.html
│   │   │   ├── article-card.html
│   │   │   ├── article-card-hero.html
│   │   │   ├── ad-slot.html
│   │   │   ├── pagination.html
│   │   │   └── search.html
│   │   ├── shortcodes/
│   │   ├── index.html
│   │   ├── 404.html
│   │   └── robots.txt
│   ├── assets/
│   │   ├── css/
│   │   │   └── main.css
│   │   └── js/
│   │       ├── main.js            # Alpine init + nav móvil
│   │       └── search.js          # cliente Pagefind
│   ├── static/
│   │   ├── favicon.ico
│   │   ├── apple-touch-icon.png
│   │   ├── site.webmanifest
│   │   └── og-default.jpg
│   ├── data/                      # generado por Go en cada build
│   │   ├── site.toml              # ajustes para plantillas
│   │   └── ads.toml               # banners activos para plantillas
│   └── public/                    # OUTPUT (gitignore)
├── web/admin/
│   ├── package.json
│   ├── tsconfig.json
│   ├── tsconfig.node.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── postcss.config.mjs
│   ├── index.html
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── router.ts
│   │   ├── api/
│   │   │   ├── client.ts          # fetch wrapper + CSRF
│   │   │   ├── articles.ts
│   │   │   ├── ads.ts
│   │   │   ├── settings.ts
│   │   │   ├── auth.ts
│   │   │   ├── images.ts
│   │   │   └── ai.ts              # SSE consumer
│   │   ├── stores/
│   │   │   ├── auth.ts
│   │   │   ├── articles.ts
│   │   │   ├── ads.ts
│   │   │   └── settings.ts
│   │   ├── views/
│   │   │   ├── Login.vue
│   │   │   ├── Articulos.vue
│   │   │   ├── ArticuloEditor.vue
│   │   │   ├── Publicidad.vue
│   │   │   ├── PublicidadEditor.vue
│   │   │   └── Ajustes.vue
│   │   ├── components/
│   │   │   ├── Toast.vue
│   │   │   ├── Modal.vue
│   │   │   ├── Spinner.vue
│   │   │   ├── ImageUpload.vue
│   │   │   ├── ToneSelect.vue
│   │   │   ├── StreamingTextarea.vue
│   │   │   └── NavBar.vue
│   │   ├── composables/
│   │   │   ├── useToast.ts
│   │   │   └── useSSE.ts
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── styles/
│   │       └── main.css
│   └── dist/                      # OUTPUT (gitignore)
├── data/                          # gitignore (datos en runtime)
│   ├── admin.json
│   ├── settings.json
│   ├── ads.json
│   └── ads/
├── static-uploads/                # gitignore
│   └── images/YYYY/MM/{hash}-{w}.{avif,webp}
├── scripts/
│   ├── dev.sh                     # arranca todo en dev
│   ├── build.sh                   # build de producción
│   └── seed.sh                    # datos demo
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── CONTRIBUTING.md
└── PLAN.md                        # este archivo
```

### 5.1 `.gitignore` (mínimo obligatorio)

```
# Build outputs
site/public/
web/admin/dist/
diegonoticias              # binario compilado

# Runtime data
data/
static-uploads/
*.log

# Node/Vite
node_modules/

# Editor / OS
.DS_Store
.idea/
.vscode/
*.swp

# Secrets
.env
*.local
```

---

## 6. Dependencias y versiones

### 6.1 Go (módulo)

```
go 1.23+    # usar último estable disponible
```

| Dependencia | Propósito |
|---|---|
| `github.com/alexedwards/argon2id` | Hash de contraseña |
| `github.com/davidbyttow/govips/v2` | Procesamiento de imágenes (libvips) |
| `github.com/gosimple/slug` | Slugs URL-safe con transliteración |
| `github.com/google/uuid` | IDs de banners |
| `github.com/adrg/frontmatter` | Parser de frontmatter YAML |
| `golang.org/x/time/rate` | Rate limiting de generaciones IA |
| `github.com/joho/godotenv` | Cargar `.env` en dev |

**Standard library**: `net/http`, `embed`, `encoding/json`, `os/exec`, `log/slog`, `sync`, `crypto/rand`, `crypto/subtle`.

> **Nota**: nada de frameworks HTTP (Gin, Echo, Fiber). Usa `net/http` y mux nativo (`http.ServeMux` con patterns Go 1.22+).

### 6.2 Sistema (instalado en VPS y/o dev)

| Binario | Versión mínima | Notas |
|---|---|---|
| `hugo` (extended) | 0.140+ | Imprescindible para Tailwind v4 vía pipes |
| `pagefind` | 1.x | Distribuir como binario |
| `tailwindcss` (standalone CLI) | v4.x | NO el paquete npm; el binario suelto |
| `libvips` | 8.14+ | Dependencia de `govips` |
| `nginx` | 1.24+ | Ya instalado |
| `goaccess` | 1.9+ | Para reportes |
| `certbot` | latest | Renovación TLS |

Documenta en README cómo instalar en Debian/Ubuntu (apt) y cómo en macOS (brew) para dev.

### 6.3 Node (solo en dev/CI, no en prod)

`web/admin/package.json`:

```json
{
  "name": "diegonoticias-admin",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "preview": "vite preview",
    "lint": "eslint --ext .ts,.vue src",
    "format": "prettier --write src"
  },
  "dependencies": {
    "vue": "^3.5.0",
    "vue-router": "^4.4.0",
    "pinia": "^2.2.0",
    "partial-json": "^0.1.7"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.1.0",
    "typescript": "^5.6.0",
    "vue-tsc": "^2.1.0",
    "vite": "^5.4.0",
    "tailwindcss": "^4.0.0",
    "@tailwindcss/vite": "^4.0.0",
    "autoprefixer": "^10.4.0",
    "eslint": "^9.0.0",
    "@typescript-eslint/parser": "^8.0.0",
    "prettier": "^3.3.0"
  }
}
```

> **Importante**: Tailwind v4 en el admin SPA va vía Vite plugin. En el sitio Hugo va vía Hugo Pipes con el binario standalone. Son **dos integraciones distintas** del mismo Tailwind.

---

## 7. Modelo de datos

### 7.1 Frontmatter de un artículo

```yaml
---
title: "Título del artículo"
slug: "titulo-del-articulo"
date: 2026-05-06T14:30:00-05:00
description: "Meta description, 140-160 caracteres."
tone: "informativo"
category: "tecnologia"
image: "/images/2026/05/abc123def456"   # base path; renderer compone -640.webp etc.
imageAlt: "Descripción accesible de la imagen"
draft: false
wordCount: 198
---

Cuerpo del artículo en Markdown simple…

Otro párrafo con *énfasis* o **negrita**…
```

**Validaciones**:
- `title`: 1–120 chars
- `slug`: kebab-case, ascii, máx 80 chars, único
- `date`: ISO 8601 con offset, generado por el server al publicar
- `description`: 100–170 chars (rango con tolerancia)
- `tone`: uno de la lista cerrada
- `category`: una palabra, lowercase, ascii, sin espacios
- `image`: opcional, formato `/images/YYYY/MM/{hash}` sin extensión ni tamaño
- `imageAlt`: requerido si hay `image`
- `wordCount`: entero, calculado al guardar (no input del usuario)

### 7.2 `data/admin.json`

```json
{
  "username": "diego",
  "passwordHash": "$argon2id$v=19$m=...",
  "createdAt": "2026-05-06T14:00:00-05:00",
  "updatedAt": "2026-05-06T14:00:00-05:00"
}
```

Generado por el comando CLI `./diegonoticias setup-admin`.

### 7.3 `data/settings.json`

```json
{
  "siteName": "Diego Noticias",
  "siteDescription": "Noticias breves y al grano.",
  "siteUrl": "https://diegonoticias.com",
  "defaultOgImage": "/og-default.jpg",
  "twitterHandle": "",
  "adsense": {
    "enabled": false,
    "clientId": "",
    "slot1Id": "",
    "slot1Enabled": false,
    "slot2Id": "",
    "slot2Enabled": false
  }
}
```

### 7.4 `data/ads.json`

```json
{
  "banners": [
    {
      "id": "uuid-v4",
      "title": "Mi cliente A",
      "imagePath": "data/ads/uuid/banner",
      "linkUrl": "https://cliente-a.com",
      "active": true,
      "slot": 1,
      "createdAt": "...",
      "updatedAt": "..."
    }
  ]
}
```

**Invariantes** (validar siempre antes de persistir):
- Total banners ≤ 7
- Activos ≤ 2
- A lo sumo 1 banner activo por slot
- `slot ∈ {1, 2}`
- `linkUrl` parseable como URL absoluta http(s)

### 7.5 Datos derivados (escritos por Go en cada build, leídos por Hugo)

`site/data/site.toml`:
```toml
siteName = "Diego Noticias"
siteDescription = "..."
twitterHandle = ""
defaultOgImage = "/og-default.jpg"
[adsense]
enabled = false
clientId = ""
# ...
```

`site/data/ads.toml`:
```toml
[[banners]]
id = "..."
title = "..."
imagePath = "/ads/uuid/banner"
linkUrl = "..."
slot = 1
```

Solo se incluyen los **activos** en `ads.toml`. Los inactivos viven en `data/ads.json` pero no se exportan.

---

## 8. Sitio público (Hugo)

### 8.1 `hugo.toml`

```toml
baseURL = "https://diegonoticias.com/"
languageCode = "es-CO"
defaultContentLanguage = "es"
title = "Diego Noticias"
enableRobotsTXT = true
enableGitInfo = false
paginate = 12
summaryLength = 30

[permalinks]
articulos = "/:slug/"

[markup.goldmark.renderer]
unsafe = true

[markup.goldmark.parser]
autoHeadingID = true

[imaging]
quality = 82
resampleFilter = "Lanczos"

[outputs]
home = ["HTML", "RSS"]
section = ["HTML", "RSS"]

[minify]
minifyOutput = true

[sitemap]
changefreq = "daily"
filename = "sitemap.xml"
priority = 0.5

[params]
ogDefault = "/og-default.jpg"
```

### 8.2 Layouts (responsabilidades)

- **`_default/baseof.html`**: HTML5 esqueleto con `<head>` (incluye partials de meta, og, jsonld), body con `<header>`, `<main>{{ block "main" . }}{{ end }}</main>` y `<footer>`.
- **`index.html`**: home. Hero con el artículo más reciente (tarjeta grande), grid de tarjetas para los siguientes, slot de publicidad intercalado entre el 5to y 6to artículo, paginación al final.
- **`_default/list.html`**: fallback para listados (categoría, tag — aunque no usemos categorías como página, Hugo las genera por defecto).
- **`articulos/single.html`**: artículo individual (header, imagen hero, título, meta, body Markdown renderizado, slot de publicidad después del 2do párrafo, sugeridos al final).
- **`partials/article-card.html`** y **`article-card-hero.html`**: tarjetas reutilizables.
- **`partials/ad-slot.html`**: dado un slot 1 o 2, renderiza AdSense si habilitado, o banner propio activo, o nada.
- **`partials/head/meta.html`**: title tag, description, canonical, robots, viewport, charset, theme-color.
- **`partials/head/og.html`**: Open Graph + Twitter Cards.
- **`partials/head/jsonld.html`**: `Article` + `BreadcrumbList` + `WebSite` (con `SearchAction`).
- **`partials/header.html`**: barra superior con logo (texto), botón buscar (abre overlay).
- **`partials/footer.html`**: créditos, link RSS, link sitemap.
- **`partials/search.html`**: overlay con input y resultados de Pagefind.
- **`partials/pagination.html`**: paginación de Hugo personalizada.
- **`404.html`**: página 404 con link al home.
- **`robots.txt`**: permite todo, referencia sitemap.

### 8.3 Tailwind y CSS

`assets/css/main.css`:
```css
@import "tailwindcss";

/* Variables de marca */
:root {
  --color-bg: #FAF7F2;
  --color-fg: #1A1A1A;
  --color-muted: #6B6B6B;
  --color-accent: #C8553D;
}

/* Tipografía editorial */
@layer base {
  html { font-family: 'Inter', system-ui, sans-serif; }
  .font-serif { font-family: 'Source Serif 4', Georgia, serif; }
}
```

> **Nota tipografía**: usar fuentes self-hosted (descargar `.woff2` y meter en `static/fonts/`). NO usar Google Fonts CDN — empeora performance y privacidad.

Hugo Pipes en `baseof.html`:
```go-html-template
{{ $css := resources.Get "css/main.css" | css.TailwindCSS | minify | fingerprint }}
<link rel="stylesheet" href="{{ $css.RelPermalink }}" integrity="{{ $css.Data.Integrity }}">
```

### 8.4 Búsqueda con Pagefind

- En cada build, después de `hugo`, correr `pagefind --site site/public`.
- Cliente JS de Pagefind se carga **solo cuando el usuario abre el overlay de búsqueda**, no en cada página.
- Búsqueda con stemming en español (Pagefind lo soporta automáticamente con `lang="es"` en `<html>`).

### 8.5 Mobile bottom nav

Solo en viewport `< md`. Posición fixed bottom, fondo translúcido con backdrop-blur. Iconos SVG inline (sin librería). 3 ítems.

### 8.6 Comportamiento esperado del sitio

- **First Contentful Paint** < 1.0s en 4G.
- **Largest Contentful Paint** < 1.5s en 4G.
- **Total Blocking Time** < 100ms.
- **Cumulative Layout Shift** < 0.05.
- **JS bundle inicial**: < 5 KB (solo Alpine inline core mínimo + un poco propio).
- **Pagefind**: lazy, solo al abrir buscador (~50 KB).
- **Lighthouse mobile**: 95+ en las 4 categorías.

---

## 9. Admin SPA (Vue 3 + TypeScript)

### 9.1 Rutas

```
/login                             → Login.vue
/articulos                         → Articulos.vue (lista)
/articulos/nuevo                   → ArticuloEditor.vue (modo create)
/articulos/:slug/editar            → ArticuloEditor.vue (modo edit)
/publicidad                        → Publicidad.vue (lista)
/publicidad/nueva                  → PublicidadEditor.vue (create)
/publicidad/:id/editar             → PublicidadEditor.vue (edit)
/ajustes                           → Ajustes.vue
```

Todas las rutas excepto `/login` requieren sesión (guard en router que verifica `auth.isAuthenticated`).

### 9.2 Stores Pinia

- `auth`: usuario actual, estado de sesión, login, logout.
- `articles`: cache de la lista paginada, `current` para edit.
- `ads`: lista de banners.
- `settings`: ajustes del sitio.

### 9.3 Pantalla `ArticuloEditor.vue`

Esta es la pantalla más compleja. Especificación:

**Modo crear**:
1. Campos:
   - **Título** (texto, opcional, máx 120 chars)
   - **Texto crudo** (textarea, requerido si no hay texto editado, mín 5 palabras, máx 300)
   - **Tono** (dropdown, requerido, default "informativo")
   - **Imagen** (upload opcional, JPG/PNG/WebP/AVIF/HEIC, máx 8 MB)
   - **Alt de imagen** (auto-generado por la IA, editable)
2. Botón **"Generar"**: deshabilitado si no hay texto crudo o tono. Llama a `POST /admin/api/articulos/generar` (SSE) con `{ rawText, tone, titleHint?, hasImage }`.
3. Mientras streaming:
   - Mostrar "Generando…" con spinner.
   - Aparecen progresivamente: título → meta description → categoría → alt → cuerpo (este último carácter por carácter).
   - Botón **"Detener"** cancela el stream (cierra EventSource y POSTea cancel).
4. Tras stream completo:
   - Todos los campos quedan **editables**.
   - Aparecen botones **"Regenerar"** (vuelve a llamar con el mismo input) y **"Publicar"**.
5. Botón **"Publicar"**: valida (título no vacío, body 170–230 palabras, etc.), llama a `POST /admin/api/articulos`. Si éxito, muestra toast con link público y redirige a `/articulos`.

**Modo editar**:
- Carga el artículo desde `GET /admin/api/articulos/:slug`.
- Mismos campos prellenados.
- Sin botón "Generar" inicialmente. Botón **"Re-generar"** que abre un modal preguntando si desea sobrescribir todo con una nueva generación (peligro: pierde edits).
- Botón **"Guardar cambios"** llama a `PUT /admin/api/articulos/:slug`.
- Botón **"Borrar"** con confirmación.

### 9.4 Componente `StreamingTextarea.vue`

Wrapper de un `<textarea>` que escucha un stream de chunks y los va concatenando. Después del fin del stream, queda como textarea normal editable. Props:
- `streaming: boolean`
- `value: string` (v-model)
- `onAppend: (chunk: string) => void`

### 9.5 Composable `useSSE.ts`

```ts
export function useSSE<T>(url: string, body: unknown) {
  // POST con fetch + ReadableStream para SSE bi-direccional
  // (EventSource solo soporta GET)
  // emite chunks, error, done
}
```

Usar `fetch` con `ReadableStream` porque EventSource es solo GET y necesitamos POST con body.

### 9.6 Parser de JSON parcial

Importar `partial-json`:
```ts
import { parse, ALL } from 'partial-json';

const partial = parse(buffer, ALL);
// partial = { title?, body?, metaDescription?, category?, imageAlt? }
```

Cada vez que llega un chunk del SSE, concatenar al buffer y re-parsear. Actualizar UI.

### 9.7 Formato de estilos del admin

- Layout simple: sidebar izquierda fija (desktop) / topbar (mobile) con navegación.
- Paleta neutra (grises + acento).
- Botones primarios: acento. Secundarios: outline. Destructivos: rojo.
- Toasts para feedback (éxito/error/warning).
- Modales para confirmaciones destructivas.

---

## 10. Backend Go

### 10.1 `main.go` (estructura)

```go
package main

import (
    "context"
    "embed"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

//go:embed all:web/admin/dist
var adminFS embed.FS

func main() {
    cmd := ""
    if len(os.Args) > 1 { cmd = os.Args[1] }

    switch cmd {
    case "setup-admin":
        runSetupAdmin()
    case "":
        runServer()
    default:
        slog.Error("unknown command", "cmd", cmd)
        os.Exit(2)
    }
}
```

### 10.2 Router (Go 1.22+ patterns)

```go
mux := http.NewServeMux()

// Sitio público (filesystem)
mux.Handle("GET /", publicHandler())

// Imágenes y ads (filesystem)
mux.Handle("GET /images/", http.StripPrefix("/images/", imageHandler()))
mux.Handle("GET /ads/", http.StripPrefix("/ads/", adImageHandler()))

// Admin SPA (embed)
mux.Handle("GET /admin/", adminSPAHandler())

// Admin API
mux.Handle("POST /admin/api/login", h.login)
mux.Handle("POST /admin/api/logout", authRequired(h.logout))
mux.Handle("GET /admin/api/me", authRequired(h.me))
mux.Handle("GET /admin/api/articulos", authRequired(h.listArticles))
mux.Handle("GET /admin/api/articulos/{slug}", authRequired(h.getArticle))
mux.Handle("POST /admin/api/articulos/generar", authRequired(rateLimited(h.generateArticle)))
mux.Handle("POST /admin/api/articulos", authRequired(h.createArticle))
mux.Handle("PUT /admin/api/articulos/{slug}", authRequired(h.updateArticle))
mux.Handle("DELETE /admin/api/articulos/{slug}", authRequired(h.deleteArticle))
mux.Handle("POST /admin/api/imagenes", authRequired(h.uploadImage))
mux.Handle("GET /admin/api/publicidad", authRequired(h.listAds))
mux.Handle("POST /admin/api/publicidad", authRequired(h.createAd))
mux.Handle("PUT /admin/api/publicidad/{id}", authRequired(h.updateAd))
mux.Handle("DELETE /admin/api/publicidad/{id}", authRequired(h.deleteAd))
mux.Handle("POST /admin/api/publicidad/{id}/activar", authRequired(h.activateAd))
mux.Handle("POST /admin/api/publicidad/{id}/desactivar", authRequired(h.deactivateAd))
mux.Handle("GET /admin/api/ajustes", authRequired(h.getSettings))
mux.Handle("PUT /admin/api/ajustes", authRequired(h.updateSettings))
mux.Handle("GET /admin/api/build/estado", authRequired(h.buildStatus))
mux.Handle("POST /admin/api/build/regenerar", authRequired(h.rebuildAll))

server := &http.Server{
    Addr:              "127.0.0.1:8080",
    Handler:           securityHeaders(loggingMiddleware(mux)),
    ReadHeaderTimeout: 10 * time.Second,
    ReadTimeout:       30 * time.Second,
    WriteTimeout:      0, // 0 para SSE; controlamos timeout por handler
    IdleTimeout:       120 * time.Second,
}
```

### 10.3 Middlewares

- **`authRequired`**: lee cookie de sesión, valida, inyecta usuario en `r.Context()`. Si no autenticado, devuelve 401 JSON.
- **`csrfRequired`** (en mutaciones): valida header `X-CSRF-Token` contra el de la sesión.
- **`rateLimited`**: para `generar`, usa `golang.org/x/time/rate` con bucket diario. Cap 100/día.
- **`securityHeaders`**: añade headers en cada respuesta.
- **`loggingMiddleware`**: log estructurado por petición.

### 10.4 Builder (orquesta hugo + pagefind)

```go
// internal/builder/builder.go
type Builder struct {
    siteDir   string
    mu        sync.Mutex
    pending   atomic.Bool
    lastBuild atomic.Pointer[BuildResult]
}

type BuildResult struct {
    Status    string // "ok" | "error"
    Error     string
    StartedAt time.Time
    EndedAt   time.Time
    Duration  time.Duration
}

func (b *Builder) Trigger() {
    if !b.pending.CompareAndSwap(false, true) { return }
    go func() {
        b.mu.Lock()
        defer b.mu.Unlock()
        defer b.pending.Store(false)
        b.run()
    }()
}

func (b *Builder) run() {
    // 1. Exportar site/data/site.toml y site/data/ads.toml desde data/*.json
    // 2. exec hugo --minify
    // 3. exec pagefind --site public
    // 4. Actualizar b.lastBuild
}
```

**Comportamiento**:
- Si llega un trigger durante un build en curso, se hace **un solo build adicional** después (coalescing). No se acumulan múltiples builds.
- Build sale a 5–10 segundos típicos. Pagefind ~1–2s.
- Si falla, log error y mantener el `site/public/` previo intacto (Hugo escribe en sitio).

### 10.5 Articles store

```go
// internal/articles/store.go
type Store struct {
    contentDir string  // site/content/articulos
}

func (s *Store) List(query string, page, pageSize int) ([]Article, int, error)
func (s *Store) Get(slug string) (*Article, error)
func (s *Store) Create(a *Article) error           // valida slug único, escribe atómico
func (s *Store) Update(slug string, a *Article) error
func (s *Store) Delete(slug string) error
```

**Listado y búsqueda en admin**: leer todos los `.md`, parsear frontmatter en memoria. Para 100 artículos esto es <50ms — basta. Si llegamos a 1000+, cachear con `fsnotify`.

### 10.6 Slug

```go
// internal/articles/slug.go
import "github.com/gosimple/slug"

func Generate(title string, existing func(string) bool) string {
    base := slug.MakeLang(title, "es")
    if len(base) > 80 { base = base[:80] }
    s := base
    for i := 2; existing(s); i++ {
        s = fmt.Sprintf("%s-%d", base, i)
    }
    return s
}
```

### 10.7 SSE para streaming IA

```go
func (h *Handler) generateArticle(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no")  // crítico con nginx

    flusher := w.(http.Flusher)
    ctx := r.Context()

    stream, err := h.ai.Generate(ctx, params)
    if err != nil { writeSSEError(w, flusher, err); return }
    defer stream.Close()

    for chunk := range stream.Chunks() {
        writeSSEData(w, flusher, chunk)
    }
    writeSSEEvent(w, flusher, "done", nil)
}
```

**nginx** debe configurarse con `proxy_buffering off;` para que el streaming pase bien.

---

## 11. Integración con Groq

### 11.1 Cliente

`internal/ai/client.go`: HTTP client a `https://api.groq.com/openai/v1/chat/completions` (compatible OpenAI). API key en env `GROQ_API_KEY`.

```go
type GenerateParams struct {
    RawText   string
    Tone      string
    TitleHint string  // opcional
    HasImage  bool
}

type GenerateResult struct {
    Title           string `json:"title"`
    Body            string `json:"body"`
    MetaDescription string `json:"metaDescription"`
    Category        string `json:"category"`
    ImageAlt        string `json:"imageAlt"`
}
```

### 11.2 Tonos (constantes)

```go
var Tones = []Tone{
    {ID: "informativo",     Label: "Informativo",     Description: "directo, factual, neutral, estilo agencia de noticias"},
    {ID: "profesional",     Label: "Profesional",     Description: "formal pero accesible, tono ejecutivo"},
    {ID: "institucional",   Label: "Institucional",   Description: "voz oficial, comunicado de prensa formal"},
    {ID: "academico",       Label: "Académico",       Description: "con datos, contexto histórico, lenguaje preciso"},
    {ID: "cronica",         Label: "Crónica",         Description: "narrativo y descriptivo, casi novelado"},
    {ID: "editorial",       Label: "Editorial",       Description: "opinión razonada con argumentos"},
    {ID: "conversacional",  Label: "Conversacional",  Description: "cercano y casual, como hablarle a un amigo"},
    {ID: "pedagogico",      Label: "Pedagógico",      Description: "explica conceptos como a alguien que no sabe del tema"},
    {ID: "dramatico",       Label: "Dramático",       Description: "tensión, urgencia, alto impacto emocional"},
    {ID: "sensacionalista", Label: "Sensacionalista", Description: "lenguaje fuerte y exagerado para enganchar, pero sin mentir ni inventar hechos"},
}
```

### 11.3 Prompt

`internal/ai/prompt.go`:

```
Eres un editor profesional de noticias en español de Colombia. Recibirás un texto crudo del usuario y debes producir un artículo completo.

REGLAS DURAS:
- Devuelve SOLO un objeto JSON válido, sin texto antes ni después, sin comillas tipográficas.
- El JSON debe tener exactamente estos campos: title, body, metaDescription, category, imageAlt.
- title: máximo 12 palabras, en español, sin punto final, sin emojis, sin comillas.
- body: entre 170 y 230 palabras (apunta a 200). Markdown simple: solo párrafos separados por línea en blanco. Puedes usar *cursiva* y **negrita** ocasionalmente. NO uses headings, listas, blockquotes, links ni HTML.
- metaDescription: entre 140 y 160 caracteres, una sola oración o dos cortas, sin comillas.
- category: una sola palabra en minúsculas, sin tildes ni espacios. Debe describir el tema.
- imageAlt: descripción accesible de la imagen del artículo en 8 a 14 palabras. Si el usuario indica que NO hay imagen, devuelve cadena vacía.

CONTEXTO:
- Tono solicitado: {{ .ToneID }} ({{ .ToneDescription }})
- ¿Hay imagen subida?: {{ if .HasImage }}sí{{ else }}no{{ end }}
{{- if .TitleHint }}
- Sugerencia de título del usuario (úsala como base si tiene sentido): "{{ .TitleHint }}"
{{- end }}

TEXTO CRUDO DEL USUARIO:
"""
{{ .RawText }}
"""

Genera ahora el JSON. Mantén fidelidad al sentido del texto crudo, no inventes hechos. Si el texto es ambiguo, sé conservador y factual.
```

### 11.4 Llamada a Groq

```json
{
  "model": "llama-3.3-70b-versatile",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ],
  "response_format": {"type": "json_object"},
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": true
}
```

### 11.5 Validación de salida

Tras consumir el stream completo, validar el JSON resultante:
- Todos los campos presentes y no vacíos (excepto `imageAlt` que puede ser vacío si no hay imagen).
- `body` palabras: 170 ≤ count ≤ 230.
- `metaDescription` chars: 100 ≤ len ≤ 170 (margen amplio porque LLM falla en contar).
- `title` palabras: ≤ 12.
- `category` matchea regex `^[a-z]+$`.

Si alguna validación falla: aún devuelve el resultado al admin (mejor algo editable que un error). El usuario puede corregir antes de publicar. Lo único que sí bloquea publicación es body fuera de rango duro o título vacío.

### 11.6 Manejo de errores

- Si el HTTP a Groq falla o devuelve no-2xx: SSE emite event `error` con mensaje genérico en español, frontend muestra error y permite escribir manualmente.
- Si el JSON final es inválido (no parsea): mismo trato.
- Si la cuota diaria está agotada (rate limit interno): 429 antes de llamar.

---

## 12. Pipeline de imágenes

### 12.1 Configuración

```go
var ImageConfig = struct{
    MaxUploadBytes int64
    Sizes          []int   // anchos
    AvifQuality    int
    WebpQuality    int
    AdAvifQuality  int     // más agresivo para banners
    AdWebpQuality  int
}{
    MaxUploadBytes: 8 * 1024 * 1024,  // 8 MB
    Sizes:          []int{320, 640, 1024, 1600},
    AvifQuality:    55,
    WebpQuality:    72,
    AdAvifQuality:  45,
    AdWebpQuality:  60,
}
```

### 12.2 Flujo

```go
func ProcessArticleImage(input []byte) (basePath string, err error) {
    img, err := vips.NewImageFromBuffer(input)
    if err != nil { return "", fmt.Errorf("decode: %w", err) }
    defer img.Close()

    img.RemoveMetadata()
    if img.HasAlpha() { /* manejar adecuadamente */ }

    hash := randomHex(8)
    now := time.Now()
    base := fmt.Sprintf("images/%04d/%02d/%s", now.Year(), now.Month(), hash)
    diskBase := filepath.Join(uploadsRoot, base)
    os.MkdirAll(filepath.Dir(diskBase), 0755)

    for _, w := range ImageConfig.Sizes {
        if w > img.Width() { continue }  // no escalar arriba
        scale := float64(w) / float64(img.Width())
        thumb, _ := img.Copy()
        thumb.Resize(scale, vips.KernelLanczos3)

        avifBytes, _, err := thumb.ExportAvif(&vips.AvifExportParams{Quality: ImageConfig.AvifQuality})
        if err != nil { return "", err }
        atomicWriteFile(diskBase+fmt.Sprintf("-%d.avif", w), avifBytes)

        webpBytes, _, err := thumb.ExportWebp(&vips.WebpExportParams{Quality: ImageConfig.WebpQuality})
        if err != nil { return "", err }
        atomicWriteFile(diskBase+fmt.Sprintf("-%d.webp", w), webpBytes)

        thumb.Close()
    }

    return "/" + base, nil
}
```

### 12.3 Render en plantilla Hugo

`partials/picture.html`:
```go-html-template
{{- $base := .base -}}
{{- $alt := .alt -}}
{{- $sizes := .sizes | default "(max-width: 768px) 100vw, 800px" -}}
<picture>
  <source type="image/avif" srcset="{{ $base }}-320.avif 320w, {{ $base }}-640.avif 640w, {{ $base }}-1024.avif 1024w, {{ $base }}-1600.avif 1600w" sizes="{{ $sizes }}">
  <source type="image/webp" srcset="{{ $base }}-320.webp 320w, {{ $base }}-640.webp 640w, {{ $base }}-1024.webp 1024w, {{ $base }}-1600.webp 1600w" sizes="{{ $sizes }}">
  <img src="{{ $base }}-1024.webp" alt="{{ $alt }}" loading="{{ .loading | default "lazy" }}" decoding="async" fetchpriority="{{ .priority | default "auto" }}">
</picture>
```

### 12.4 Hero (above-the-fold)

Para la primera imagen del home y la imagen del artículo single, agregar:
```html
<link rel="preload" as="image" href="{{ $base }}-1024.avif" type="image/avif" imagesrcset="..." imagesizes="...">
```
Y `loading="eager"` + `fetchpriority="high"` en `<img>`.

---

## 13. Sistema de publicidad

### 13.1 Modelo en runtime

```go
type Banner struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    ImagePath string    `json:"imagePath"`   // base sin extensión
    LinkURL   string    `json:"linkUrl"`
    Active    bool      `json:"active"`
    Slot      int       `json:"slot"`        // 1 o 2
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

### 13.2 Validaciones (antes de persistir)

```go
func (s *Store) Validate(b Banner, allBanners []Banner) error {
    // Slot ∈ {1,2}
    // LinkURL parseable http/https
    // Title 1..80 chars
    // Si Active==true:
    //   - count(active) including this <= 2
    //   - no otro banner activo en mismo slot
    // Total <= 7
}
```

### 13.3 Slot rendering (Hugo)

`partials/ad-slot.html`:
```go-html-template
{{- $slot := .slot -}}
{{- $settings := site.Data.site -}}

{{- if and $settings.adsense.enabled (eq $slot 1) $settings.adsense.slot1Enabled -}}
  <div class="ad-slot ad-adsense">
    <!-- AdSense slot 1 -->
    <ins class="adsbygoogle" style="display:block"
         data-ad-client="{{ $settings.adsense.clientId }}"
         data-ad-slot="{{ $settings.adsense.slot1Id }}"
         data-ad-format="auto" data-full-width-responsive="true"></ins>
    <script>(adsbygoogle = window.adsbygoogle || []).push({});</script>
  </div>
{{- else if and $settings.adsense.enabled (eq $slot 2) $settings.adsense.slot2Enabled -}}
  <!-- similar slot 2 -->
{{- else -}}
  {{- range site.Data.ads.banners -}}
    {{- if eq .slot $slot -}}
      <a class="ad-slot ad-banner" href="{{ .linkUrl }}" rel="sponsored noopener" target="_blank">
        {{ partial "picture.html" (dict "base" .imagePath "alt" .title "loading" "lazy") }}
      </a>
    {{- end -}}
  {{- end -}}
{{- end -}}
```

### 13.4 Carga de AdSense

Solo si `adsense.enabled == true`, incluir el script en `<head>` con `async`:
```html
<script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=..." crossorigin="anonymous"></script>
```

### 13.5 CSP

Cuando AdSense esté habilitado, la CSP debe permitir `googlesyndication.com` y dominios relacionados. Por ahora, mientras esté deshabilitado, CSP estricta.

---

## 14. Autenticación y seguridad

### 14.1 Flujo

1. `POST /admin/api/login` con `{username, password}` (JSON).
2. Servidor compara contra `data/admin.json` con `argon2id.ComparePasswordAndHash`.
3. Si OK, genera token aleatorio (32 bytes, base64), lo guarda en `sessions` (mapa en memoria) con expiración 7 días.
4. Setea cookie `dn_session=<token>; HttpOnly; Secure; SameSite=Strict; Path=/admin; Max-Age=...`.
5. Genera token CSRF (otro 32 bytes), lo guarda asociado a la sesión.
6. Devuelve `{username, csrfToken}` en el body.
7. El SPA guarda `csrfToken` en memoria (NO localStorage) y lo manda en header `X-CSRF-Token` en cada mutación.

### 14.2 Logout

`POST /admin/api/logout`: borra sesión y cookie.

### 14.3 Comandos CLI

`./diegonoticias setup-admin`:
- Si `data/admin.json` ya existe, pregunta `[s/N]` si sobrescribir.
- Pide username (default `diego`) y password (oculto, doble entrada).
- Genera Argon2id hash con `argon2id.DefaultParams`.
- Escribe `data/admin.json` con permisos `0600`.

### 14.4 Headers de seguridad

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: interest-cohort=()
X-Frame-Options: DENY
Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; font-src 'self'; connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'
```

`unsafe-inline` solo para los `<script>` tipo JSON-LD y Alpine init que son necesarios. Considerar nonce-based CSP a futuro si se complica.

### 14.5 Rate limiting

- IP-based en login: 5 intentos / 5 minutos. Tras eso, 429.
- Generación IA: 100 por día (window deslizante).

---

## 15. SEO

### 15.1 Implementación obligatoria

| Item | Implementado en |
|---|---|
| `<title>` y `<meta description>` por página | `partials/head/meta.html` |
| Canonical URL | `partials/head/meta.html` |
| Open Graph completo | `partials/head/og.html` |
| Twitter Cards | `partials/head/og.html` |
| JSON-LD `Article` por artículo | `partials/head/jsonld.html` |
| JSON-LD `WebSite` con `SearchAction` | `partials/head/jsonld.html` (solo home) |
| JSON-LD `BreadcrumbList` | `partials/head/jsonld.html` |
| Sitemap.xml | Hugo nativo |
| robots.txt | `layouts/robots.txt` |
| RSS feed | Hugo nativo |
| Favicon multi-tamaño + manifest | `static/` |
| Imágenes responsivas | pipeline + `picture.html` |
| Lazy loading | `loading="lazy"` |
| Preload de hero | `<link rel="preload">` |
| `lang="es-CO"` en `<html>` | `baseof.html` |
| Encabezados semánticos (h1, h2…) | en plantillas |
| Alt en imágenes | required en frontmatter |
| URL kebab-case sin acentos | slug.go |

### 15.2 Verificación

Tras la fase 9, ejecutar:
- Lighthouse mobile y desktop, exigir 95+ en SEO y Performance.
- `curl -s https://diegonoticias.com/sitemap.xml | head` para validar.
- `curl -s https://diegonoticias.com/index.xml` (RSS).
- Validar JSON-LD en https://search.google.com/test/rich-results.
- Validar Open Graph en https://www.opengraph.xyz/.

---

## 16. Despliegue

### 16.1 nginx (sitio)

`/etc/nginx/sites-available/diegonoticias.com`:

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name diegonoticias.com www.diegonoticias.com;
    location /.well-known/acme-challenge/ { root /var/www/certbot; }
    location / { return 301 https://diegonoticias.com$request_uri; }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name www.diegonoticias.com;
    ssl_certificate /etc/letsencrypt/live/diegonoticias.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/diegonoticias.com/privkey.pem;
    return 301 https://diegonoticias.com$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name diegonoticias.com;

    ssl_certificate /etc/letsencrypt/live/diegonoticias.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/diegonoticias.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    gzip on;
    gzip_vary on;
    gzip_min_length 256;
    gzip_types text/plain text/css application/javascript application/json application/xml image/svg+xml;

    client_max_body_size 10M;
    access_log /var/log/nginx/diegonoticias.access.log combined;
    error_log /var/log/nginx/diegonoticias.error.log warn;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE support
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
    }
}
```

### 16.2 systemd

`/etc/systemd/system/diegonoticias.service`:

```ini
[Unit]
Description=Diego Noticias
After=network.target

[Service]
Type=simple
User=diegonoticias
Group=diegonoticias
WorkingDirectory=/opt/diegonoticias
ExecStart=/opt/diegonoticias/diegonoticias
EnvironmentFile=/opt/diegonoticias/.env
Restart=on-failure
RestartSec=5

NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/opt/diegonoticias/data /opt/diegonoticias/static-uploads /opt/diegonoticias/site /opt/diegonoticias/logs
ProtectHome=true
PrivateTmp=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true

LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### 16.3 Estructura en VPS

```
/opt/diegonoticias/
├── diegonoticias              # binario
├── .env                       # secretos (chmod 600)
├── data/                      # admin.json, settings.json, ads.json, ads/
├── static-uploads/            # imágenes de artículos
├── site/                      # proyecto Hugo (con content/, layouts/, etc.)
└── logs/
```

### 16.4 Script de despliegue (`scripts/deploy.sh`)

```bash
#!/usr/bin/env bash
set -euo pipefail

# 1. Build local
cd web/admin && npm ci && npm run build && cd ../..
GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o build/diegonoticias ./cmd/server
tar czf build/site-assets.tar.gz site/layouts site/assets site/static site/archetypes site/hugo.toml

# 2. Subir
scp build/diegonoticias build/site-assets.tar.gz vps:/opt/diegonoticias/
ssh vps "cd /opt/diegonoticias && tar xzf site-assets.tar.gz && systemctl restart diegonoticias"

# 3. Verificar
ssh vps "systemctl status diegonoticias --no-pager"
```

> El primer despliegue requiere setup adicional (instalar hugo, pagefind, tailwind CLI, libvips, certbot, crear usuario de sistema, etc.). Documentar en README.

### 16.5 GoAccess

Configurar reporte HTML diario:
```
goaccess /var/log/nginx/diegonoticias.access.log -o /opt/diegonoticias/data/stats.html --log-format=COMBINED --real-time-html
```

Servir `stats.html` desde el admin (ruta protegida) o dejarlo en disco para que el admin haga `tail` por SSH.

---

## 17. Variables de entorno

`.env.example` (committeado al repo):

```
# Servidor
DN_LISTEN=127.0.0.1:8080
DN_LOG_LEVEL=info               # debug | info | warn | error
DN_ENV=production               # production | development

# Rutas (todas relativas al binario o absolutas)
DN_DATA_DIR=./data
DN_UPLOADS_DIR=./static-uploads
DN_SITE_DIR=./site

# Binarios externos
DN_HUGO_BIN=hugo
DN_PAGEFIND_BIN=pagefind

# Groq
GROQ_API_KEY=                   # sk_...
GROQ_MODEL=llama-3.3-70b-versatile
GROQ_MAX_PER_DAY=100

# Sesión
DN_SESSION_TTL_HOURS=168        # 7 días

# Sitio
DN_SITE_URL=https://diegonoticias.com
```

`.env` real va en VPS y NO al repo.

---

## 18. Convenciones de código

### 18.1 Go

- Formateo: `gofmt -s` (incluido en CI).
- Linter: `go vet` + opcionalmente `staticcheck`.
- Estructura: `cmd/` + `internal/`. Nada de `pkg/` salvo que algo se publique como librería.
- Errores: `fmt.Errorf("contexto: %w", err)`. Nunca `errors.New("...")` de un error wrappeado.
- Logs: `slog.Info("event", "key", val)`. Nunca `fmt.Println`.
- Concurrencia: `sync.Mutex` para estado compartido. `context.Context` propagado en handlers.
- Tests: `_test.go` al lado del archivo. Tabla de casos cuando aplica. No mocks de stdlib.

### 18.2 Vue/TS

- Composition API + `<script setup lang="ts">` siempre.
- Componentes: PascalCase. Composables: `useXxx`.
- Tipado estricto: `tsconfig.json` con `strict: true`.
- Linter: ESLint + reglas de Vue. Prettier para formato.
- Imports absolutos: `@/components/...` con alias en Vite.
- Pinia stores con typed state y actions async.

### 18.3 CSS/Tailwind

- Mobile-first: `class="text-base md:text-lg"`.
- No `!important`.
- Componentes: agrupar utilities largas en `@apply` solo si se reutilizan en 3+ lugares.
- Variables CSS para colores de marca; resto via Tailwind config extendida.

### 18.4 Commits

Formato sugerido:
```
fase X: <imperativo en presente, ≤72 chars>

(opcional: cuerpo explicando el porqué)
```

Ejemplo: `fase 4: implementar CRUD de artículos en admin SPA`.

---

## 19. Fases de implementación

> **Cada fase tiene**: objetivos, archivos a tocar, tareas detalladas y "Definition of Done" (DoD). Solo avanza a la siguiente cuando la actual cumpla DoD.

### Fase 1 — Esqueleto del proyecto

**Objetivo**: binario Go que arranca, sirve "Diego Noticias" en `/`, integrado con nginx y TLS en VPS.

**Archivos**:
- `go.mod`, `cmd/server/main.go`, `internal/config/config.go`
- `.gitignore`, `.env.example`, `LICENSE` (MIT), `README.md` esqueleto
- `scripts/dev.sh`

**Tareas**:
1. `go mod init github.com/awakeelectronik/diegonoticias`.
2. Crear `cmd/server/main.go` que levante `http.Server` en `:8080` con un handler que devuelva `<h1>Diego Noticias</h1>`.
3. Logging con `slog` JSON en producción, texto en dev.
4. `internal/config/config.go` que carga `.env` (godotenv) y env vars.
5. Graceful shutdown en SIGTERM.
6. `Makefile` o scripts con `make dev`, `make build`, `make test`.

**DoD**:
- `go build ./...` compila sin warnings.
- `go vet ./...` limpio.
- `./diegonoticias` arranca y `curl localhost:8080` devuelve el HTML.
- README explica cómo correr en local.
- Configurado `.gitignore`, `.env.example`, `LICENSE`.

---

### Fase 2 — Auth admin

**Objetivo**: login funcional con sesión + CSRF. Comando `setup-admin` operativo. SPA admin placeholder protegida.

**Archivos**:
- `internal/auth/{auth,session,password,csrf}.go`
- `internal/api/{auth,middleware,router}.go`
- `cmd/server/setup_admin.go`
- `web/admin/` proyecto Vite inicial con login mínimo

**Tareas**:
1. Crear el proyecto `web/admin/` con Vite + Vue 3 + TS + Tailwind v4 (vía `@tailwindcss/vite`) + Vue Router + Pinia.
2. Vista `Login.vue` con formulario simple (POST a `/admin/api/login`).
3. App.vue con router-view; route `/login` y `/articulos` (placeholder "Hola, diego").
4. Guard de router que verifica sesión llamando a `/admin/api/me`.
5. Backend:
   - `setup-admin` CLI (lee usuario y password por TTY con `golang.org/x/term`).
   - `POST /login`, `POST /logout`, `GET /me` con cookie de sesión y CSRF token.
   - Middleware `authRequired` y `csrfRequired`.
   - Sesiones en memoria (`sync.Map`).
6. Embed del SPA admin con `//go:embed all:web/admin/dist` y handler que sirve archivos estáticos en `/admin/`.
7. SPA fallback: si la ruta dentro de `/admin/` no es archivo, servir `index.html`.

**DoD**:
- `./diegonoticias setup-admin` crea `data/admin.json`.
- `./diegonoticias` levanta server.
- `cd web/admin && npm run build` compila SPA.
- En `/admin` se ve la pantalla de login.
- Login con credenciales correctas redirige a `/admin/articulos` y muestra "Hola, diego".
- Login con credenciales malas muestra error.
- Logout regresa a login y `/me` devuelve 401.
- Reiniciar el binario invalida sesiones (esperado, sin persistencia).

---

### Fase 3 — Hugo base

**Objetivo**: proyecto Hugo con layouts mínimos y Tailwind v4 funcionando, output servido por el binario Go.

**Archivos**:
- `site/hugo.toml`, `site/archetypes/articulos.md`
- `site/layouts/_default/{baseof,single,list}.html`
- `site/layouts/index.html`, `site/layouts/404.html`
- `site/layouts/partials/{head/meta,head/og,head/jsonld,header,footer,article-card,pagination}.html`
- `site/assets/css/main.css`, `site/assets/js/main.js`
- `site/static/` (favicon placeholder)
- `internal/builder/builder.go`
- 2–3 artículos seed en `site/content/articulos/`

**Tareas**:
1. `hugo new site site --format toml --force` (ajustar lo necesario para no ensuciar).
2. Configurar `hugo.toml` (sección 8.1).
3. Layouts mínimos pero funcionales con clases Tailwind básicas (no diseño final aún).
4. `assets/css/main.css` con `@import "tailwindcss"` y variables de marca.
5. Pipe Tailwind v4 vía `css.TailwindCSS`.
6. Builder Go: `internal/builder/builder.go` con método `Trigger()` que invoca `hugo --minify`.
7. Endpoint público `GET /` que sirve archivos desde `site/public/` con cache headers correctos:
   - `*.html`: `Cache-Control: no-cache` (siempre revalidar).
   - `_astro/*`, `*.css?v=...`, hashed assets: `Cache-Control: public, max-age=31536000, immutable`.
8. Crear 2-3 artículos seed manualmente en `site/content/articulos/`.
9. Comando `./diegonoticias` ejecuta build inicial al arrancar si `site/public/` no existe.

**DoD**:
- `hugo` invocado por el binario al arranque produce `site/public/` con HTMLs.
- En `/` se ven los artículos seed listados.
- En `/<slug>/` se ve el artículo individual.
- Tailwind aplica estilos.
- Lighthouse SEO > 90 ya en este punto.
- 404 funciona.

---

### Fase 4 — CRUD de artículos (sin IA aún)

**Objetivo**: admin puede crear, listar, buscar, editar y borrar artículos. Cada acción dispara rebuild de Hugo. La generación con IA viene en F6.

**Archivos**:
- `internal/articles/{article,frontmatter,slug,store}.go`
- `internal/api/articles.go`
- `web/admin/src/views/{Articulos,ArticuloEditor}.vue`
- `web/admin/src/api/articles.ts`
- `web/admin/src/stores/articles.ts`
- `web/admin/src/components/{NavBar,Toast,Spinner,Modal}.vue`

**Tareas**:
1. Articles store en Go con List/Get/Create/Update/Delete.
2. Slug generator con `gosimple/slug` y manejo de colisiones.
3. Frontmatter parser/serializer con `adrg/frontmatter`.
4. Endpoints API protegidos.
5. Builder triggered en cada mutación (Create, Update, Delete).
6. SPA: vista `Articulos.vue` con tabla/lista, búsqueda client-side por título, paginación.
7. SPA: vista `ArticuloEditor.vue` con campos básicos (título, body Markdown, tono, descripción, categoría, alt, imagen — esta última solo placeholder por ahora).
8. Toast / spinner / modal de confirmación.

**DoD**:
- Crear un artículo en admin → aparece en `/` del sitio público en < 5s tras publicar.
- Editar un artículo → cambia en sitio público.
- Borrar un artículo → desaparece (con confirmación).
- Búsqueda en admin filtra correctamente.
- Validaciones: título no vacío al guardar (si todos los campos requeridos están), body 170-230 palabras, etc.
- Tests unitarios: slug generator (acentos, longitud, colisiones), frontmatter round-trip, validación de body.

---

### Fase 5 — Pipeline de imágenes

**Objetivo**: admin puede subir imágenes que se comprimen y sirven con `<picture>` srcset.

**Archivos**:
- `internal/images/pipeline.go`
- `internal/api/images.go`
- `site/layouts/partials/picture.html`
- `web/admin/src/components/ImageUpload.vue`

**Tareas**:
1. Instalar libvips en sistema (documentar). Importar `govips`.
2. `ProcessArticleImage(input []byte) (basePath string, err error)`.
3. Endpoint `POST /admin/api/imagenes` con multipart, devuelve `{basePath, alt}`.
4. Servir `/images/*` desde `static-uploads/images/` con cache headers immutable.
5. Componente `ImageUpload.vue` con preview, drag-and-drop opcional.
6. Integrar en `ArticuloEditor.vue`.
7. Plantilla Hugo `partials/picture.html` usando frontmatter.image.
8. Render en `articulos/single.html` y en `article-card.html`.

**DoD**:
- Subir un JPG de 4 MB → guarda 8 archivos (4 anchos × 2 formatos), peso total <500 KB.
- Borrar el original tras procesar.
- Admin muestra preview de la imagen subida.
- Sitio público renderiza con `<picture>` correcto.
- Network tab del navegador muestra que solo carga el AVIF al tamaño correcto.
- Test unitario: pipeline procesa una imagen sintética sin error.

---

### Fase 6 — Integración con Groq

**Objetivo**: generar artículo con un click. Streaming en vivo. Editable. Regenerable. Falla bien.

**Archivos**:
- `internal/ai/{client,prompt,stream}.go`
- `internal/api/ai.go`
- `internal/ratelimit/ratelimit.go`
- `web/admin/src/api/ai.ts`
- `web/admin/src/composables/useSSE.ts`
- `web/admin/src/components/StreamingTextarea.vue`

**Tareas**:
1. Cliente Groq con `net/http` + `encoding/json`.
2. Plantilla del prompt como `text/template` (sección 11.3).
3. SSE handler que proxy-ea el stream Groq → admin.
4. Rate limiter diario (env `GROQ_MAX_PER_DAY`).
5. Validación post-stream del JSON.
6. Frontend: composable `useSSE` que hace POST con body y consume `ReadableStream`.
7. Frontend: parser de JSON parcial (`partial-json`) que actualiza UI por chunk.
8. Botones "Generar", "Detener", "Regenerar" en `ArticuloEditor.vue`.
9. Si falla la API: mostrar toast en español, dejar campos editables, no perder el `rawText`.

**DoD**:
- Con `GROQ_API_KEY` válida, un input de prueba genera un artículo en 2-5 segundos.
- Streaming visible (palabras aparecen progresivamente).
- "Detener" cancela limpiamente.
- "Regenerar" sobrescribe sin perder el input crudo.
- Tras generar, todos los campos son editables.
- Sin API key o cuota agotada, error en español, app sigue usable.
- Cap de 100/día aplicado (test manual seteando MAX a 2).

---

### Fase 7 — Templates finales y búsqueda

**Objetivo**: sitio público con el diseño de los mockups. Pagefind operativo. Mobile bottom nav.

**Archivos**:
- Refinar todos los `site/layouts/**/*.html`.
- `site/assets/css/main.css` con tokens de diseño.
- `site/assets/js/search.js` con cliente Pagefind.
- `site/layouts/partials/{search,article-card-hero}.html`.
- Builder: integrar `pagefind --site site/public`.

**Tareas**:
1. Implementar diseño según mockups (referencia: los 3 PNG enviados).
2. Mobile-first: stack vertical de tarjetas, hero al inicio, bottom nav fixed.
3. Desktop: hero grande + grid de 3 columnas, sidebar opcional en single.
4. Tipografía: Inter para UI, Source Serif 4 para títulos (self-hosted woff2).
5. Paleta neutra cálida (variables CSS).
6. Search: overlay activado desde header / bottom nav, carga lazy de Pagefind.
7. Builder ejecuta Pagefind tras Hugo.
8. Test de accesibilidad: navegar todo con teclado, contraste AA.

**DoD**:
- Capturas del sitio matchean razonablemente los mockups.
- Búsqueda funciona y devuelve resultados con highlights.
- Bottom nav solo en mobile.
- Lighthouse mobile: 95+ en SEO, Performance, Accessibility, Best Practices.
- LCP < 1.5s en 4G simulado.
- JS bundle inicial < 5 KB (sin contar Pagefind lazy).

---

### Fase 8 — Sistema de publicidad

**Objetivo**: CRUD de banners con constraints. Slot rendering. AdSense toggleable.

**Archivos**:
- `internal/ads/{store,validate}.go`
- `internal/api/ads.go`
- `web/admin/src/views/{Publicidad,PublicidadEditor}.vue`
- `web/admin/src/api/ads.ts`
- `site/layouts/partials/ad-slot.html`

**Tareas**:
1. Store de banners con `data/ads.json`.
2. Validaciones (sección 13.2).
3. Pipeline de imagen para banners (más agresivo).
4. Endpoints CRUD + activar/desactivar.
5. SPA: lista con thumbnail, estado, slot. Editor con upload.
6. Builder exporta `site/data/ads.toml` solo con activos.
7. Plantilla `ad-slot.html` con lógica de prioridad AdSense > banner propio > vacío.
8. Integración en `index.html` (entre 5to y 6to artículo) y `articulos/single.html` (después del 2do párrafo).

**DoD**:
- Subir 7 banners totales → ok.
- Subir el 8vo → error claro.
- Activar 3 banners → error.
- Activar 2 banners en mismo slot → error.
- Banner activo aparece en sitio público tras rebuild.
- AdSense deshabilitado: banners propios visibles.
- AdSense habilitado en slot 1: AdSense gana sobre banner propio en slot 1.

---

### Fase 9 — SEO completo, ajustes, accesibilidad y polish final

**Objetivo**: cerrar todos los detalles SEO, ajustes editables desde admin, accesibilidad pulida, README completo.

**Archivos**:
- `internal/api/settings.go`, `internal/settings/store.go`
- `web/admin/src/views/Ajustes.vue`
- `site/layouts/partials/head/jsonld.html` (refinar)
- README, CONTRIBUTING.

**Tareas**:
1. Endpoint y SPA para `Ajustes.vue`: nombre del sitio, descripción, twitter handle, OG default, AdSense config.
2. JSON-LD `Article` completo (datePublished, dateModified, image, author).
3. JSON-LD `WebSite` con `SearchAction` en home.
4. JSON-LD `BreadcrumbList`.
5. RSS feed validado.
6. Sitemap con `lastmod` por artículo.
7. Auditoría accesibilidad: axe DevTools sin errores críticos.
8. README con: requisitos, setup local, deploy, comandos, screenshots.
9. CONTRIBUTING con reglas básicas.
10. GitHub Actions: workflow de CI que corre `go vet`, `go test`, build admin, build hugo con dataset seed.
11. Configurar GoAccess en VPS.

**DoD**:
- Lighthouse mobile: 95+ en las 4 categorías.
- Validador JSON-LD: rich results detectado para artículos.
- Validador OG: card preview correcto.
- Sitemap y RSS válidos.
- README permite a un tercero clonar y correr local en <30 min.
- CI pasa en verde.
- Repo público listo para anunciar.

---

## 20. Checklist de open source

Al cerrar Fase 9, antes de publicar el repo:

- [ ] LICENSE MIT con año y nombre.
- [ ] README en español con: descripción, screenshots, badges (build status, license, go version), requisitos, setup local, deploy, troubleshooting, FAQ, créditos.
- [ ] CONTRIBUTING.md con guía de issues, PRs, estilo de commits.
- [ ] CODE_OF_CONDUCT.md (Contributor Covenant).
- [ ] `.env.example` con TODAS las vars documentadas.
- [ ] Auditar historial: `git log -p | grep -i -E "(api[_-]?key|password|secret|token)"` debe estar limpio.
- [ ] Issue templates (bug, feature).
- [ ] PR template.
- [ ] GitHub Actions: CI verde.
- [ ] Topics: `cms`, `hugo`, `go`, `golang`, `vue`, `vue3`, `typescript`, `tailwindcss`, `groq`, `ai`, `news`, `blog`, `static-site-generator`.
- [ ] Demo deployable: `make demo` o instrucciones claras.
- [ ] Captura de pantalla principal en README.
- [ ] Versión inicial taggeada `v0.1.0`.

---

## 21. Glosario

- **SSG**: Static Site Generator. Genera HTML al momento de build.
- **SSR**: Server-Side Rendering. Genera HTML por petición.
- **SSE**: Server-Sent Events. Streaming unidireccional servidor→cliente sobre HTTP.
- **Frontmatter**: bloque YAML al inicio de un Markdown con metadata.
- **Permalink**: URL canónica de una página.
- **Slug**: parte de la URL derivada del título.
- **Hero**: el elemento visual principal de una página, arriba de todo.
- **DoD**: Definition of Done.

---

## Final

Si llegaste leyendo hasta aquí (humano o agente), tienes contexto suficiente para arrancar.

**Próximo paso**: empezar con la **Fase 1**.

- Path de módulo Go fijado: **`github.com/awakeelectronik/diegonoticias`**.
- El acceso al VPS no es necesario hasta el primer despliegue (entre F3 y F9, según se quiera probar en producción).

Cualquier ambigüedad, **pregunta antes de improvisar**.
