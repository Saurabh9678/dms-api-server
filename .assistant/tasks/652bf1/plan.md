# Plan

## Context

The user module currently owns the User model and repository (FindByPhone, Create) but has no API surface. A profile update API is needed so authenticated users can set their display name. The name field already exists in the `users` table as a nullable VARCHAR(100). The `RequireAuth` middleware in `pkg/middleware/auth.go` is a stub (always 401); it must be replaced with a real implementation before any protected route can work.

## Objective

Create `PATCH /api/v1/users/profile` — protected, JWT-authenticated, device-context-enforced — that accepts `{ "name": "<value>" }` (at least one non-whitespace word), takes userID from the access token, updates `users.name`, and returns the updated profile.

## Key Changes

### 1. Fix `pkg/middleware/auth.go`
- Define a local `TokenParser` interface (`ParseAccessToken(string) (uint64, error)`) — avoids importing `internal` from `pkg`.
- Export `UserIDKey = "user_id"` constant for context storage.
- Replace the stub `RequireAuth()` with `RequireAuth(TokenParser) gin.HandlerFunc` that extracts the Bearer token, parses it, stores userID in context, and returns `INVALID_ACCESS_TOKEN` / 401 on failure.

### 2. Extend `internal/modules/user/repository.go`
- Add `FindByID(ctx, id uint64) (*User, error)` — returns `ErrUserNotFound` on miss.
- Add `UpdateName(ctx, userID uint64, name string) error` — GORM model-scoped update; returns `ErrUserNotFound` if RowsAffected == 0.

### 3. Update `internal/modules/user/dto.go`
- Add `UpdateProfileRequest { Name string \`json:"name" binding:"required"\` }`.
- Add `ProfileResponse { ID uint64, Name string, PhoneNumber string, CountryCode string }`.

### 4. Create `internal/modules/user/service.go`
- Define `Service` interface: `UpdateProfile(ctx, userID uint64, req UpdateProfileRequest) (*ProfileResponse, error)`.
- Implement `service` struct backed by a `userRepo` interface (FindByID, UpdateName).
- Business logic: trim name, reject if empty (return `apperrors.NewAppError(CodeInvalidRequest, "invalid request", 400, nil)`), call `UpdateName`, then `FindByID` to return the updated record.

### 5. Create `internal/modules/user/handler.go`
- `Handler` struct holding `Service`.
- `UpdateProfile(c *gin.Context)`: read userID from `c.MustGet(middleware.UserIDKey).(uint64)`, bind body, call service, respond with `response.OK`.

### 6. Create `internal/modules/user/routes.go`
- `RegisterRoutes(group *gin.RouterGroup, h *Handler)` mounting `PATCH /profile` on a `/users` sub-group.

### 7. Update `internal/bootstrap/dependencies.go`
- Expose `TokenProvider tokenprovider.Provider` on the `Dependencies` struct.
- Build `user.NewService(userRepo)` and `user.NewHandler(userSvc)`, add `UserHandler *user.Handler` to `Dependencies`.

### 8. Update `internal/bootstrap/router.go`
- Create a protected sub-group: `protected := api.Group(""); protected.Use(middleware.RequireAuth(deps.TokenProvider))`.
- Call `user.RegisterRoutes(protected, deps.UserHandler)` on the protected group.

### 9. Documentation
- `docs/api/user.postman_collection.json` — add the `PATCH /api/v1/users/profile` item (auth, headers, request/response examples).
- `docs/modules/user.md` — add endpoint flow: route entry → device-context middleware → auth middleware → handler → service → repository → response.
- `docs/knowledge-base.md` — record: profile update endpoint added, RequireAuth wired with TokenParser interface, name trimmed server-side.

### 10. Tests
- `tests/unit/user/handler_test.go` — table-driven tests: missing auth (401), empty name (400), whitespace-only name (400), valid name (200).
- `tests/unit/user/service_test.go` — test UpdateProfile with stub repo: empty name returns error, valid name calls UpdateName + FindByID.

## Files Impacted

| Action | File |
|--------|------|
| Modify | `pkg/middleware/auth.go` |
| Modify | `internal/modules/user/repository.go` |
| Modify | `internal/modules/user/dto.go` |
| Create | `internal/modules/user/service.go` |
| Create | `internal/modules/user/handler.go` |
| Create | `internal/modules/user/routes.go` |
| Modify | `internal/bootstrap/dependencies.go` |
| Modify | `internal/bootstrap/router.go` |
| Modify | `docs/api/user.postman_collection.json` |
| Modify | `docs/modules/user.md` |
| Modify | `docs/knowledge-base.md` |
| Create | `tests/unit/user/handler_test.go` |
| Create | `tests/unit/user/service_test.go` |

## Execution Steps

1. **`pkg/middleware/auth.go`** — define `TokenParser` interface, `UserIDKey` const, implement `RequireAuth(TokenParser)`.
2. **`internal/modules/user/repository.go`** — add `FindByID` and `UpdateName`.
3. **`internal/modules/user/dto.go`** — add `UpdateProfileRequest` and `ProfileResponse`.
4. **`internal/modules/user/service.go`** — define `Service` interface + `service` implementation.
5. **`internal/modules/user/handler.go`** — implement `UpdateProfile` handler.
6. **`internal/modules/user/routes.go`** — register `PATCH /users/profile`.
7. **`internal/bootstrap/dependencies.go`** — wire user service + handler, expose `TokenProvider`.
8. **`internal/bootstrap/router.go`** — add protected group, register user routes.
9. **Tests** — create handler + service unit tests.
10. **Docs** — update Postman collection, module doc, knowledge base.
11. **Validate** — run `gofmt ./...`, `go vet ./...`, `go test ./...`, `make build`, `make graphify-update`.

## Risks / Notes

- **RequireAuth change is a breaking replacement** of the stub — the old stub was always-401, so the change is safe (no existing protected routes).
- **Name trimming is server-side** — the stored name is trimmed; the client should be aware.
- **No DB migration needed** — `name` column already exists as nullable VARCHAR(100) in the `users` table.
- **`tokenprovider.Provider` satisfies `TokenParser`** structurally — no explicit declaration needed in infra layer.
- The service calls `UpdateName` followed by `FindByID`; this is two DB round-trips. Acceptable for a low-frequency profile update.

## Definition of Done

- `PATCH /api/v1/users/profile` with a valid Bearer token, device-context headers, and `{"name": "John"}` returns 200 with the updated profile.
- Missing or invalid Bearer token returns 401 with `INVALID_ACCESS_TOKEN`.
- Empty or whitespace-only name returns 400 with `INVALID_REQUEST`.
- `go test ./...` passes.
- `make build` succeeds.
- `docs/api/user.postman_collection.json` contains the new endpoint.
