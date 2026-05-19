# Plan

## Context

The `users` table has a nullable `name` column but no API to set it. A user completing OTP auth gets a session with a JWT. This task adds a protected profile-update endpoint so authenticated users can set their display name. The previous execution session already added `CodeInvalidName` to `pkg/errors/codes.go`; `pkg/middleware/auth.go` still has a stub that always returns 401.

## Key Changes

1. **`pkg/middleware/auth.go`** — Replace stub with a real `RequireAuth(parser TokenParser)` middleware. Define a local `TokenParser` interface (`ParseAccessToken(string) (uint64, error)`) and an exported `ContextKeyUserID = "userID"` constant. Extracts `Authorization: Bearer <token>`, parses userID, stores it in gin context, returns `401 INVALID_ACCESS_TOKEN` on failure.

2. **`internal/modules/user/errors.go`** — Add `ErrInvalidName` sentinel and register mapper → `CodeInvalidName`, HTTP 422.

3. **`internal/modules/user/dto.go`** — Add `UpdateProfileRequest { Name string binding:"required" }` and `UpdateProfileResponse { ID uint64; Name string }`.

4. **`internal/modules/user/repository.go`** — Add `UpdateName(ctx, userID uint64, name string) error` using GORM `Model(&User{ID: userID}).Update("name", name)` with a `RowsAffected == 0` guard returning `ErrUserNotFound`.

5. **`internal/modules/user/service.go`** — New file. Define `Service` interface with `UpdateProfile(ctx, userID uint64, req UpdateProfileRequest) (*UpdateProfileResponse, error)`. Private `userRepo` interface only exposes `UpdateName`. Name validation: trim → non-empty → at least one Unicode letter (`\p{L}`) → only valid chars (`^[\p{L}\s'\-]+$`); violations return `ErrInvalidName`.

6. **`internal/modules/user/handler.go`** — New file. `Handler { service Service }`. `UpdateProfile`: reads `userID` from context (`c.GetUint64(middleware.ContextKeyUserID)`), binds JSON body, calls service, responds `200 OK` with `UpdateProfileResponse`.

7. **`internal/modules/user/routes.go`** — New file. `RegisterRoutes(rg *gin.RouterGroup, h *Handler)` registers `PATCH /user/me` → `h.UpdateProfile`.

8. **`internal/bootstrap/dependencies.go`** — Add `UserHandler *user.Handler` and `TokenProvider tokenprovider.Provider` to `Dependencies`. Wire `user.NewService(userRepo)` → `user.NewHandler(svc)` → assign to deps.

9. **`internal/bootstrap/router.go`** — After auth routes, create a protected sub-group: `protected := api.Group(""); protected.Use(middleware.RequireAuth(deps.TokenProvider))`. Call `user.RegisterRoutes(protected, deps.UserHandler)`. Also import the user module package.

10. **Tests** — `tests/unit/user/handler_test.go`: test missing body (400), missing userID in context (400/401), valid call (200). `tests/unit/user/service_test.go`: test empty name, spaces-only, numbers-only, valid name, name with hyphen/apostrophe.

11. **Docs** — Update `docs/api/user.postman_collection.json` with `PATCH /api/v1/user/me` item. Update `docs/modules/user.md` with endpoint flow. Update `docs/knowledge-base.md` with new decisions.

## Files Impacted

| File | Action |
|---|---|
| `pkg/errors/codes.go` | Already done — `CodeInvalidName` added |
| `pkg/middleware/auth.go` | Modify — implement real RequireAuth |
| `internal/modules/user/errors.go` | Modify — add ErrInvalidName + mapper |
| `internal/modules/user/dto.go` | Modify — add DTOs |
| `internal/modules/user/repository.go` | Modify — add UpdateName |
| `internal/modules/user/service.go` | Create |
| `internal/modules/user/handler.go` | Create |
| `internal/modules/user/routes.go` | Create |
| `internal/bootstrap/dependencies.go` | Modify — wire user module, expose TokenProvider |
| `internal/bootstrap/router.go` | Modify — add protected group + user routes |
| `tests/unit/user/handler_test.go` | Create |
| `tests/unit/user/service_test.go` | Create |
| `docs/api/user.postman_collection.json` | Modify — add endpoint |
| `docs/modules/user.md` | Modify — add endpoint flow |
| `docs/knowledge-base.md` | Modify — add decisions |

## Execution Steps

1. Implement `RequireAuth` + `TokenParser` interface + `ContextKeyUserID` in `pkg/middleware/auth.go`
2. Add `ErrInvalidName` + mapper to `internal/modules/user/errors.go`
3. Update `internal/modules/user/dto.go` with request/response DTOs
4. Add `UpdateName` to `internal/modules/user/repository.go`
5. Create `internal/modules/user/service.go` with Service interface, local repo interface, and name validation (package-level compiled regexps)
6. Create `internal/modules/user/handler.go`
7. Create `internal/modules/user/routes.go`
8. Update `internal/bootstrap/dependencies.go` to wire user module and expose `TokenProvider`
9. Update `internal/bootstrap/router.go` to add protected group with user routes
10. Write unit tests in `tests/unit/user/handler_test.go` and `tests/unit/user/service_test.go`
11. Update docs: `docs/modules/user.md`, `docs/api/user.postman_collection.json`, `docs/knowledge-base.md`
12. Run `gofmt ./...`, `go vet ./...`, `go test ./...`, `make build`, `make graphify-update`

## Risks / Notes

- `pkg/middleware/auth.go` changing from no-arg stub to `RequireAuth(parser)` is a breaking signature change; `contract_test.go` in auth does not call `RequireAuth` so no existing test breaks.
- The `TokenProvider` field on `Dependencies` exposes `tokenprovider.Provider` (the interface), not the concrete JWT type, keeping dependency direction clean.
- `UpdateName` uses a soft-delete-aware GORM query via the `User` model (GORM respects `deleted_at` in its `Model()` + `Update()` path automatically through the embedded `SoftDeleteableModel`).
- Name regex uses `\p{L}` (Unicode letter class supported by Go's `regexp` RE2 engine).

## Definition of Done

- `PATCH /api/v1/user/me` returns `200` with updated name for a valid authenticated request.
- Missing/invalid Bearer token → `401 INVALID_ACCESS_TOKEN`.
- Missing/invalid device-context headers → `400 INVALID_DEVICE_CONTEXT` (inherited from existing middleware).
- Invalid name (empty, no letters, bad chars) → `422 INVALID_NAME`.
- `go test ./...` passes with new unit tests covering handler and service validation.
- `make build` succeeds.
- Postman collection and module docs updated.
