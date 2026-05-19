# Plan

## Context

The user module currently only exposes data operations to the auth module. There are no user-facing API endpoints. This task adds the first user-facing endpoint — a profile update API — which requires completing the user module (handler, service, routes) and also completing the currently-stub `RequireAuth` middleware to perform real JWT validation.

## Objective

Implement `PATCH /api/v1/user/me` — a protected endpoint that reads the caller's user ID from their access token and updates their `name` field. Name must contain at least one letter (no whitespace-only values). User ID is never accepted from the request body.

## Key Changes

### 1. Complete `RequireAuth` middleware (`pkg/middleware/auth.go`)
- Define a local `TokenParser` interface (`ParseAccessToken(string) (uint64, error)`) to avoid concrete infra dependency.
- Replace the stub with a real factory `RequireAuth(parser TokenParser) gin.HandlerFunc` that extracts the Bearer token, validates it, and sets `ContextKeyUserID = "user_id"` in gin context; returns 401 with `INVALID_ACCESS_TOKEN` on any failure.
- Add local `extractBearerToken` helper (the one in `auth/handler.go` is unexported and in a different package).

### 2. Extend user repository (`internal/modules/user/repository.go`)
- Add `FindByID(ctx, id) (*User, error)` — queries by primary key, maps `gorm.ErrRecordNotFound` → `ErrUserNotFound`.
- Add `UpdateName(ctx, userID, name) (*User, error)` — issues a targeted UPDATE, returns `ErrUserNotFound` if `RowsAffected == 0`, then returns the updated user via `FindByID`.

### 3. Add `ErrInvalidName` to user errors (`internal/modules/user/errors.go`)
- Register mapper: `ErrInvalidName` → `CodeInvalidRequest` / HTTP 400 / message `"invalid name"`.

### 4. Add DTOs (`internal/modules/user/dto.go`)
- `UpdateProfileRequest { Name string \`json:"name" binding:"required"\` }`
- `UpdateProfileResponse { Name string \`json:"name"\` }`

### 5. Add user Service (`internal/modules/user/service.go`)
- Define `Service` interface: `UpdateName(ctx, userID uint64, req UpdateProfileRequest) (*UpdateProfileResponse, error)`.
- Define local `userRepository` interface (only `UpdateName` + `FindByID`) to keep dependency clean.
- `service.UpdateName` trims the name, validates it (non-empty after trim; contains at least one unicode letter; only letters/spaces/hyphens/apostrophes allowed) — returns `ErrInvalidName` on violation. On valid input, calls repo `UpdateName` and returns `UpdateProfileResponse`.

### 6. Add user Handler (`internal/modules/user/handler.go`)
- `Handler { service Service }`, `NewHandler(service Service) *Handler`.
- `UpdateProfile(c *gin.Context)`: reads `ContextKeyUserID` from context (401 if missing), binds JSON body (400 on bind error), calls `service.UpdateName`, returns 200 with `response.OK`.

### 7. Add user Routes (`internal/modules/user/routes.go`)
- `RegisterRoutes(rg *gin.RouterGroup, h *Handler)` — creates `/user` sub-group, registers `PATCH /me`.

### 8. Wire bootstrap (`internal/bootstrap/dependencies.go` + `router.go`)
- `dependencies.go`: add `UserHandler *user.Handler` and `TokenProvider tokenprovider.Provider` to `Dependencies`; build `userSvc := user.NewService(userRepo)` and `userHandler := user.NewHandler(userSvc)`; assign both to the returned struct.
- `router.go`: after `auth.RegisterRoutes`, add:
  ```go
  protected := api.Group("")
  protected.Use(middleware.RequireAuth(deps.TokenProvider))
  user.RegisterRoutes(protected, deps.UserHandler)
  ```
  The `protected` group inherits device-context enforcement from the parent `api` group.

### 9. Documentation
- `docs/api/user.postman_collection.json`: add `PATCH /api/v1/user/me` item with auth header, device-context headers, request/response/error examples.
- `docs/modules/user.md`: add endpoint flow — route entry → device-context middleware → RequireAuth middleware → `Handler.UpdateProfile` → `Service.UpdateName` (validation, repo call) → `Repository.UpdateName` → response branches.
- `docs/knowledge-base.md`: record that `RequireAuth` is now implemented (not a stub), and that user profile update is the first user-facing API endpoint.

