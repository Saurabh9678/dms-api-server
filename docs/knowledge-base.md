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
- Auth trigger endpoints (`/auth/send-otp`, `/auth/register`, `/auth/login`) all share the same `triggerOTP` flow. `/auth/send-otp` is the unified canonical endpoint; `/auth/register` and `/auth/login` are deprecated functional aliases. All return `requestId` in payload.
- `triggerOTP` does NOT touch the `users` table. Users are created only in `VerifyOTP` after successful OTP verification.
- OTP rate limiting is enforced by `triggerOTP` before any user lookup: phone-wide cooldown (default 60s, `AUTH_OTP_COOLDOWN_SECONDS`) and phone-wide daily cap (default 10, `AUTH_OTP_MAX_DAILY_SENDS`). Implemented with `FindLatestByPhone` (cooldown) and `CountRecentByPhone` (daily cap) on the `user_otps` table. No Redis required.
- Rate limit errors: `ErrOTPCooldown` → `429 OTP_COOLDOWN`; `ErrOTPRateLimitExceeded` → `429 OTP_RATE_LIMIT_EXCEEDED`.
- **Data ownership boundary**: `user_otps.country_code + user_otps.phone_number` are **snapshots** used only for OTP scoping and rate limiting. `users.country_code + users.phone_number` is the **canonical identity** for all business logic. See `docs/database/tables/user_otps.md` and `docs/database/tables/users.md`.
- `VerifyOTP` find-or-create user pattern: `FindByPhone` → if `ErrUserNotFound` → `Create` → if `ErrDuplicatedKey` (concurrent race) → `FindByPhone` again (re-fetch winning record). The resulting `users.User` is authoritative; OTP record phone fields are used only as lookup keys.
- Verify OTP response includes `required_name` (bool). `true` means `foundUser.Name == ""` — client should prompt via `PATCH /api/v1/user/me`. `false` means name is already set. `VerifyOTPResponse` is a separate DTO from `TokenResponse` (refresh-token response).
- Auth logout now uses `Authorization: Bearer <accessToken>` with `X-Platform`, no request body, and revokes active sessions only for that platform.
- OTP verify enforces single active session per user+platform by revoking existing active sessions on that platform before creating a new session.
- Migration `000018`: adds UNIQUE INDEX on `users(country_code, phone_number)`. Migration `000019`: drops `user_id` from `user_otps`, adds `country_code`/`phone_number` columns and composite index `(country_code, phone_number, created_at)` for phone-based OTP queries.
- Missing or invalid `/api/v1` device-context headers return `INVALID_DEVICE_CONTEXT` with message `invalid request`.
- `RequireAuth` middleware in `pkg/middleware/auth.go` implements token-based access control. It takes a `TokenParser` interface (defined locally in `pkg/`) to parse JWT access tokens without creating a dependency on `internal/` packages. The middleware extracts Bearer tokens from `Authorization` headers, parses them, sets `userID` in context, and returns 401 `INVALID_ACCESS_TOKEN` on failure.
- Protected routes are registered on a sub-group created within the API v1 group that chains `RequireAuth(tokenProvider)` middleware. This ensures protected endpoints automatically inherit `RequireDeviceContext` from the parent group.
- User profile name validation follows the convention: validation logic lives in the service layer (business rules), while binding tags handle required/format constraints. Valid name characters: Unicode letters, spaces, hyphens, apostrophes (regex `^[\p{L}\s''-]+$`). Empty or blank names are rejected with 400 `INVALID_REQUEST`.
- `PATCH /api/v1/user/me` updates the authenticated user's profile name.
- `GET /api/v1/user/me` returns the authenticated user's profile: `name` (*string, null if not set), `phone_number` (*string, concat of country_code+phone_number, null if both empty), `showroom_roles` ([]ShowroomRole with showroom_id, showroom_name, role). Roles come from `user_showroom_relations` joined with `showrooms` and `user_roles`. Empty array `[]` is returned if no relations exist.
- User module routes are registered on a `/user` sub-group inside `RegisterRoutes` (`internal/modules/user/routes.go`), matching the auth module pattern. Individual endpoints (e.g., `PATCH /me`) are then registered on this sub-group, ensuring logical grouping and future-proof extensibility.
- Dashboard module (`internal/modules/dashboard/`) provides `GET /api/v1/dashboard` — a protected executive overview endpoint. Uses a sales-anchored profit model: only SOLD vehicles (present in `customer_vehicle_sales`) affect revenue/profit; unsold vehicles are inventory assets. Duration filter (`1w`/`1m`/`3m`/`6m`/`12m`/`lifetime`) applies to sales (by `sale_date`) and expenses (by `expense.date`) independently. Inventory metrics always reflect current state. The dashboard repository uses raw GORM SQL aggregations directly against `vehicles`, `vehicle_pricing`, `vehicle_expenses`, `customer_vehicle_sales`, `vehicle_showroom_relations`, and `showrooms` — it does not import other modules' repositories. Optional `showroom_id` query param scopes all metrics to a single showroom. `top_vehicle_types` returns only types with ≥1 sale, ordered by `vehicles_sold DESC`.

