---
name: api-conventions
description: Architecture, layering, and coding rules for this Go (Echo v5) REST API. Use whenever adding or modifying a backend feature here — entities, endpoints, DTOs, repositories, usecases, handlers, routes, middleware, migrations, i18n, or Swagger. Enforces DTO-only boundaries, the standardized response envelope, localized errors, JWT/RBAC, wire DI, and goose migrations.
---

# API conventions

Go REST API: **Echo v5** + **sqlx/squirrel/Postgres** + **google/wire** DI +
**go-i18n** (en default, Accept-Language) + **swaggo**. JWT auth with RBAC.
Module: `github.com/yodzafar/url-shortener-app`.

## Layered structure (dependencies point inward)

| Layer | Path | Responsibility |
|---|---|---|
| domain | `internal/domain` | Entities, repository **interfaces**, domain errors. No framework imports. **Never serialized to the API.** |
| dto | `internal/dto` | Request/response payloads + mappers (`NewUserResponse`). Imports only `domain`. |
| repository | `internal/repository` | sqlx + squirrel implementations of domain interfaces. |
| usecase | `internal/usecase` | Business logic. Depends on domain interfaces + dto + pkg. |
| handler | `internal/handler` | Echo handlers: bind → validate → usecase → `response.*`. |
| routes | `internal/server/routes` | Route registration, route groups, auth/role middleware. |
| middleware | `internal/middleware` | `auth` (JWT Bearer + RBAC), `lang` (i18n localizer + request logger), `error` (central handler). |
| apperror | `internal/apperror` | `AppError{Status, Code, MessageID, Data, Details}`, constructors, `From()`. Imports only `domain`. |
| pkg | `internal/pkg/{response,token,validation,logger,db}` | Envelope, JWT manager, validator, colorful logger, DB. |
| config | `internal/config` | Env config (`getEnv`, `mustEnv`, `getDurationEnv`, ...). |
| wire | `internal/wire` | google/wire DI (`wire.go` + generated `wire_gen.go`). |
| cmd | `cmd/{main,migrate,grantadmin}` | Entrypoint, goose runner, admin bootstrap. |
| migrations | `migrations/*.sql` | goose migrations. |
| i18n | `i18n/` + `locales/{en,uz,ru}.json` | Translations. |

## Hard rules

1. **DTO-only boundary.** Handlers accept and return `internal/dto` types — **never** `domain.*` entities. Map domain→DTO with `dto.NewXResponse(...)`. Swagger `@Success`/`@Param` reference `dto.*`, not `domain.*`.
2. **camelCase JSON** for every API field — requests, responses, and pagination (`firstName`, `accessToken`, `pageSize`, `totalItems`). Query params too (`?page=&pageSize=`).
3. **Standardized envelope** for every response via `internal/pkg/response`:
   - Success: `response.OK(c, data)` / `Created` / `List(c, data, response.NewPagination(page, size, total))` → `{success:true, data, meta?, error:null}`.
   - Error: returned errors are rendered by the central handler as `{success:false, data:null, error:{code, message, details?}}`. `data`/`error` are explicit `null` (no `omitempty`); pass literal `nil` for empty payloads.
4. **Never write error JSON in a handler.** Return `apperror.*` (e.g. `apperror.BadRequest().Wrap(err)`, `Unauthorized()`, `Forbidden()`, `Validation(details)`) or a `domain.Err*`. `middleware.ErrorHandler` localizes and renders it. Map new domain errors in `apperror.From()` and add i18n keys to **all three** locale files.
5. **Validation.** Tag DTOs with `validate:"..."`. In handlers: `if err := h.validator.Validate(appMiddleware.GetLocalizer(c), &req); err != nil { return err }` → 422 with field-level `details` keyed by json name. Add new tag→key mappings in `internal/pkg/validation` and locale keys `validation.*`.
6. **Auth & RBAC.** Protect routes with `authMW.RequireAuth()`; admin-only with `authMW.RequireRole(domain.RoleAdmin)`. The user (with role) is loaded from the DB every request via `GetUser(c)` → revocation is instant. Self-service endpoints use the id from `GetUser(c)`.
7. **DB.** Use `sqlx` `GetContext`/`SelectContext` for struct scans, `squirrel` (`psql` builder) for queries. Soft-delete via `is_deleted`; filter `is_deleted=false` in all reads/updates. Map unique-violations to `domain.ErrUserAlreadyExists`.
8. **DI.** Add constructors to the relevant `wire.NewSet` in `internal/wire/wire.go`, extend `provideHandlers`/`server.Handlers` if needed, then `make wire`. Keep `wire_gen.go` in sync (regenerate, don't hand-drift).
9. **Migrations.** Add goose files `migrations/NNNNN_name.sql` (`-- +goose Up/Down`, `StatementBegin/End`). Apply with `make migrate-up`.
10. **i18n.** Default language English; detection is Accept-Language only (no cookie). Every user-facing error/message has a key in `locales/en.json`, `uz.json`, `ru.json`.
11. **Swagger.** Annotate every handler (`@Summary`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router`, `@Security BearerAuth` for protected). Regenerate with `make swag` (uses `--parseInternal` only → short `pkg.Type` model names; do **not** add `--parseDependency`).
12. **Logging.** App logs go through `slog` (default set in `server.New`). Dev = colorful console, prod = JSON. Don't use the std `log` package in app code.

## Recipe: add a feature (e.g. a new resource)

1. `domain`: entity struct (`db:`/`json:` tags) + repository interface + `Err*`.
2. `migrations`: goose SQL; `make migrate-up`.
3. `repository`: implement the interface (squirrel + sqlx, soft-delete aware).
4. `dto`: request/response structs (camelCase, `validate` tags) + `NewXResponse` mapper.
5. `usecase`: business logic depending on the domain interface.
6. `handler`: bind → validate → usecase → `response.*`; full Swagger annotations.
7. `routes`: register under a group with the right auth/role middleware.
8. `wire`: add providers + `provideHandlers`/`Handlers`; `make wire`.
9. `apperror` + locales: map any new domain errors; add i18n keys to en/uz/ru.
10. Verify: `go build ./...`, `go vet ./...`, `make swag`. Drive endpoints in-process with `httptest` against `srv.Handler()` (the sandbox blocks listening servers; never rely on a bound port).

## Gotchas
- The sandbox kills processes that bind a port (and blocks foreground `sleep`). Verify by building the wired server and exercising `srv.Handler()` with `net/http/httptest` in a throwaway `cmd/*` driver, then delete it.
- `apperror` must not import `validation`/`response` (cycle). `response` imports only echo.
- After editing handler Swagger types, always rerun `make swag` and confirm `docs/swagger.json` definitions are `dto.*`/`response.*` only — no `domain.*`.
