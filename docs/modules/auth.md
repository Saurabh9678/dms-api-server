# Auth Module

## Responsibility

- User authentication via OTP flow and token lifecycle.
- OTP trigger endpoints return a client-facing `requestId` used for OTP verification.

## Key Components

- Routes and handlers for auth endpoints.
- Service for OTP trigger, verification, token refresh, and logout.
- Repositories for OTP/session persistence and user lookup/create.
- `/api/v1/*` device-context contract enforces `X-Platform` and non-empty `X-Device-Id` headers for auth routes.
- Logout uses `Authorization: Bearer <accessToken>` plus platform/device headers and does not require request body payload.
- Logout revokes active sessions only for the authenticated user on the requested platform.
- OTP verify enforces single active session per user+platform by revoking existing active sessions before creating a new one.

## Boundaries

- Depends on provider interfaces (`otp`, `token`), not infra concrete types.
- Handler layer performs request/response mapping only.

## Endpoint Flow Details

### Common Entry Flow (`/api/v1/auth/*`)

1. Request enters router group `/api/v1`.
2. `RequireDeviceContext` middleware validates:
   - `X-Platform` is present and one of `web`, `ios_mobile`, `android_mobile`, `desktop`.
   - `X-Device-Id` is present and non-empty.
3. On missing/invalid headers, request is aborted with `400`, code `INVALID_DEVICE_CONTEXT`, message `invalid request`.
4. On valid device context, request reaches auth handler methods.

### `POST /api/v1/auth/register`

1. Handler binds JSON (`countryCode`, `phoneNumber`) and headers.
2. Handler calls `service.Register(...)`.
3. Service flow:
   - Normalizes phone inputs.
   - Finds user by country code + phone; creates user if missing.
   - Generates OTP code and unique 8-char `requestId` (retry on duplicate key).
   - Stores OTP in `user_otps` with platform, device, expiry, and request ID.
   - Sends OTP using OTP provider.
4. Success response: `200` with envelope message `OTP sent successfully` and payload `{ message, requestId }`.
5. Failure responses:
   - Invalid body/header format: `400 INVALID_REQUEST` or `400 INVALID_DEVICE_CONTEXT`.
   - Business/infrastructure failures map through centralized error mapper.

### `POST /api/v1/auth/login`

1. Handler binds JSON (`countryCode`, `phoneNumber`) and headers.
2. Handler calls `service.Login(...)`.
3. Service uses same OTP trigger flow as register (`triggerOTP`).
4. Success response: `200` with envelope message `OTP sent successfully` and payload `{ message, requestId }`.
5. Failure responses follow same structure as register.

### `POST /api/v1/auth/verify-otp`

1. Handler binds JSON (`requestId`, `otpCode`) and headers.
2. Handler calls `service.VerifyOTP(...)`.
3. Service flow:
   - Fetches OTP by `requestId + platform + otp_for`.
   - Rejects used OTP (`OTP_ALREADY_USED`), expired OTP (`OTP_EXPIRED`), or attempts exceeded (`OTP_ATTEMPTS_EXCEEDED`).
   - On wrong code, increments `attempt_count` and returns `INVALID_OTP`.
   - On valid code, marks OTP used.
   - Issues access/refresh token pair via token provider.
   - Revokes all active sessions for same user+platform.
   - Creates new session row in `user_sessions`.
4. Success response: `200` with envelope message `OTP verified successfully` and payload `{ accessToken, refreshToken, expiresIn, tokenType }`.
5. Failure responses:
   - Validation/device context errors -> `400`.
   - OTP/token/session business errors -> mapped auth error codes, generally `401`.

### `POST /api/v1/auth/refresh-token`

1. Handler binds JSON (`refreshToken`) and calls `service.RefreshToken(...)`.
2. Service flow:
   - Hashes refresh token.
   - Finds session by token hash; rejects if missing (`INVALID_REFRESH_TOKEN`).
   - Rejects revoked session (`SESSION_REVOKED`).
   - If expired, revokes session then returns `INVALID_REFRESH_TOKEN`.
   - Rotates tokens via token provider.
   - Persists rotated refresh-token hash and new expiry in same session.
3. Success response: `200` with envelope message `Token refreshed successfully` and payload `{ accessToken, refreshToken, expiresIn, tokenType }`.
4. Failure responses:
   - Invalid body -> `400 INVALID_REQUEST`.
   - Invalid/expired/revoked refresh token -> mapped auth errors, generally `401`.

### `POST /api/v1/auth/logout`

1. Handler reads `Authorization: Bearer <accessToken>` and auth headers.
2. Handler builds `LogoutRequest` with access token and platform.
3. Service flow:
   - Parses access token to extract `user_id`.
   - Revokes all active sessions for `user_id + platform` with reason `user logout`.
4. Success response: `200` with envelope message `Logged out successfully` and `data: null`.
5. Failure responses:
   - Missing/invalid bearer token -> `400 INVALID_REQUEST`.
   - Invalid access token -> `401 INVALID_ACCESS_TOKEN`.
   - Missing/invalid device context -> `400 INVALID_DEVICE_CONTEXT`.

## Documentation Update Checklist

- Update this file when auth responsibilities, flows, or boundaries change.
- Update `docs/api/auth.postman_collection.json` for endpoint contract changes.
- Keep OTP verification contract aligned with `requestId + otpCode` request body and header-driven platform/device values.
- Keep endpoint flow sections synchronized with router, middleware, handler, service, and response behavior.
