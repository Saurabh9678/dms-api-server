# Knowledge Base

This file is the living project memory for architecture, conventions, and implementation history.

## How To Use

- Read this file before implementing changes.
- Update this file after implementation changes.
- Link to detailed docs in `docs/architecture/`, `docs/modules/`, `docs/providers/`, `docs/api/`, `docs/database/`, and `docs/workflows/`.
- For quick API/function tracing, consult module flow docs in `docs/modules/*.md` first.
- For quick schema validation, consult per-table docs in `docs/database/tables/*.md`.

## Architecture Decisions

- Record approved architecture decisions and rationale.

## Module Responsibilities

- Summarize each module's ownership and boundaries.

## Provider Responsibilities

- Summarize external provider abstractions and ownership.

## Dependency Rules

- Track allowed dependency direction and boundary constraints.

## Conventions

- Track naming, folder placement, and implementation conventions.
- Local development environment variables are documented in `.env.example`.
- Database connection should be configured through `DB_URL`.
- Every `/api/v1/*` endpoint requires `X-Platform` and non-empty `X-Device-Id` headers.
- Module docs must capture endpoint flow end-to-end: entry route, middleware behavior, handler mapping, service business logic, and response branches.
- Schema docs are maintained as one file per table under `docs/database/tables/` and include columns, types, nullability, defaults, PK/FK, and index/unique constraints.


## Important Implementation Details

- Capture non-obvious design details needed for future work.
- Local Postgres Docker setup is defined in `docker-compose.yml` and uses database `dms` on `localhost:5432`.
- Local DB start/stop/log commands are exposed via `make docker-postgres-up`, `make docker-postgres-down`, and `make docker-postgres-logs`.
- Auth register/login responses now include `requestId` in payload; verify-otp accepts `requestId` plus `otpCode`.
- Verify OTP response now includes `required_name` (bool). `true` means `user.name IS NULL` — client should prompt the user to set a profile name. `false` means name is already set. The service fetches the user by ID after session creation to determine this. `VerifyOTPResponse` is a separate DTO from `TokenResponse` (which is used by refresh-token).
- Auth logout now uses `Authorization: Bearer <accessToken>` with `X-Platform`, no request body, and revokes active sessions only for that platform.
- OTP verify enforces single active session per user+platform by revoking existing active sessions on that platform before creating a new session.
- Missing or invalid `/api/v1` device-context headers return `INVALID_DEVICE_CONTEXT` with message `invalid request`.
- `RequireAuth` middleware in `pkg/middleware/auth.go` implements token-based access control. It takes a `TokenParser` interface (defined locally in `pkg/`) to parse JWT access tokens without creating a dependency on `internal/` packages. The middleware extracts Bearer tokens from `Authorization` headers, parses them, sets `userID` in context, and returns 401 `INVALID_ACCESS_TOKEN` on failure.
- Protected routes are registered on a sub-group created within the API v1 group that chains `RequireAuth(tokenProvider)` middleware. This ensures protected endpoints automatically inherit `RequireDeviceContext` from the parent group.
- User profile name validation follows the convention: validation logic lives in the service layer (business rules), while binding tags handle required/format constraints. Valid name characters: Unicode letters, spaces, hyphens, apostrophes (regex `^[\p{L}\s''-]+$`). Empty or blank names are rejected with 400 `INVALID_REQUEST`.
- `PATCH /api/v1/user/me` is the first protected user endpoint. It updates the authenticated user's profile name by extracting user ID from the JWT context set by `RequireAuth`.
- User module routes are registered on a `/user` sub-group inside `RegisterRoutes` (`internal/modules/user/routes.go`), matching the auth module pattern. Individual endpoints (e.g., `PATCH /me`) are then registered on this sub-group, ensuring logical grouping and future-proof extensibility.

## Known Caveats

- Document pitfalls, limitations, and sharp edges.

## Important Workflows

- Document required implementation, testing, and release workflows.
- `make verify` is the required validation entrypoint: lint (`scripts/verify-lint.sh`), tests, **100% function coverage on changed packages** (`scripts/verify-changed-coverage.sh` vs `origin/main`), build, and graphify update.
- Local database setup workflow is documented in `docs/database/local-postgres.md`.
- API behavior change workflow: update Postman collection + module endpoint-flow documentation in the same task.
- Schema change workflow: update relevant files in `docs/database/tables/` in the same task.
