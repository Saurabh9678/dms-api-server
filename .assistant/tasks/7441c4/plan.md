# Plan

## Context

The `users` table has a `name` field (`VARCHAR(100)`, nullable). No user-facing API exists yet — the user module only has a model, repository (FindByPhone, Create), and empty DTOs. The `RequireAuth` middleware in `pkg/middleware/auth.go` is a stub that always returns 401. This task wires the token-based auth guard and creates the first protected user endpoint.

## Objective

Implement `PATCH /api/v1/user/me` — a protected endpoint that reads user ID from the JWT access token and updates the `name` field. Name must be non-empty after trim and match valid name characters (Unicode letters, spaces, hyphens, apostrophes).

## Key Changes

### 1. `pkg/middleware/auth.go` — Implement RequireAuth
- Define a local `TokenParser` interface (`ParseAccessToken(string) (uint64, error)`) so `pkg/` stays clean from `internal/` imports.
- Export `ContextKeyUserID = "userID"` constant for handler use.
- `RequireAuth(parser TokenParser)` extracts the Bearer token, calls `parser.ParseAccessToken`, sets `userID` in context, or aborts with 401 / `INVALID_ACCESS_TOKEN`.
- Reuse the inline Bearer extraction logic (same pattern as in `auth/handler.go`).

### 2. `internal/modules/user/dto.go` — Add DTOs
- `UpdateProfileRequest { Name string \`json:"name" binding:"required"\` }`
- `UpdateProfileResponse { Name string \`json:"name"\` }`

### 3. `internal/modules/user/repository.go` — Add UpdateName
- `UpdateName(ctx context.Context, userID uint64, name string) error`
- GORM: `Model(&User{}).Where("id = ?", userID).Update("name", name)`
- Return `ErrUserNotFound` if `RowsAffected == 0`.

### 4. `internal/modules/user/service.go` (new file)
- Local `profileRepo` interface: `UpdateName(ctx, userID uint64, name string) error`
- `Service` interface: `UpdateProfile(ctx, userID uint64, req UpdateProfileRequest) (*UpdateProfileResponse, error)`
- Validation in implementation: trim name → check non-empty → regexp `^[\p{L}\s''-]+$` → call repo → return `UpdateProfileResponse{Name: trimmedName}`
- Return `apperrors.NewAppError(CodeInvalidRequest, "invalid request", 400, nil)` on validation failure.

### 5. `internal/modules/user/handler.go` (new file)
- Extract `userID` from `c.Get(middleware.ContextKeyUserID)` → assert as `uint64`.
- Bind JSON body; invalid binding → 400 `INVALID_REQUEST`.
- Call `h.service.UpdateProfile`; propagate error via `response.FromError`.
- Success: `response.OK(c, "profile updated", resp)`.

### 6. `internal/modules/user/routes.go` (new file)
- `RegisterRoutes(group *gin.RouterGroup, h *Handler)` registers `PATCH /user/me`.

### 7. `internal/bootstrap/dependencies.go` — Wire user handler
- Add `UserHandler *user.Handler` and `TokenProvider tokenprovider.Provider` fields to `Dependencies`.
- In `buildDependencies`: construct `userSvc := user.NewService(userRepo)` and `userHandler := user.NewHandler(userSvc)`; also expose `tokenProvider`.

### 8. `internal/bootstrap/router.go` — Register protected routes
- Create protected sub-group: `protected := api.Group(""); protected.Use(middleware.RequireAuth(deps.TokenProvider))`
- Call `user.RegisterRoutes(protected, deps.UserHandler)` on the protected group.

### 9. Tests
- `tests/unit/user/handler_test.go`: fake service, test valid update (200), empty name (400), missing userID in context (401), service error propagation.
- `tests/unit/user/service_test.go`: fake repo, test valid name (calls UpdateName, returns response), empty name (error), invalid chars (error), repo error propagation.

### 10. Documentation
- `docs/api/user.postman_collection.json` — Add `PATCH /api/v1/user/me` item with auth, device-context headers, request/response/error examples.
- `docs/modules/user.md` — Add endpoint flow: route entry → `RequireDeviceContext` → `RequireAuth` → `UpdateProfile` handler → service validation → repo → response branches.
- `docs/knowledge-base.md` — Record new endpoint, auth middleware implementation, name validation convention.

