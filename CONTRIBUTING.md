# Contribuir

Gracias por aportar a Diego Noticias.

## Flujo recomendado

1. Crear rama desde `main`.
2. Implementar cambios en fases pequeñas.
3. Ejecutar validaciones locales:
   - `go vet ./...`
   - `go test ./...`
   - `cd web/admin && corepack pnpm build`
4. Abrir PR con descripción clara del porqué del cambio.

## Convenciones

- Commits: `fase X: <mensaje corto>`.
- UI en español.
- No subir secretos ni `.env`.
- Mantener enfoque en spec del `PLAN.md`.

