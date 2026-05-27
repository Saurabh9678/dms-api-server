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
   - Returns `name` as `*string` (nil if empty), `phone_number` as `*string` (concat of country_code + phone_number, nil if both empty)
   - Returns `showroom_roles` as a slice (empty array if none)
6. **Repository** (`internal/modules/user/repository.go`):
   - `FindByID` queries `users` table by primary key; returns `ErrUserNotFound` if not found
   - `FindShowroomRolesByUserID` joins `user_showroom_relations`, `showrooms`, and `user_roles`; returns `[]ShowroomRole`

**Response**:
- **200 OK**: `{"success": true, "message": "profile fetched", "data": {"name": "John Doe" | null, "phone_number": "+919999999999" | null, "showroom_roles": [{"showroom_id": 1, "showroom_name": "Showroom A", "role": "owner"}]}}`
- **400 INVALID_DEVICE_CONTEXT**: Missing or invalid device-context headers
- **401 INVALID_ACCESS_TOKEN**: Missing or invalid Bearer token
- **404 USER_NOT_FOUND**: User ID from token not found in database

## Documentation Update Checklist

- Update this file for user model/repository/responsibility changes.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
