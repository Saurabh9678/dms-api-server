# Plan

## Context

The user module exists with a `User` model that has a `name` field, but has no API surface yet. The task is to add a protected profile update endpoint that lets the authenticated user update their name. The `RequireAuth` middleware in `pkg/middleware/auth.go` is currently a stub (always returns 401) and must be properly implemented as part of this task.

## Objective

Implement `PATCH /api/v1/users/me` — a protected route that reads `userID` from the JWT access token and updates the `name` field on the `users` table. Name must be non-empty after trimming and contain at least one Unicode letter.

## Key Changes

### 1. `pkg/errors/codes.go`
- Add `CodeInvalidName = "INVALID_NAME"`

### 2. `pkg/middleware/auth.go`
- Define local `TokenParser` interface: `ParseAccessToken(string) (uint64, error)`
- Define exported `ContextKeyUserID = "user_id"` constant
- Implement `RequireAuth(parser TokenParser) gin.HandlerFunc`:
  - Extract `Authorization: Bearer <token>` header
  - Call `parser.ParseAccessToken(token)` — on error return `401 INVALID_ACCESS_TOKEN`
  - Set `userID` in gin context via `c.Set(ContextKeyUserID, userID)`

### 3. `internal/modules/user/dto.go`
- Replace placeholder comment with:
  ```go
  type UpdateProfileRequest struct {
      Name string `json:"name" binding:"required"`
  }
  type UpdateProfileResponse struct {
      ID   uint64 `json:"id"`
      Name string `json:"name"`
  }
  ```

### 4. `internal/modules/user/repository.go`
- Add `UpdateName(ctx, userID uint64, name string) error`:
  - GORM: `Model(&User{}).Where("id = ? AND deleted_at IS NULL", userID).Update("name", name)`
  - Check `RowsAffected == 0` → return `ErrUserNotFound`

### 5. `internal/modules/user/errors.go`
- Add `ErrInvalidName = errors.New("invalid name")`
- Register mapper: `ErrInvalidName` → `CodeInvalidName`, `400`, message `"invalid name"`

