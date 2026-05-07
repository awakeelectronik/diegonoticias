# Diego Noticias — estado actual

Snapshot funcional del proyecto a 2026-05-07. Sustituye al `PLAN.md` original como referencia de "qué es esto hoy". El plan se conserva como contexto histórico pero ya no es fiel.

---

## 1. Qué es

Sitio de noticias minimalista en español (es-CO) con dos piezas en un solo binario Go:

- **Sitio público**: estático, generado por Hugo. Servido por el binario en `127.0.0.1:8080` detrás de nginx (TLS + compresión).
- **Admin SPA**: Vue 3 + TS + Pinia + Tailwind v4, embebida vía `go:embed` y montada en `/admin/`. CRUD de artículos, publicidad y ajustes.

Flujo central: el admin pega un texto crudo + escoge tono → backend llama a Groq y devuelve un JSON `{title, body, metaDescription, category, imageAlt}` → el admin lo edita si quiere → al guardar se escribe el `.md` y se dispara `hugo --minify`.

---

## 2. Arquitectura

```
Internet ── nginx (TLS, gzip, brotli) ──► Go binary :8080
                                          ├─ GET  /              → site/public/* (estático)
                                          ├─ GET  /articulos/    → site/public/articulos/*
                                          ├─ GET  /admin/*       → SPA Vue (embed)
                                          ├─ *    /admin/api/*   → handlers JSON
                                          └─ GET  /images/*      → static-uploads/images/*

site/content/articulos/*.md   ← fuente de verdad (Markdown + frontmatter YAML)
data/admin.json               ← user + hash argon2id
data/settings.json            ← branding + AdSense
data/ads.json                 ← banners propios
static-uploads/images/        ← variantes AVIF/WebP de cada imagen subida

Cada mutación admin → builder.Build() → exporta data/site.toml + data/ads.toml a site/data/, corre `hugo --minify` y `pagefind` (vestigio).
```

Sesiones en memoria (`sync.Map`) — reiniciar el binario invalida sesiones. Sin base de datos.

---

## 3. Estructura del repo