- Vehicle listing endpoint `GET /api/v1/vehicle/listing` returns vehicles grouped by category (`cars`, `bikes`, `scooties`). Default status filter is `ready_for_sale`; accepts multiple statuses and types via repeated query params. Pagination (`page`/`limit`) applies uniformly across all categories. Repository uses LATERAL JOINs to resolve current status (latest `vehicle_statuses` row by `id DESC`) and current pricing (latest `vehicle_pricing` row by `id DESC`). Categories excluded by a type filter are omitted from the response (`omitempty`). Price range filters apply to `price_tag` field on `vehicle_pricing`.

- **Vehicle update sold-check**: Both `PATCH /vehicle/:id` and `PATCH /vehicle/:id/pricing` block updates when the vehicle is in `sold` status. The check uses a lightweight `GetCurrentStatus` query (`SELECT status FROM vehicle_statuses WHERE vehicle_id = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT 1`) — it does **not** call `GetByIDWithFullDetails`. Returns 422 `VEHICLE_UPDATE_FORBIDDEN`.

- **`registration_number` is immutable**: It is intentionally excluded from `UpdateVehicleRequest`. It cannot be changed via any update API. The field is still returned in `UpdateVehicleResponse`.

- **Vehicle pricing upsert semantics**: `PATCH /vehicle/:id/pricing` acts as an upsert. If no pricing record exists (`GetPricingByVehicleID` returns nil), it creates one (`CreatePricing`) — `buying_price` > 0 and a valid `buying_date` are required in this case. If a record already exists, it calls `UpdatePricingFields` with only the provided fields. `tagged_at` defaults to `time.Now()` and `currency` defaults to `inr` on creation if omitted.

- **Showroom membership check for vehicle updates**: `GetVehicleShowroomID` queries `vehicle_showroom_relations` directly (`SELECT showroom_id WHERE vehicle_id = ? AND deleted_at IS NULL`) — no JOIN needed since the table already holds both IDs. Returns `ErrVehicleNotFound` if no row. Handler then checks if `showroomID` is in the `ContextKeyShowroomRoles` map; missing → 404 `VEHICLE_NOT_FOUND` (no information leak).

