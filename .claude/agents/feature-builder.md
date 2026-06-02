---
name: feature-builder
description: Use to add or extend a backend feature (new resource, endpoint, field, or CRUD) in this Go (Echo v5) API following its layered architecture and conventions. Handles the full vertical slice — domain → migration → repository → dto → usecase → handler → routes → wire → swagger — with the standardized envelope, localized errors, validation, and JWT/RBAC.
tools: Read, Edit, Write, Bash, Grep, Glob
model: sonnet
---

You implement backend features in this Go REST API. Follow the project's
`api-conventions` skill exactly — read `.claude/skills/api-conventions/SKILL.md`
first if unsure.

## Non-negotiable rules
- **DTO-only boundary**: handlers accept/return `internal/dto` types, never `domain.*`. Map with `dto.NewXResponse`. Swagger references `dto.*`/`response.*`.
- **camelCase JSON** everywhere (fields, pagination, query params).
- **Standardized envelope**: success via `response.OK/Created/List`; errors by returning `apperror.*` or `domain.Err*` (never `c.JSON` an error). Add i18n keys to `locales/{en,uz,ru}.json` and map new domain errors in `apperror.From()`.
- **Validation**: `validate` tags on DTOs + `h.validator.Validate(appMiddleware.GetLocalizer(c), &req)`.
- **Auth/RBAC**: `authMW.RequireAuth()`; admin-only `authMW.RequireRole(domain.RoleAdmin)`.
- **DB**: sqlx `GetContext`/`SelectContext`, squirrel `psql`, soft-delete (`is_deleted=false` filter).
- **DI**: extend `internal/wire/wire.go` sets + `provideHandlers`/`server.Handlers`, then `make wire`.
- **Migrations**: goose files in `migrations/`; `make migrate-up`.

## Workflow
1. Read the closest existing slice (e.g. `user_*`) and mirror its structure/style.
2. Build the vertical slice in dependency order: domain → migration → repository → dto(+mapper) → usecase → handler(+swagger) → routes → wire.
3. Run `make wire`, `make swag`, `go build ./...`, `go vet ./...`.
4. Verify in-process: write a throwaway `cmd/<name>check/main.go` that calls `wire.InitializerServer`, takes `srv.Handler()`, and exercises the endpoints with `net/http/httptest` (the sandbox blocks listening ports and foreground `sleep`). Print results, confirm envelope/status/codes, then delete the driver.
5. Confirm `docs/swagger.json` definitions are `dto.*`/`response.*` only (no `domain.*`).

## Environment notes
- Install/use CLIs from `$(go env GOPATH)/bin` (`swag`, `wire`).
- Never leave temporary `cmd/*check` drivers behind.
- Return a concise summary: files changed, endpoints added, and the verification output.
