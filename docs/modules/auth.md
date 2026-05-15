# Auth Module

## Responsibility

- User authentication via OTP flow and token lifecycle.

## Key Components

- Routes and handlers for auth endpoints.
- Service for OTP trigger, verification, token refresh, and logout.
- Repositories for OTP/session persistence and user lookup/create.

## Boundaries

- Depends on provider interfaces (`otp`, `token`), not infra concrete types.
- Handler layer performs request/response mapping only.

## Documentation Update Checklist

- Update this file when auth responsibilities, flows, or boundaries change.
- Update `docs/api/auth.postman_collection.json` for endpoint contract changes.
