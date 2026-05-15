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

## Documentation Update Checklist

- Update this file when auth responsibilities, flows, or boundaries change.
- Update `docs/api/auth.postman_collection.json` for endpoint contract changes.
- Keep OTP verification contract aligned with `requestId + otpCode` request body and header-driven platform/device values.