```
cmd/server/
  main.go                  # arranque, graceful shutdown, init runtime libvips
  setup_admin.go           # CLI `./diegonoticias setup-admin` (TTY o env vars)

internal/
  config/config.go         # carga .env (en dev) y env vars DN_*
  api/
    router.go              # mux + middlewares + adminSPA + publicSite + images handlers
    middleware.go          # authRequired, csrfRequired (cookie + header X-CSRF-Token)
    auth.go                # POST /login, GET /me, POST /logout
    articles.go            # CRUD artículos (dispara builder.Build() en cada mutación)
    ai.go                  # POST /articulos/generar (síncrono; cap diario via DailyLimiter)
    images.go              # POST /imagenes (multipart, máx 8MB)
    ads.go                 # CRUD banners
    settings.go            # GET/PUT ajustes
  auth/
    admin_store.go         # data/admin.json {username, passwordHash, createdAt, updatedAt}
    password.go            # argon2id hash + compare
    session.go             # SessionManager en memoria, TTL 7 días, token + csrf por sesión
    csrf.go                # ConstantTimeCompare
    auth.go                # constante cookie name (dn_session)
  articles/
    article.go             # Frontmatter struct (title, slug, date, description, tone,
                           #   category, image, imageAlt, draft, wordCount) + Article{Body}
    slug.go                # GenerateSlug con gosimple/slug, manejo de colisiones
    store.go               # List/Get/Create/Update/Delete sobre site/content/articulos/*.md
                           # escritura atómica (.tmp + rename), parse con adrg/frontmatter
  ai/
    client.go              # Groq client. Complete() (sync). Modelo:
                           #   llama-3.3-70b-versatile. response_format=json_object.
    prompt.go              # plantilla text/template del prompt + tabla de tonos
  ads/
    store.go               # CRUD sobre data/ads.json
    validate.go            # reglas: total ≤7, máx 2 activos, máx 1 activo por slot,
                           #   slot ∈ {1,2}, título 1..80, imagePath requerido
  builder/builder.go       # Build(): exportData() + `hugo --minify`. Mutex global, registra lastBuild
  images/
    pipeline.go            # ProcessArticleImage: hash 8B + ruta YYYY/MM, escritura atómica
    pipeline_vips.go       # build tag `vips`: govips → 4 anchos × {avif, webp}
    pipeline_stub.go       # build sin tag: error claro
    runtime_vips.go        # vips.Startup MaxCacheSize/Mem=0 (RAM mínima)
  ratelimit/ratelimit.go   # DailyLimiter (cap diario IA, default 100, env GROQ_MAX_PER_DAY)
  settings/store.go        # CRUD data/settings.json (siteName, siteUrl, twitter, AdSense)

site/                      # proyecto Hugo
  hugo.toml                # baseURL, es-CO, permalinks /:slug/, paginación 12, RSS home+sección
  content/articulos/*.md   # 3 artículos seed
  layouts/
    _default/{baseof,single,list}.html
    articulos/single.html  # artículo individual con hero image + ad-slot 2
    index.html             # home: hero (1er artículo) + grid + ad-slot 1 entre 5to-6to
    partials/
      head/{meta,og,jsonld}.html
      header.html          # brand DIEGO+NOTICIAS + link Admin
      footer.html          # SVG inline RRSS (FB/IG/X) + copy
      article-card.html    # tarjeta con thumb + categoría
      ad-slot.html         # AdSense > banner propio > vacío
      picture.html         # <picture> AVIF+WebP × 4 anchos
      pagination.html
  assets/css/main.css      # tokens de marca + componentes (no Tailwind en sitio público)

web/admin/                 # SPA Vite
  package.json             # vue 3.5, vue-router 4, pinia 2, tailwindcss v4
  src/
    router.ts              # /login /articulos /articulos/nuevo /articulos/:slug/editar
                           # /publicidad /publicidad/nueva /publicidad/:id/editar /ajustes
    stores/auth.ts         # estado de sesión + csrfToken
    api/{client,auth,articles,ads,images,settings}.ts
    views/
      Login.vue
      Articulos.vue        # lista + acciones
      ArticuloEditor.vue   # textarea raw + tono + imagen + botón Generar (sync) + preview
      Publicidad.vue
      PublicidadEditor.vue
      Ajustes.vue
    components/
      ImageUpload.vue
  dist/                       # build de Vite, COMMITTEADO (ver .gitignore: web/admin/dist NO ignorado)

scripts/dev.sh             # `go run ./cmd/server` con DN_ENV=development
Makefile                   # dev / build (-tags vips) / test / vet / tidy
.github/workflows/ci.yml   # vet + test + build backend + admin + hugo
data/                      # gitignored, runtime
static-uploads/            # gitignored, runtime
```

**Qué NO existe** (pese a que el plan lo mencionaba): `internal/articles/frontmatter.go`, `slug_test.go`, `prompt_test.go`. Los tests fueron descartados a favor de pruebas manuales.

---

## 4. Stack y versiones

| Capa | Tech |
|---|---|
| Backend | Go 1.25, `net/http` con router patterns 1.22+ |
| Sitio público | Hugo extended, Markdown + YAML frontmatter |
| CSS sitio | CSS plano con tokens (no Tailwind) |
| Admin | Vue 3.5 + TS + Vite + Pinia + Vue Router + Tailwind v4 (`@tailwindcss/vite`) |
| IA | Groq, modelo `llama-3.3-70b-versatile`, `response_format: json_object` |
| Imágenes | govips (libvips) → AVIF (Q55) + WebP (Q72) en 320/640/1024/1600 |
| Auth | Argon2id (`alexedwards/argon2id`) + cookie HttpOnly SameSite=Strict + CSRF header |
| Frontmatter | `adrg/frontmatter` + `gopkg.in/yaml.v2` |
| Slugs | `gosimple/slug` con `MakeLang("es")` |
| TOML export | `BurntSushi/toml` |
| .env | `joho/godotenv` (solo en dev) |

`go.mod` declara `go 1.25.0`. El README dice "Go 1.23+" — desactualizado.

---

## 5. Modelo de datos

### Frontmatter de artículo (`site/content/articulos/{slug}.md`)

```yaml
title: "El ritual del objeto"
slug: "el-ritual-del-objeto"
date: 2026-05-05T18:00:00-05:00
description: "..."          # meta description, 140-160 chars
tone: "conversacional"      # uno de los 10 tonos cerrados
category: "design"          # palabra libre devuelta por la IA
image: "/images/2026/05/abcd1234"   # opcional, basePath sin sufijo
imageAlt: "..."             # opcional
draft: false
wordCount: 176              # calculado en cada Update
```

