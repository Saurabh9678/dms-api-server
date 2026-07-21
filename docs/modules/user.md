# User Module

## Responsibility

- Own user domain models and persistence operations.
- Provide user profile management APIs.

## Key Components

- User models and DTOs.
- Repository for user lookup, creation, and profile updates.
- Service layer for profile management with name validation.
- Handler for PATCH and GET /api/v1/user/me endpoints.

## Boundaries

- Exposes user data operations to modules such as auth.
- Keep user domain ownership within this module.

## API Endpoints

### PATCH /api/v1/user/me — Update Profile

**Middleware Chain**: `RequireDeviceContext` → `RequireAuth`

**Flow**:
1. **Route Entry**: `PATCH /api/v1/user/me` — registered as `PATCH /me` on `/user` sub-group inside `RegisterRoutes` (`internal/modules/user/routes.go`), which is mounted on the protected sub-group in `internal/bootstrap/router.go`
2. **RequireDeviceContext Middleware**: Validates `X-Platform` and `X-Device-Id` headers; returns 400 `INVALID_DEVICE_CONTEXT` on failure
3. **RequireAuth Middleware**: Extracts Bearer token from `Authorization` header, parses JWT via `TokenProvider.ParseAccessToken`, sets user ID in context; returns 401 `INVALID_ACCESS_TOKEN` on invalid token
4. **Handler** (`internal/modules/user/handler.go`):
   - Extracts `userID` from context (set by `RequireAuth`)
   - Binds JSON request body to `UpdateProfileRequest` (requires `name` field)
   - Calls `Service.UpdateProfile(ctx, userID, req)`
   - Returns 200 with `UpdateProfileResponse` on success or error response on failure
5. **Service** (`internal/modules/user/service.go`):
   - Validates name: trims whitespace, checks non-empty, validates regex `^[\p{L}\s''-]+$` (Unicode letters, spaces, hyphens, apostrophes)
   - Calls `Repository.UpdateName(ctx, userID, trimmedName)`
   - Returns 400 `INVALID_REQUEST` on validation failure
   - Propagates repository errors (e.g., 404 `USER_NOT_FOUND`)
   - Returns `UpdateProfileResponse` on success
6. **Repository** (`internal/modules/user/repository.go`):
   - `UpdateName` uses GORM `Model().Where("id = ?", userID).Update("name", name)`
   - Returns `ErrUserNotFound` if `RowsAffected == 0`, otherwise returns error or nil

**Response**:
- **200 OK**: `{"success": true, "message": "profile updated", "data": {"name": "<trimmed_name>"}}`
- **400 INVALID_REQUEST**: Name is empty after trim or contains invalid characters
- **400 INVALID_DEVICE_CONTEXT**: Missing or invalid device-context headers
- **401 INVALID_ACCESS_TOKEN**: Missing or invalid Bearer token
- **404 USER_NOT_FOUND**: User ID from token not found in database

### GET /api/v1/user/me — Get Profile

**Middleware Chain**: `RequireDeviceContext` → `RequireAuth`

**Flow**:
1. **Route Entry**: `GET /api/v1/user/me` — registered as `GET /me` on `/user` sub-group inside `RegisterRoutes` (`internal/modules/user/routes.go`)
2. **RequireDeviceContext Middleware**: Validates `X-Platform` and `X-Device-Id` headers; returns 400 `INVALID_DEVICE_CONTEXT` on failure
3. **RequireAuth Middleware**: Extracts Bearer token, parses JWT, sets user ID in context; returns 401 `INVALID_ACCESS_TOKEN` on failure
4. **Handler** (`internal/modules/user/handler.go`):
   - Extracts `userID` from context
   - Calls `Service.GetProfile(ctx, userID)`
   - Returns 200 with `GetProfileResponse` on success
5. **Service** (`internal/modules/user/service.go`):
   - Calls `Repository.FindByID` to fetch user (name, country_code, phone_number)
   - Calls `Repository.FindShowroomRolesByUserID` to fetch all showroom-role pairs
   - Calls `Repository.LoadOnboardingFlags` for active showroom membership and vehicle assignment flags
   - Returns `name` as `*string` (nil if empty), `country_code` separately, and `phone_number` as the local stored phone number without country code
   - Returns `required_name: true` when the authenticated user has not set a name
   - Returns `has_showrooms` and `has_vehicles` for client onboarding decisions
   - Returns `showroom_roles` as a slice (empty array if none). Each role keeps numeric `showroom_id` for existing clients and includes `external_showroom_id` for the generated 8-character showroom identifier.
6. **Repository** (`internal/modules/user/repository.go`):
   - `FindByID` queries `users` table by primary key; returns `ErrUserNotFound` if not found
   - `FindShowroomRolesByUserID` joins `user_showroom_relations`, `showrooms`, and `user_roles`; returns `[]ShowroomRole` with numeric `showroom_id` and string `external_showroom_id`
   - `LoadOnboardingFlags` returns `has_showrooms` when the user has at least one active/non-deleted showroom membership and `has_vehicles` when a non-deleted vehicle belongs to any active/non-deleted showroom for that user

**Response**:
- **200 OK**: `{"success": true, "message": "profile fetched", "data": {"name": "John Doe" | null, "country_code": "+91" | null, "phone_number": "9999999999" | null, "showroom_roles": [{"showroom_id": 1, "external_showroom_id": "SHOP0001", "showroom_name": "Showroom A", "role": "owner"}], "required_name": false, "has_showrooms": true, "has_vehicles": true}}`
- **400 INVALID_DEVICE_CONTEXT**: Missing or invalid device-context headers
- **401 INVALID_ACCESS_TOKEN**: Missing or invalid Bearer token
- **404 USER_NOT_FOUND**: User ID from token not found in database

## Documentation Update Checklist

- Update this file for user model/repository/responsibility changes.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
