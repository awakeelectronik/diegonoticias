# Diego Noticias

Sitio minimalista de noticias con:

- **Sitio público** generado con Hugo (estático, SEO-first).
- **Admin** SPA (Vue) embebida en un binario Go.

## Requisitos (dev)

- Go 1.23+

## libvips y ahorro de RAM

El runtime de imágenes se inicializa con `MaxCacheSize=0` y `MaxCacheMem=0` para minimizar
residuo de memoria entre procesamientos.

- Build normal (sin libvips): no activa runtime VIPS.
- Build con libvips: compilar con tag `vips`.

```bash
go build -tags vips -o diegonoticias ./cmd/server
```

## Correr en local (fase 1)

1. Copia variables:

```bash
cp .env.example .env
```

2. Inicia el servidor:

```bash
make dev
```

3. Prueba:

```bash
curl -i http://127.0.0.1:8080/
```

