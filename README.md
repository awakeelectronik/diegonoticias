# Diego Noticias

> Blog de noticias minimalista con generación de artículos por IA. Sitio estático Hugo + admin Vue, todo en un binario Go.

[![CI](https://github.com/awakeelectronik/diegonoticias/actions/workflows/ci.yml/badge.svg)](https://github.com/awakeelectronik/diegonoticias/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

<!-- Pega aquí una captura del home y otra del editor admin cuando las tengas:
     ![Home](docs/screenshot-home.png)
     ![Editor admin](docs/screenshot-admin.png)
-->

## Qué es

- **Sitio público** estático generado con Hugo. SEO-first, AVIF/WebP, Lighthouse alto.
- **Admin** SPA en Vue 3 + TypeScript, embebida en el binario Go vía `go:embed`, montada en `/admin/`.
- **Generación con IA** (Groq, modelo `llama-3.3-70b-versatile`): pegas un texto crudo, eliges un tono, y el backend devuelve título + cuerpo (~200 palabras) + meta + categoría + alt en un JSON. Lo editas si quieres y publicas.
- **Sin base de datos**: la fuente de verdad son los Markdown de Hugo y unos JSON en `data/`. Cada mutación dispara `hugo --minify` automáticamente.
- **Imágenes** procesadas con libvips → 4 anchos × {AVIF, WebP}, servidas inmutables.
- **Publicidad propia** (banners con slots) + integración opcional con AdSense.

## Stack

Go 1.25 · Hugo extended · Vue 3.5 + Vite + Pinia + Tailwind v4 · libvips (govips) · Argon2id · Groq.

## Setup local

Requisitos: Go 1.25+, Node 22+, Hugo extended 0.140+, libvips (para procesar imágenes).

```bash
# 1. Clonar y entrar
git clone https://github.com/awakeelectronik/diegonoticias.git
cd diegonoticias

# 2. Variables de entorno (mínimo: GROQ_API_KEY si quieres usar la IA)
cp .env.example .env
$EDITOR .env

# 3. Build del admin (genera web/admin/dist/)
cd web/admin && corepack pnpm install && corepack pnpm build && cd ../..

# 4. Compilar binario con soporte de imágenes
make build

# 5. Crear usuario admin (interactivo)
./diegonoticias setup-admin

# 6. Levantar
./diegonoticias
```

Sitio público en `http://127.0.0.1:8080/`, admin en `http://127.0.0.1:8080/admin/`.

> Sin `GROQ_API_KEY` el sitio funciona pero el botón **Generar** del editor falla. Todo lo demás (CRUD manual, publicidad, ajustes, imágenes) sigue operativo.

## Comandos

```bash
make dev      # go run, sin tag vips (sin pipeline de imágenes)
make build    # compila con tag vips, requiere libvips en sistema
make vet      # go vet ./...
make tidy     # go mod tidy
```

## Despliegue

Binario único detrás de nginx (TLS, gzip/brotli). Sin Docker, sin systemd unit en el repo todavía. Estructura recomendada en VPS y detalles operativos en [`ESTADO.md`](ESTADO.md).

## Documentación

- [`ESTADO.md`](ESTADO.md) — arquitectura, endpoints, modelo de datos, variables de entorno y notas de operación. Lectura obligatoria antes de tocar el código.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — flujo de PRs.
- [`PLAN.md`](PLAN.md) — plan original de implementación (histórico, no fiel al estado actual).

## Licencia

MIT — ver [`LICENSE`](LICENSE).
