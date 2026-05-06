# Diego Noticias

Sitio minimalista de noticias con:

- **Sitio público** generado con Hugo (estático, SEO-first).
- **Admin** SPA (Vue) embebida en un binario Go.

## Requisitos (dev)

- Go 1.23+

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