- **Showroom creation**: `POST /api/v1/showroom` (protected). Accepts `multipart/form-data` with `name` (required), `geolocation` (optional JSON string), `showroom_logo` (optional file), `showroom_banner` (optional file). Creates showroom and assigns creator as `owner` in a single DB transaction via `showroom.Repository.CreateWithOwner`. File upload is best-effort and happens AFTER transaction commit — upload failure does not roll back or fail the request. Files stored at `{STORAGE_BASE_PATH}/{userID}/{showroomID}/{datetime}.{ext}`. Storage is abstracted via `internal/providers/storage.Provider`; current implementation is `internal/infra/storage.LocalProvider`. Accepted file types: `.jpg`, `.jpeg`, `.png`; max 10 MB.
- **Showroom cross-module isolation**: The showroom repository inserts into `user_showroom_relations` using an unexported `ownerRelation` struct with an explicit `TableName()`. The user module is NOT imported. The `owner` role is looked up by `type = 'owner'` in `user_roles`; if not found, `ErrOwnerRoleNotFound` is returned and the transaction is rolled back.
- **`showroom_banner` migration**: `000020_add_showroom_banner_to_showrooms` adds `showroom_banner TEXT` to `showrooms`. Run migrations before deploying this feature.
- **`STORAGE_BASE_PATH` config**: Defaults to `./uploads`. Set via environment variable to configure local file storage base directory.
- **Postman collection folder pattern**: All API collections now use named folder wrappers matching the `auth` collection pattern (`{ "name": "<module>", "item": [...] }`). Applies to `auth`, `user`, `vehicle`, `dashboard`, and `showroom` collections.
- **Showroom update**: `PATCH /api/v1/showroom/:id`. Permission: owner or manager. Request: `multipart/form-data` with all fields optional (name, geolocation, showroom_logo, showroom_banner, remove_logo, remove_banner). Geolocation is replace-only (not clearable). Logo/banner can be cleared via `remove_logo=true`/`remove_banner=true` flags; a new file upload in the same request overrides the remove flag. File upload is best-effort — if upload fails, the field is silently left unchanged. Uses `repo.GetByID` (new method) for existence check and initial state; `repo.UpdateShowroomFields(map[string]any)` for partial updates (nil values in map write SQL NULL). Merges updates in memory for response — no second DB fetch needed. Returns full showroom object. New error code: `SHOWROOM_NOT_FOUND` (404). New repo sentinel: `ErrShowroomNotFound`.
- **Showroom member management**: `POST/GET/DELETE/PATCH /api/v1/showroom/:id/member` and `DELETE/PATCH /api/v1/showroom/:id/member/:user_id`. Permission matrix — Owner: all 4 ops on any member. Manager: add employees, list all, remove employees only; cannot change roles. Employee: self-removal only. Self-role-change is always blocked (403). Assignable roles: `manager` and `employee` (owner is reserved for creation). Soft-delete pattern: `UPDATE user_showroom_relations SET deleted_at = NOW()` using GORM `Model().Where().Update()` since `ownerRelation` has no `gorm.DeletedAt` field. User existence checked via `COUNT` query against `users` table using unexported `userRecord` struct. Duplicate membership checked via `COUNT` on `user_showroom_relations`. Member role fetched via `Scan` into `[]roleResult` slice (avoids `First` ORDER BY PK issue). Error codes: `FORBIDDEN` (403), `TARGET_USER_NOT_FOUND` (422), `ALREADY_A_MEMBER` (409), `MEMBER_NOT_FOUND` (404). List response includes `name` and `phone_number` as `*string` (null if empty). Pagination: `page`/`limit` query params parsed via `parseIntParam(raw, def, min, max)` helper; defaults page=1, limit=20, max limit=100. `RequireShowroomRoles` middleware must wrap all member management routes; `RegisterRoutes` uses a nested sub-group with `showroomRolesMW` applied.

## Known Caveats

- Document pitfalls, limitations, and sharp edges.

## Important Workflows

- Document required implementation, testing, and release workflows.
- `make verify` is the required validation entrypoint: lint (`scripts/verify-lint.sh`), tests, **100% function coverage on changed packages** (`scripts/verify-changed-coverage.sh` vs `origin/main`), build, and graphify update.
- Local database setup workflow is documented in `docs/database/local-postgres.md`.
- API behavior change workflow: update Postman collection + module endpoint-flow documentation in the same task.
- Schema change workflow: update relevant files in `docs/database/tables/` in the same task.
