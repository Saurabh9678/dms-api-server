# Plan

## Objective

Create a `PATCH /api/v1/users/me` endpoint that lets an authenticated user update their name. User ID is taken from the `Authorization: Bearer <token>` JWT. Only the `name` field is accepted in the request body. The name must be a non-empty valid name string (at least one alphabetic word; may contain letters, spaces, hyphens, apostrophes).

---

## Key Changes

### 1. Fix `pkg/middleware/auth.go` — implement `RequireAuth` as a factory

The current stub always returns 401 without inspecting the token. Replace it with a factory function accepting a `TokenParser` interface (defined in the same file) so `pkg` stays free of `internal` imports.

```go
type TokenParser interface {
    ParseAccessToken(token string) (uint64, error)
}

const ContextKeyUserID = "user_id"

func RequireAuth(parser TokenParser) gin.HandlerFunc { ... }
func GetUserID(c *gin.Context) (uint64, bool)        { ... }  // helper
```

`RequireAuth` extracts the `Bearer` token, calls `parser.ParseAccessToken`, sets `ContextKeyUserID` in gin context, then calls `c.Next()`. On any failure it aborts with `401 INVALID_ACCESS_TOKEN`.

### 2. Add error code to `pkg/errors/codes.go`

Add `CodeInvalidName = "INVALID_NAME"` for name validation failures.

### 3. Update `internal/modules/user/errors.go`

Add `ErrInvalidName` sentinel error and an `init()` mapper → `400 INVALID_NAME "invalid name"`.

### 4. Update `internal/modules/user/dto.go`

Add:
```go
type UpdateProfileRequest struct {
    Name string `json:"name" binding:"required"`
}

type UpdateProfileResponse struct {
    Name string `json:"name"`
}
```

### 5. Update `internal/modules/user/repository.go`

Add two methods:
- `FindByID(ctx, id uint64) (*User, error)` — returns `ErrUserNotFound` on miss.
- `UpdateName(ctx, id uint64, name string) error` — uses `db.Model(&User{}).Where("id = ?", id).Update("name", name)`.

### 6. Create `internal/modules/user/service.go`

Define `Service` interface + `service` struct:

```go
type Service interface {
    UpdateProfile(ctx context.Context, userID uint64, req UpdateProfileRequest) (*UpdateProfileResponse, error)
}

type userRepo interface {
    FindByID(ctx context.Context, id uint64) (*User, error)
    UpdateName(ctx context.Context, id uint64, name string) error
}
```

`UpdateProfile` logic:
1. `strings.TrimSpace(req.Name)` → check not empty and matches `^[A-Za-z][A-Za-z\s'\-.]*$` → return `ErrInvalidName` if invalid.
2. `FindByID` to confirm user exists (handles soft-deleted tokens).
3. `UpdateName`.
4. Return `&UpdateProfileResponse{Name: trimmedName}`.

### 7. Create `internal/modules/user/handler.go`

```go
type Handler struct { service Service }

func (h *Handler) UpdateProfile(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    // bind JSON
    // call service
    // response.OK(c, "Profile updated successfully", resp)
}
```

### 8. Create `internal/modules/user/routes.go`

```go
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
    rg.Group("/users").PATCH("/me", h.UpdateProfile)
}
```

### 9. Update `internal/bootstrap/dependencies.go`

- Add `UserHandler *user.Handler` and `TokenProvider tokenprovider.Provider` to `Dependencies`.
- Wire `user.NewService(userRepo)` → `user.NewHandler(...)` → store in `Dependencies`.
- Expose `tokenProvider` in `Dependencies` so `router.go` can pass it to `RequireAuth`.

### 10. Update `internal/bootstrap/router.go`

- Create a protected sub-group: `protected := api.Group(""); protected.Use(middleware.RequireAuth(deps.TokenProvider))`.
- Register: `user.RegisterRoutes(protected, deps.UserHandler)`.

### 11. Add unit tests in `tests/unit/user/`

Create `tests/unit/user/service_test.go` and `tests/unit/user/handler_test.go`:

**Service tests** (fake `userRepo`):
- Empty name → `ErrInvalidName`.
- Whitespace-only name → `ErrInvalidName`.
- Name starting with digit/symbol → `ErrInvalidName`.
- Valid single-word name → success.
- Valid multi-word name → success.
- User not found → `ErrUserNotFound`.

