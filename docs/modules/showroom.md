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

## Endpoint: POST /api/v1/showroom/:id/member

**Auth:** Required (Bearer token + device context headers + showroom membership).

**Request:** `application/json`
- `user_id` (uint64, required) — ID of the user to add.
- `role` (string, required) — one of `manager`, `employee`.

**Flow:**
1. `middleware.RequireDeviceContext` — validates `X-Platform` and `X-Device-Id`.
2. `middleware.RequireAuth` — validates Bearer token, sets `userID` in context.
3. `middleware.RequireShowroomRoles` — loads `map[showroomID → role]` into context.
4. `Handler.AddMember` — extracts caller userID, parses showroom ID from path, reads showroom roles from context, delegates to service.
5. `service.AddMember`:
   a. Validates `role` is `manager` or `employee`.
   b. Checks caller's role in the showroom (403 `FORBIDDEN` if not owner or manager).
   c. Manager permission check: can only assign `employee` role (403 `FORBIDDEN` otherwise).
   d. Calls `repo.AddMember`: verifies target user exists, checks for duplicate membership, inserts `user_showroom_relations`.
6. Returns 201 with `AddMemberResponse`.

**Response branches:**
- `201 Created` — member added successfully.
- `400 INVALID_REQUEST` — missing/invalid body, or role not `manager`/`employee`.
- `401 INVALID_ACCESS_TOKEN` — missing or invalid Bearer token.
- `403 FORBIDDEN` — caller not owner/manager, or manager trying to add non-employee.
- `409 ALREADY_A_MEMBER` — target user is already an active member.
- `422 TARGET_USER_NOT_FOUND` — target `user_id` does not exist.

## Endpoint: GET /api/v1/showroom/:id/member

**Auth:** Required (Bearer token + device context headers + showroom membership).

**Query params:**
- `page` (int, optional) — page number, default 1, min 1.
- `limit` (int, optional) — items per page, default 20, min 1, max 100.

**Flow:**
1. `middleware.RequireDeviceContext`, `RequireAuth`, `RequireShowroomRoles` — same as above.
2. `Handler.ListMembers` — parses showroom ID, pagination params (with `parseIntParam`), reads showroom roles, delegates to service.
3. `service.ListMembers`:
   a. Checks caller's role (403 `FORBIDDEN` if not owner or manager).
   b. Calls `repo.ListMembers`: JOIN `user_showroom_relations` + `users` + `user_roles`; returns `MemberRecord` slice and total count.
   c. Maps results to `MemberItem` (null for empty name/phone_number).
4. Returns 200 with `ListMembersResponse` (members array, total, page, limit).

**Response branches:**
- `200 OK` — members listed.
- `401 INVALID_ACCESS_TOKEN` — missing or invalid Bearer token.
- `403 FORBIDDEN` — caller not owner or manager.

## Endpoint: DELETE /api/v1/showroom/:id/member/:user_id

**Auth:** Required (Bearer token + device context headers + showroom membership).

**Flow:**
1. `middleware.RequireDeviceContext`, `RequireAuth`, `RequireShowroomRoles`.
2. `Handler.RemoveMember` — parses showroom ID and target user ID from path, delegates to service.
3. `service.RemoveMember`:
   a. Self-removal: if caller == target, allowed if caller has any role in the showroom; blocked if caller has no role (403 `FORBIDDEN`).
   b. Non-self: caller must be owner or manager (403 `FORBIDDEN` otherwise).
   c. Manager restriction: may only remove `employee` role members (403 `FORBIDDEN` for manager/owner target); uses `GetMemberRole` to fetch target's role.
   d. Owner block: owner cannot be removed even by themselves unless they are the target and this is self-removal (the owner check blocks: owner role is not `employee`, so manager path fails; owner-self scenario also blocked because the target's role lookup returns `owner` which is not removable by manager).
   e. Calls `repo.RemoveMember`: soft-deletes the `user_showroom_relations` row (UPDATE `deleted_at`).
4. Returns 200 with null data.

**Response branches:**
- `200 OK` — member removed.
- `400 INVALID_REQUEST` — invalid showroom ID or user ID in path.
- `401 INVALID_ACCESS_TOKEN` — missing or invalid Bearer token.
- `403 FORBIDDEN` — caller lacks permission.
- `404 MEMBER_NOT_FOUND` — target is not an active member of the showroom.

## Endpoint: PATCH /api/v1/showroom/:id/member/:user_id

**Auth:** Required (Bearer token + device context headers + showroom membership).

**Request:** `application/json`
- `role` (string, required) — one of `manager`, `employee`.

**Flow:**
1. `middleware.RequireDeviceContext`, `RequireAuth`, `RequireShowroomRoles`.
2. `Handler.UpdateMemberRole` — parses showroom ID and target user ID from path, delegates to service.
3. `service.UpdateMemberRole`:
   a. Validates `role` is `manager` or `employee`.
   b. Caller must be owner (403 `FORBIDDEN` otherwise).
   c. Blocks self-role-change: caller == target → 403 `FORBIDDEN`.
   d. Calls `repo.UpdateMemberRole`: looks up role ID, updates `user_showroom_relations.role_id`.
4. Returns 200 with null data.

**Response branches:**
- `200 OK` — role updated.
- `400 INVALID_REQUEST` — missing/invalid body, or role not `manager`/`employee`.
- `401 INVALID_ACCESS_TOKEN` — missing or invalid Bearer token.
- `403 FORBIDDEN` — caller not owner, or self-role-change attempt.
- `404 MEMBER_NOT_FOUND` — target is not an active member of the showroom.

## Boundaries

- Cross-module dependency on `user_showroom_relations` table is handled via unexported `ownerRelation` struct with explicit `TableName()` — the user module is NOT imported.
- User existence is checked via unexported `userRecord` struct with `TableName() = "users"` — no import of user module.
- Keep showroom ownership isolated to this module.
- Document any approved dependency on other modules.

## Documentation Update Checklist

- Update this file whenever showroom responsibilities or contracts change.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