### 10. Tests
- `tests/unit/user/handler_test.go`: handler tests using a fake service — success, missing auth context, invalid JSON body.
- `tests/unit/user/service_test.go`: service tests using a fake repo — success, whitespace-only name (→ ErrInvalidName), empty name (→ ErrInvalidName), invalid characters (→ ErrInvalidName), repo error propagation.

## Files Impacted

| File | Action |
|---|---|
| `pkg/middleware/auth.go` | Modify — implement real RequireAuth |
| `internal/modules/user/repository.go` | Modify — add FindByID, UpdateName |
| `internal/modules/user/errors.go` | Modify — add ErrInvalidName |
| `internal/modules/user/dto.go` | Modify — add request/response DTOs |
| `internal/modules/user/service.go` | Create — Service interface + implementation |
| `internal/modules/user/handler.go` | Create — Handler + UpdateProfile |
| `internal/modules/user/routes.go` | Create — RegisterRoutes |
| `internal/bootstrap/dependencies.go` | Modify — wire UserHandler + TokenProvider |
| `internal/bootstrap/router.go` | Modify — add protected group + user routes |
| `docs/api/user.postman_collection.json` | Modify — add endpoint item |
| `docs/modules/user.md` | Modify — add endpoint flow |
| `docs/knowledge-base.md` | Modify — record decisions |
| `tests/unit/user/handler_test.go` | Create — handler unit tests |
| `tests/unit/user/service_test.go` | Create — service unit tests |

## Execution Steps

1. Implement `RequireAuth` in `pkg/middleware/auth.go` (define `TokenParser` interface + `ContextKeyUserID` constant).
2. Add `FindByID` and `UpdateName` to `internal/modules/user/repository.go`.
3. Add `ErrInvalidName` to `internal/modules/user/errors.go`.
4. Update `internal/modules/user/dto.go` with request/response structs.
5. Create `internal/modules/user/service.go` with validation and repo delegation.
6. Create `internal/modules/user/handler.go` reading from context + binding + service call.
7. Create `internal/modules/user/routes.go`.
8. Update `internal/bootstrap/dependencies.go` to wire `UserHandler` and expose `TokenProvider`.
9. Update `internal/bootstrap/router.go` to add protected group and register user routes.
10. Update `docs/api/user.postman_collection.json`.
11. Update `docs/modules/user.md` and `docs/knowledge-base.md`.
12. Create `tests/unit/user/handler_test.go` and `tests/unit/user/service_test.go`.
13. Run `gofmt ./...`, `go vet ./...`, `go test ./...`, `make build`, `make graphify-update`.

## Risks / Notes

- **`extractBearerToken` duplication**: A similar function exists in `auth/handler.go` (unexported). The middleware needs its own copy because it lives in a different package (`pkg/middleware`) with no access to auth internals. This is acceptable — both are small, pure helpers.
- **GORM `Update` + `updated_at`**: GORM auto-manages `updated_at` on the `SoftDeleteableModel`. The targeted `Model(&User{}).Where("id=?").Update("name")` call will also set `updated_at` automatically.
- **Name regex scope**: Validation allows unicode letters + spaces + hyphens + apostrophes and requires at least one letter. This covers most international names without adding a heavy dependency.
- **Token expiry on protected route**: `ParseAccessToken` in the JWT provider validates expiry, so expired tokens will fail at the middleware layer automatically.

## Definition of Done

- `PATCH /api/v1/user/me` returns 200 with updated name when valid JWT + valid name provided.
- Returns 401 (`INVALID_ACCESS_TOKEN`) when Authorization header is missing/invalid.
- Returns 400 (`INVALID_REQUEST`) when body is malformed.
- Returns 400 (`INVALID_REQUEST` / message `invalid name`) when name is empty, whitespace-only, or contains invalid characters.
- Returns 400 (`INVALID_DEVICE_CONTEXT`) when device-context headers are missing/invalid (inherited from parent middleware).
- All unit tests pass (`go test ./...`).
- Build passes (`make build`).
- Postman collection and module docs updated.