### `data/admin.json` (gitignored)

```json
{ "username": "diego", "passwordHash": "$argon2id$...", "createdAt": "...", "updatedAt": "..." }
```

### `data/settings.json`

```json
{
  "siteName": "Diego Noticias",
  "siteDescription": "...",
  "siteUrl": "https://diegonoticias.com",
  "defaultOgImage": "/og-default.jpg",
  "twitterHandle": "",
  "adsense": { "enabled": false, "clientId": "", "slot1Id": "", "slot1Enabled": false, "slot2Id": "", "slot2Enabled": false }
}
```

### `data/ads.json`

```json
{ "banners": [ { "id": "uuid", "title": "...", "imagePath": "/images/...", "active": true, "slot": 1, "createdAt": "...", "updatedAt": "..." } ] }
```

### Datos derivados (escritos por el builder en cada Build)

- `site/data/site.toml` ← copia de `data/settings.json`
- `site/data/ads.toml` ← solo banners con `active: true`

Hugo los lee con `site.Data.site` y `site.Data.ads`.

---

## 6. Endpoints

Todos bajo `/admin/api/*`. JSON. Auth por cookie `dn_session`. Mutaciones requieren header `X-CSRF-Token`.

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| POST | `/login` | — | login, devuelve `{username, csrfToken}` y setea cookie |
| GET | `/me` | sesión | devuelve sesión actual |
| POST | `/logout` | CSRF | invalida sesión |
| GET | `/articulos` | sesión | lista artículos (ordenados por fecha desc) |
| GET | `/articulos/{slug}` | sesión | un artículo |
| POST | `/articulos` | CSRF | crear (auto-slug si vacío, rebuild) |
| PUT | `/articulos/{slug}` | CSRF | actualizar (rename si cambia slug, rebuild) |
| DELETE | `/articulos/{slug}` | CSRF | borrar (rebuild) |
| POST | `/articulos/generar` | CSRF | IA síncrona, devuelve JSON `{title, body, metaDescription, category, imageAlt}` |
| POST | `/imagenes` | CSRF | multipart `image` + `alt`, devuelve `{basePath, alt}` |
| GET | `/ajustes` | sesión | leer settings |
| PUT | `/ajustes` | CSRF | guardar settings (NO dispara rebuild) |
| GET | `/publicidad` | sesión | listar banners |
| POST | `/publicidad` | CSRF | crear (rebuild) |
| PUT | `/publicidad/{id}` | CSRF | actualizar (rebuild) |
| DELETE | `/publicidad/{id}` | CSRF | borrar (rebuild) |

Headers de seguridad globales: `X-Content-Type-Options`, `Referrer-Policy: strict-origin-when-cross-origin`, `X-Frame-Options: DENY`, `Permissions-Policy: interest-cohort=()`, CSP estricta (`default-src 'self'`).

---

## 7. Flujo de creación de artículo

1. Admin entra a `/admin/articulos/nuevo`.
2. Pega texto crudo en textarea, escoge tono, opcionalmente sube imagen y escribe título-hint.
3. Clic en **Generar** → `POST /admin/api/articulos/generar` → backend:
   - Verifica cap diario (`DailyLimiter`).
   - Construye prompt desde plantilla (`prompt.go`), incluye tono + descripción + título-hint + flag hasImage.
   - Llama a Groq con `stream: false`, `response_format: json_object`.
   - Parsea JSON. Si `len(words(body)) < 150`, reintenta UNA vez añadiendo nota interna pidiendo 180-230 palabras. Devuelve la versión más larga.
4. Admin recibe `{title, body, metaDescription, category, imageAlt}`, edita lo que quiera en la preview.
5. Clic en **Guardar** → `POST /admin/api/articulos`:
   - Genera slug con `gosimple/slug` (es), trunca a 80, resuelve colisiones con sufijo `-2`, `-3`...
   - `wordCount` recalculado.
   - Escritura atómica del `.md` (`.tmp` + rename).
   - `builder.Build()`: exporta TOML derivados, corre `hugo --minify` (build con mutex global).
6. Sitio público actualizado.

**Falla bien**: sin `GROQ_API_KEY`, error en español devuelto al admin sin perder el `rawText`.

---

## 8. Pipeline de imágenes

Compilación con tag `vips` (requiere libvips en sistema). Sin el tag, el binario corre pero subir imagen falla con error claro.