## Files Impacted

| File | Action |
|------|--------|
| `pkg/middleware/auth.go` | Modify — implement RequireAuth, add TokenParser interface, export ContextKeyUserID |
| `internal/modules/user/dto.go` | Modify — add UpdateProfileRequest, UpdateProfileResponse |
| `internal/modules/user/repository.go` | Modify — add UpdateName method |
| `internal/modules/user/service.go` | Create — Service interface + implementation with name validation |
| `internal/modules/user/handler.go` | Create — Handler with UpdateProfile |
| `internal/modules/user/routes.go` | Create — RegisterRoutes |
| `internal/bootstrap/dependencies.go` | Modify — add UserHandler, TokenProvider fields and wiring |
| `internal/bootstrap/router.go` | Modify — add protected group, register user routes |
| `tests/unit/user/handler_test.go` | Create — handler unit tests |
| `tests/unit/user/service_test.go` | Create — service unit tests |
| `docs/api/user.postman_collection.json` | Modify — add endpoint documentation |
| `docs/modules/user.md` | Modify — add endpoint flow |
| `docs/knowledge-base.md` | Modify — record new decisions |

## Existing Functions to Reuse

- `response.OK`, `response.Error`, `response.FromError` — `pkg/response/`
- `apperrors.NewAppError`, `apperrors.CodeInvalidRequest`, `apperrors.CodeInvalidAccessToken` — `pkg/errors/`
- `tokenprovider.Provider.ParseAccessToken` — `internal/providers/token/provider.go` (already used in `auth/service.go`)
- `user.NewRepository(db)` — already wired in `bootstrap/dependencies.go`
- `ErrUserNotFound` mapper registration — already in `internal/modules/user/errors.go` (auto-maps to 404 USER_NOT_FOUND)

## Execution Steps

1. Update `pkg/middleware/auth.go` — implement RequireAuth with TokenParser interface and ContextKeyUserID constant.
2. Update `internal/modules/user/dto.go` — add request/response DTOs.
3. Update `internal/modules/user/repository.go` — add `UpdateName` method.
4. Create `internal/modules/user/service.go` — Service interface + validated implementation.
5. Create `internal/modules/user/handler.go` — Handler wiring context userID → service.
6. Create `internal/modules/user/routes.go` — route registration.
7. Update `internal/bootstrap/dependencies.go` — expose TokenProvider, wire UserHandler.
8. Update `internal/bootstrap/router.go` — create protected group, register user routes.
9. Create `tests/unit/user/handler_test.go` and `service_test.go`.
10. Update `docs/api/user.postman_collection.json`, `docs/modules/user.md`, `docs/knowledge-base.md`.
11. Run `gofmt ./...`, `go vet ./...`, `go test ./...`, `make build`, `make graphify-update`.

## Risks / Notes

- **RequireAuth stub change**: The current `RequireAuth()` takes no arguments. Callers (currently none in production code — only `pkg/middleware/auth.go` defines it) must be updated to pass the `TokenParser`. Existing smoke/unit tests don't call `RequireAuth` directly, so no test breakage.
- **No migration needed**: The `name` column already exists and is `VARCHAR(100)` nullable. The `UpdateName` method uses GORM's `Update` (not `Save`), so `updated_at` will be set correctly via GORM's auto-update.
- **Name validation scope**: Validation lives in the service layer, not the binding tag, per existing pattern (binding handles required/format; service handles business rules).
- **`tokenprovider.Provider` in Dependencies**: Adding it as a field is the minimal change — it's already constructed locally in `buildDependencies`; just needs to be surfaced to `newRouter`.

## Definition of Done

- `PATCH /api/v1/user/me` returns 200 with `{"success": true, "message": "profile updated", "data": {"name": "<name>"}}` on valid request with a valid JWT.
- Returns 400 `INVALID_REQUEST` for empty/blank name or invalid name characters.
- Returns 401 `INVALID_ACCESS_TOKEN` for missing or invalid Bearer token.
- Returns 400 `INVALID_DEVICE_CONTEXT` for missing/invalid `X-Platform` or `X-Device-Id` (inherited from group middleware).
- All unit tests pass (`go test ./...`).
- Build passes (`make build`).
- Postman collection and module docs updated.
