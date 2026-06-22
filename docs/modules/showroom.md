# Showroom Module

## Responsibility

- Own showroom domain models and showroom-specific business behaviors.
- Create showrooms and atomically assign the creating user as owner.
- Store and manage showroom logo and banner images via the storage provider.

## Key Components

- `model.go` — `Showroom` struct mapping to the `showrooms` table.
- `dto.go` — `CreateShowroomRequest` and `CreateShowroomResponse`.
- `repository.go` — GORM-backed `Repository` with `CreateWithOwner` (transactional) and `UpdateFilePaths`.
- `service.go` — Business logic: name validation, geolocation JSON validation, file type/size validation, upload orchestration.
- `handler.go` — Parses `multipart/form-data`, extracts userID from context, delegates to service.
- `routes.go` — Registers `POST /showroom` under the protected group.

## Endpoint: POST /api/v1/showroom

**Auth:** Required (Bearer token + device context headers).

**Request:** `multipart/form-data`
- `name` (string, required) — showroom display name.
- `geolocation` (string, optional) — JSON object: `address`, `city`, `state`, `pincode`, `lat`, `lng`. Stored in `showroom_geolocation` column as JSON.
- `showroom_logo` (file, optional) — jpg/jpeg/png, max 10 MB.
- `showroom_banner` (file, optional) — jpg/jpeg/png, max 10 MB.

**Flow:**
1. `middleware.RequireDeviceContext` — validates `X-Platform` and `X-Device-Id`.
2. `middleware.RequireAuth` — validates Bearer token, sets `userID` in context.
3. `Handler.CreateShowroom` — extracts userID, parses multipart form, delegates to service.
4. `service.CreateShowroom`:
   a. Validates name (non-empty after trim).
   b. Validates geolocation JSON if present.
   c. Validates files: size ≤ 10 MB, extension `.jpg`/`.jpeg`/`.png`.
   d. Calls `repo.CreateWithOwner` in a single DB transaction: inserts showroom, looks up `owner` role, inserts `user_showroom_relations` row.
   e. Uploads logo and banner via `storage.Provider` (best-effort; upload failure does not fail the request).
   f. Calls `repo.UpdateFilePaths` if any upload succeeded.
5. Returns 201 with `CreateShowroomResponse`.

**Response branches:**
- `201 Created` — showroom created successfully.
- `400 INVALID_REQUEST` — empty name, invalid geolocation JSON, or multipart parse failure.
- `400 FILE_TOO_LARGE` — file exceeds 10 MB.
- `400 INVALID_FILE_TYPE` — file extension not jpg/jpeg/png.
- `401 INVALID_ACCESS_TOKEN` — missing or invalid Bearer token.

## File Storage

- Provider: `storage.Provider` interface (`internal/providers/storage/provider.go`).
- Current implementation: `internal/infra/storage.LocalProvider` — writes to `{STORAGE_BASE_PATH}/{userID}/{showroomID}/{datetime}.{ext}`.
- Provider is injected via DI in `bootstrap/dependencies.go`.

## Boundaries

- Cross-module dependency on `user_showroom_relations` table is handled via unexported `ownerRelation` struct with explicit `TableName()` — the user module is NOT imported.
- Keep showroom ownership isolated to this module.
- Document any approved dependency on other modules.

## Documentation Update Checklist

- Update this file whenever showroom responsibilities or contracts change.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