- Entrada: multipart, máx 8MB.
- `ProcessArticleImage` genera hash de 8 bytes, ruta `images/YYYY/MM/{hash}`.
- Por cada ancho de `[320, 640, 1024, 1600]`: redimensiona con Lanczos3 (incluso hacia arriba si el original es más estrecho), exporta AVIF y WebP, escritura atómica.
- Devuelve `basePath` SIN extensión ni ancho. El template `picture.html` arma el `<picture>` con AVIF + WebP × 4 anchos.
- Verificación post-pipeline: comprueba que el archivo `{base}-320.webp` existe; si no, error.
- Servido en `/images/*` con `Cache-Control: public, max-age=31536000, immutable`.

Mismo pipeline se reutiliza para banners.

---

## 9. Sitio público (Hugo)

- **Home**: hero del primer artículo (con imagen prioritaria, fetchpriority=high), grid del resto, ad-slot 1 inyectado entre la 5ª y 6ª tarjeta.
- **Artículo single** (`articulos/single.html`): título, fecha, hero image opcional, contenido, ad-slot 2 al final.
- **Header**: brand wordmark `DIEGO`+`NOTICIAS` con tipografía diferenciada, link a `/admin/`.
- **Footer**: SVG inline para Facebook / Instagram / X (URLs hardcodeadas en partial). Copyright dinámico con `now.Year`.
- **Bottom-nav móvil**: 2 items (Inicio · Admin). Originalmente eran 3 — el de "Buscar" se retiró.
- **JSON-LD**: `Article` por página, `WebSite` simple en home (sin SearchAction).
- **OG/Twitter**: meta tags completos en `head/og.html`, `twitter:card=summary_large_image`.
- **RSS**: outputs `["HTML", "RSS"]` en home y sección.
- **Permalinks**: `/:slug/` con trailing slash.
- **Cache headers** (handler Go): `*.html` → `no-cache`, todo lo demás → `public, max-age=31536000, immutable`.
- **CSS**: `assets/css/main.css` plano con variables CSS de marca. **NO** se usa Tailwind en el sitio público (sí en el admin).

---

## 10. Admin SPA

Rutas (router con `createWebHistory('/admin/')`):

- `/login`
- `/articulos`, `/articulos/nuevo`, `/articulos/:slug/editar`
- `/publicidad`, `/publicidad/nueva`, `/publicidad/:id/editar`
- `/ajustes`
- `/` redirige a `/articulos`

Guard global: si no hay sesión, redirige a `/login`. Si hay sesión y va a `/login`, redirige a `/articulos`.

Store `auth` (Pinia) mantiene `username`, `csrfToken`, `checked`. Llama a `GET /me` en boot.

Vista `ArticuloEditor.vue`: textarea de texto crudo, select de tono (10 opciones), `ImageUpload`, botones **Generar** / **Guardar** / **Cancelar**. Preview en tarjeta debajo con conteo de palabras y categoría. Anillo verde tras generar exitoso.

`ImageUpload.vue` muestra preview vía `/images/{basePath}-640.webp` y reporta errores si las variantes no se generaron en el VPS.

---

## 11. Auth

- Comando: `./diegonoticias setup-admin` interactivo (TTY oculta password con `golang.org/x/term`) o batch via `DN_SETUP_USERNAME` + `DN_SETUP_PASSWORD`.
- Username default: `diego`. Password mín 8 caracteres. Hash Argon2id (`DefaultParams`).
- Login → cookie `dn_session` (HttpOnly, SameSite=Strict, Path=/admin, `Secure: false` — confiar en nginx para TLS).
- TTL sesión: 7 días, sliding expiration NO (expira absoluto).
- CSRF: token random de 32 bytes generado en login, devuelto en JSON, exigido en header `X-CSRF-Token` en mutaciones. `ConstantTimeCompare`.
- Sesiones en memoria → reinicio invalida todas. Sin "remember me", sin refresh token.

---

## 12. Publicidad

Reglas en `internal/ads/validate.go`:

- Slots disponibles: `1` (home, entre 5ª y 6ª tarjeta) y `2` (final del artículo single).
- Total banners ≤ 7.
- Activos simultáneos ≤ 2.
- Máx 1 activo por slot.
- Título 1..80 caracteres, `imagePath` requerido.