### 6. `internal/modules/user/service.go` *(new file)*
```go
type userRepository interface {
    UpdateName(ctx context.Context, userID uint64, name string) error
}

type Service interface {
    UpdateProfile(ctx context.Context, userID uint64, req UpdateProfileRequest) (*UpdateProfileResponse, error)
}

type service struct{ repo userRepository }

func NewService(repo *Repository) Service { return &service{repo: repo} }

func (s *service) UpdateProfile(ctx, userID, req) ...
    // 1. name := strings.TrimSpace(req.Name)
    // 2. validate: regexp `\p{L}` must match (at least one Unicode letter)
    //    only allowed chars: letters, spaces, hyphens, apostrophes → `^[\p{L}\s'\-]+$`
    //    violation → return ErrInvalidName
    // 3. s.repo.UpdateName(ctx, userID, name)
    // 4. return &UpdateProfileResponse{ID: userID, Name: name}, nil
```

### 7. `internal/modules/user/handler.go` *(new file)*
```go
type Handler struct{ service Service }
func NewHandler(s Service) *Handler

func (h *Handler) UpdateProfile(c *gin.Context):
    // 1. Extract userID from c.Get(middleware.ContextKeyUserID)
    // 2. ShouldBindJSON(&req) → 400 on error
    // 3. h.service.UpdateProfile(...) → response.FromError on error
    // 4. response.OK(c, "Profile updated successfully", resp)
```

### 8. `internal/modules/user/routes.go` *(new file)*
```go
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
    users := rg.Group("/users")
    users.PATCH("/me", h.UpdateProfile)
}
```

### 9. `internal/bootstrap/dependencies.go`
- Add `UserHandler *user.Handler` and `TokenParser middleware.TokenParser` to `Dependencies`
- Wire up: `userSvc := user.NewService(userRepo)`, `deps.UserHandler = user.NewHandler(userSvc)`, `deps.TokenParser = tokenProvider`

### 10. `internal/bootstrap/router.go`
```go
protected := api.Group("")
protected.Use(middleware.RequireAuth(deps.TokenParser))
user.RegisterRoutes(protected, deps.UserHandler)
```

### 11. `tests/unit/user/` *(replace doc.go placeholder)*
- `handler_test.go`: test UpdateProfile — missing body → 400, valid request → 200, missing user ID in context → 401
- `service_test.go`: test UpdateProfile — valid name → success, empty name → ErrInvalidName, numbers-only → ErrInvalidName, name with space → success, name with hyphen/apostrophe → success

### 12. Documentation
- `docs/modules/user.md` — add endpoint flow section for `PATCH /api/v1/users/me`
- `docs/api/user.postman_collection.json` — add profile update item (method, auth, request, response, errors)
- `docs/knowledge-base.md` — note profile update API, `RequireAuth` middleware contract, `ContextKeyUserID` key

## Files Impacted

| File | Action |
|------|--------|
| `pkg/errors/codes.go` | Add `CodeInvalidName` |
| `pkg/middleware/auth.go` | Implement `RequireAuth`, `TokenParser`, `ContextKeyUserID` |
| `internal/modules/user/dto.go` | Add DTOs |
| `internal/modules/user/repository.go` | Add `UpdateName` |
| `internal/modules/user/errors.go` | Add `ErrInvalidName` + mapper |
| `internal/modules/user/service.go` | **NEW** — Service interface + impl |
| `internal/modules/user/handler.go` | **NEW** — Handler |
| `internal/modules/user/routes.go` | **NEW** — Route registration |
| `internal/bootstrap/dependencies.go` | Wire user handler + token parser |
| `internal/bootstrap/router.go` | Add protected group + user routes |
| `tests/unit/user/handler_test.go` | **NEW** — handler unit tests |
| `tests/unit/user/service_test.go` | **NEW** — service unit tests |
| `docs/modules/user.md` | Endpoint flow docs |
| `docs/api/user.postman_collection.json` | Postman entry |
| `docs/knowledge-base.md` | Update |

## Execution Steps

1. Add `CodeInvalidName` to `pkg/errors/codes.go`
2. Implement `RequireAuth` + `TokenParser` interface + `ContextKeyUserID` in `pkg/middleware/auth.go`
3. Update `internal/modules/user/dto.go` with request/response DTOs
4. Add `UpdateName` to `internal/modules/user/repository.go` (with `RowsAffected` check)
5. Add `ErrInvalidName` + mapper to `internal/modules/user/errors.go`
6. Create `internal/modules/user/service.go` with Service interface, local repo interface, and name validation
7. Create `internal/modules/user/handler.go`
8. Create `internal/modules/user/routes.go`
9. Update `internal/bootstrap/dependencies.go` to wire user module and expose `TokenParser`
10. Update `internal/bootstrap/router.go` to add protected group with user routes
11. Write unit tests in `tests/unit/user/`
12. Update docs: `docs/modules/user.md`, `docs/api/user.postman_collection.json`, `docs/knowledge-base.md`

## Risks / Notes

- **`RequireAuth` is currently a stub**: It always aborts with 401. Changing it to a real implementation is a prerequisite for the protected route to work. No existing routes use it yet so the change is safe.
- **Name regex**: Use `regexp.MustCompile` at package level (not per-call) to avoid recompilation. Pattern `^[\p{L}\s'\-]+$` covers Unicode letters, spaces, hyphens, apostrophes.
- **GORM `Update` vs `Updates`**: `Update("name", value)` explicitly updates a single named column regardless of zero-value — correct for string fields.
- **`RowsAffected` check**: If the user ID from the token doesn't match any active user row, `RowsAffected == 0` allows returning a clean `ErrUserNotFound`.
- **`TokenParser` as interface in `pkg/middleware`**: The `JWTProvider` from `internal/infra/token` satisfies `middleware.TokenParser` implicitly; no import from `internal` into `pkg`.

## Definition of Done

- [ ] `PATCH /api/v1/users/me` with valid `Authorization: Bearer <token>` + device-context headers + `{"name": "John Doe"}` returns `200` with updated name
- [ ] Missing/invalid `Authorization` header → `401 INVALID_ACCESS_TOKEN`
- [ ] Empty or numbers-only name → `400 INVALID_NAME`
- [ ] Name with valid chars (letters, space, hyphen, apostrophe) → accepted
- [ ] `go vet ./...`, `go test ./...`, `make build` all pass
- [ ] Postman collection updated, module and knowledge-base docs updated