**Handler tests** (fake service + gin test engine with `ContextKeyUserID` pre-set):
- Missing `user_id` in context → 401.
- Missing/invalid body → 400 `INVALID_REQUEST`.
- Service returns `ErrInvalidName` → 400 `INVALID_NAME`.
- Success → 200 with name in data.

### 12. Documentation updates

- `docs/modules/user.md` — add endpoint flow for `PATCH /api/v1/users/me`.
- `docs/api/user.postman_collection.json` — add Update Profile item with auth headers, request/response examples, and error examples.
- `docs/knowledge-base.md` — record the new user profile update endpoint and `RequireAuth` middleware wiring.

---

## Files Impacted

| File | Action |
|------|--------|
| `pkg/middleware/auth.go` | Replace stub with `TokenParser` interface + `RequireAuth` factory + `GetUserID` helper |
| `pkg/errors/codes.go` | Add `CodeInvalidName` |
| `internal/modules/user/errors.go` | Add `ErrInvalidName` + mapper |
| `internal/modules/user/dto.go` | Add `UpdateProfileRequest`, `UpdateProfileResponse` |
| `internal/modules/user/repository.go` | Add `FindByID`, `UpdateName` |
| `internal/modules/user/service.go` | **NEW** — `Service` interface + implementation |
| `internal/modules/user/handler.go` | **NEW** — `Handler` with `UpdateProfile` |
| `internal/modules/user/routes.go` | **NEW** — `RegisterRoutes` |
| `internal/bootstrap/dependencies.go` | Wire user svc/handler, expose `TokenProvider` |
| `internal/bootstrap/router.go` | Protected sub-group + user route registration |
| `tests/unit/user/service_test.go` | **NEW** — service unit tests |
| `tests/unit/user/handler_test.go` | **NEW** — handler unit tests |
| `docs/modules/user.md` | Add endpoint flow |
| `docs/api/user.postman_collection.json` | Add endpoint doc |
| `docs/knowledge-base.md` | Record new API and middleware wiring |

---

## Execution Steps

1. Update `pkg/middleware/auth.go` (TokenParser interface, RequireAuth factory, GetUserID helper).
2. Add `CodeInvalidName` to `pkg/errors/codes.go`.
3. Update `internal/modules/user/errors.go` (ErrInvalidName + mapper).
4. Update `internal/modules/user/dto.go`.
5. Update `internal/modules/user/repository.go` (FindByID, UpdateName).
6. Create `internal/modules/user/service.go`.
7. Create `internal/modules/user/handler.go`.
8. Create `internal/modules/user/routes.go`.
9. Update `internal/bootstrap/dependencies.go`.
10. Update `internal/bootstrap/router.go`.
11. Create `tests/unit/user/service_test.go` and `tests/unit/user/handler_test.go`.
12. Update all docs.
13. Run validation: `gofmt ./... && go vet ./... && go test ./... && make build`.

---

## Risks / Notes

- `pkg/middleware/auth.go` currently has a stub that **any existing code relying on `RequireAuth()` (no args)** would break. A quick grep shows it is not yet called anywhere in routes — so the signature change is safe.
- `TokenProvider tokenprovider.Provider` added to `Dependencies` exposes an `internal/` type in `bootstrap`. This is acceptable because `bootstrap` is itself `internal/bootstrap` — it's the composition root, not a public API.
- The name regex `^[A-Za-z][A-Za-z\s'\-.]*$` intentionally rejects names starting with a space, digit, or symbol. Unicode names (e.g. accented chars) are excluded for now as requirements do not specify. This can be relaxed later.
- No DB migration is needed — `name` column already exists and is nullable.

---

## Definition of Done

- [ ] `PATCH /api/v1/users/me` returns `200` with updated name when called with a valid token and valid name.
- [ ] Returns `401 INVALID_ACCESS_TOKEN` if token is missing or invalid.
- [ ] Returns `400 INVALID_DEVICE_CONTEXT` if X-Platform/X-Device-Id are missing/invalid.
- [ ] Returns `400 INVALID_NAME` if name is empty, whitespace-only, or fails regex.
- [ ] Returns `404 USER_NOT_FOUND` if the token user no longer exists.
- [ ] All unit tests pass.
- [ ] `go vet ./...`, `go test ./...`, `make build` all pass.
- [ ] Docs updated: `docs/modules/user.md`, `docs/api/user.postman_collection.json`, `docs/knowledge-base.md`.