Renderizado (`partials/ad-slot.html`): si `adsense.enabled && slotN_enabled` para el slot pedido → bloque AdSense. Si no, primer banner propio activo en ese slot. Si tampoco hay → vacío. La integración real de AdSense (script de Google) está deshabilitada/placeholder.

---

## 13. SEO

- `<title>`, `<meta description>`, canonical en cada página.
- OG completo (type, site_name, title, description, url, image).
- Twitter card `summary_large_image`.
- JSON-LD `Article` (headline, datePublished, dateModified, description, mainEntityOfPage) y `WebSite` en home.
- `enableRobotsTXT = true` en Hugo.
- Sitemap: el default de Hugo (no se ha verificado lastmod por artículo).
- RSS: home y sección.
- Imágenes con AVIF+WebP, lazy loading default, eager + fetchpriority=high para hero.
- Permalinks limpios.

---

## 14. Build, deploy, runtime

```bash
make build          # CGO_ENABLED=1 go build -tags vips -o diegonoticias ./cmd/server
make dev            # go run ./cmd/server
cd web/admin && pnpm install && pnpm build   # genera dist/ (committed)
```

- Build inicial de Hugo se dispara al arrancar el binario si `site/public/index.html` no existe.
- Cada mutación admin invoca `builder.Build()` (mutex global, ~ms en sitios pequeños).
- nginx delante del binario hace TLS, gzip/brotli y reverse-proxy.
- No hay Dockerfile, ni `scripts/deploy.sh`, ni unit systemd en el repo. Despliegue es manual al VPS.

---

## 15. Variables de entorno

Cargadas con `godotenv` solo en `DN_ENV=development`. Defaults entre paréntesis.

| Var | Default | Uso |
|---|---|---|
| `DN_ENV` | `development` | dev habilita carga de `.env` y logs en texto (prod = JSON) |
| `DN_LISTEN` | `127.0.0.1:8080` | dirección del HTTP |
| `DN_DATA_DIR` | `./data` | admin.json, settings.json, ads.json |
| `DN_UPLOADS_DIR` | `./static-uploads` | imágenes procesadas |
| `DN_SITE_DIR` | `./site` | proyecto Hugo |
| `DN_HUGO_BIN` | `hugo` | binario Hugo |
| `DN_LOG_LEVEL` | `info` | debug/info/warn/error |
| `GROQ_API_KEY` | — | requerido para IA |
| `GROQ_MODEL` | `llama-3.3-70b-versatile` | modelo |
| `GROQ_MAX_PER_DAY` | `100` | cap diario de generaciones |
| `DN_SETUP_USERNAME` | — | batch setup-admin |
| `DN_SETUP_PASSWORD` | — | batch setup-admin |

---

## 16. Decisiones removidas frente al plan original

- **Streaming SSE de IA**: el plan lo declaraba como decisión cerrada; el editor usa generación síncrona. Handler SSE y composable cliente fueron eliminados.
- **Buscador Pagefind**: removido del sitio público (UI, partial, JS y SearchAction de JSON-LD) y del builder.
- **Tests**: el plan exigía tests de slug, prompt e imágenes; se prescindió de ellos a favor de QA manual.

---

## 17. Notas de operación

- `cookie.Secure: false` en `internal/api/auth.go`: confía en que nginx termina TLS. Si alguna vez se sirve directo sin proxy TLS, activarlo.
- `web/admin/dist/` está committeado (ver `.gitignore`). Tras cambiar el SPA hay que correr `pnpm build` y commitear el dist.
- `data/` y `static-uploads/` están en `.gitignore`: nunca se versionan. Despliegue tiene que crearlos en el VPS.
- `BuildInitialIfNeeded()` solo dispara Hugo al arranque si `site/public/index.html` no existe. Cambios en plantillas requieren mutación admin (o `hugo` manual) para verse.
- `PUT /admin/api/ajustes` NO dispara rebuild: cambios en `settings.json` (incluyendo branding) solo se reflejan en el sitio tras la siguiente mutación de artículo o banner.

---

## 18. Glosario rápido

- **basePath**: ruta de imagen sin sufijo de ancho ni extensión (ej. `/images/2026/05/abcd1234`). El template arma las 8 variantes.
- **Slot**: posición fija de un banner publicitario (1 = home, 2 = single).
- **Builder**: orquesta export de TOML + Hugo.
- **Frontmatter**: YAML al inicio del Markdown con metadata.
- **Atomic write**: escribir a `{path}.tmp` y renombrar (visible solo cuando está completo).
