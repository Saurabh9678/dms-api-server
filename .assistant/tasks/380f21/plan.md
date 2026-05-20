# Plan

## Objective

Refactor the user module route registration to use a `/user` sub-group (matching the auth module pattern), so that each user endpoint is registered relative to the group root rather than with a hardcoded `/user/` prefix. The full API URL `PATCH /api/v1/user/me` remains unchanged — this is a code-organisation change only.

## Key Changes

1. **`internal/modules/user/routes.go`** — Replace the flat `group.PATCH("/user/me", ...)` registration with a `/user` sub-group and register `PATCH /me` on it.
2. **`tests/unit/user/routes_test.go`** (new) — Add a smoke-style test that calls `user.RegisterRoutes` to satisfy the 100% coverage gate on the changed `internal/modules/user` package.
3. **`docs/modules/user.md`** — Update the "Route Entry" description to reflect the sub-group structure.
4. **`docs/knowledge-base.md`** — Update the routing convention note for user endpoints.

## Files Impacted

| File | Change |
|---|---|
| `internal/modules/user/routes.go` | Core change — sub-group registration |
| `tests/unit/user/routes_test.go` | New — coverage for `RegisterRoutes` |
| `docs/modules/user.md` | Doc update — route entry description |
| `docs/knowledge-base.md` | Doc update — routing convention note |

Not changing: `docs/api/user.postman_collection.json` (URL is identical), `internal/bootstrap/router.go` (caller is unchanged), all handler/service/repository files.

## Execution Steps

### Step 1 — Update `internal/modules/user/routes.go`

```go
package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler) {
    user := group.Group("/user")
    user.PATCH("/me", h.UpdateProfile)
}
```

### Step 2 — Add `tests/unit/user/routes_test.go`

Create `package user_test` test that:
- Builds a minimal gin engine with `RequireDeviceContext` and `RequireAuth` middleware stubs
- Calls `user.RegisterRoutes(protected, user.NewHandler(fakeService))`
- Makes `PATCH /user/me` with valid device-context headers and a set user ID in context
- Asserts 200 OK is returned (proving the route is reachable)
- Also asserts a request to a non-existent path returns 404 (proving the sub-group is scoped)

The fake service can be a minimal struct (reuse the pattern from `handler_test.go`).

### Step 3 — Update `docs/modules/user.md`

Change line:
> **Route Entry**: `PATCH /api/v1/user/me` registered on protected sub-group in `internal/bootstrap/router.go`

To:
> **Route Entry**: `PATCH /api/v1/user/me` — registered as `PATCH /me` on `/user` sub-group inside `RegisterRoutes` (`internal/modules/user/routes.go`), which is mounted on the protected sub-group in `internal/bootstrap/router.go`

### Step 4 — Update `docs/knowledge-base.md`

Change the line:
> `PATCH /api/v1/user/me` is the first protected user endpoint...

To also note that user routes are registered on a `/user` sub-group inside `RegisterRoutes`, matching the auth module pattern.

### Step 5 — Run validation

```bash
make verify
```

Ensure: zero lint errors, all tests pass, 100% coverage on `internal/modules/user`, build succeeds.

## Risks / Notes

- **No URL change**: The final endpoint remains `PATCH /api/v1/user/me`. Postman collection, existing handler tests, and all callers are unaffected.
- **Coverage gate**: `RegisterRoutes` was previously uncovered by `tests/unit/user/`. The new `routes_test.go` is required to pass `make verify-coverage` after the change.
- **Gin variable shadowing**: Inside `RegisterRoutes`, the local variable `user` shadows the package name. Use `userGroup` or `g` as the variable name to avoid a lint warning.

## Definition of Done

- [ ] `internal/modules/user/routes.go` registers routes on a `/user` sub-group
- [ ] `tests/unit/user/routes_test.go` covers `RegisterRoutes` (100% function coverage)
- [ ] `docs/modules/user.md` updated — route entry description reflects sub-group
- [ ] `docs/knowledge-base.md` updated — routing convention note updated
- [ ] `make verify` passes with zero errors
