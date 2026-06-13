# `users` Table

## Purpose

- Stores core user identity and profile primitives shared across modules.
- `country_code` + `phone_number` is the **canonical phone identity** — the authoritative source for all user lookups, token issuance, and business logic.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `email`: `VARCHAR`, nullable.
- `phone_number`: `VARCHAR`, not null.
- `country_code`: `VARCHAR`, not null.
- `name`: `VARCHAR`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.
- Unique index: `idx_users_country_code_phone_number` on `(country_code, phone_number)` (added in migration `000018`). Enforces one account per phone number across country codes.

## Foreign Keys Referencing This Table

- `moderators.user_id -> users.id`.
- `user_showroom_relations.user_id -> users.id`.
- `vehicle_images.uploaded_by -> users.id`.
- `vehicle_documents.uploaded_by -> users.id`.
- `vehicle_statuses.added_by -> users.id`.
- `user_sessions.user_id -> users.id`.

> **Note**: `user_otps.user_id -> users.id` was dropped in migration `000019`. OTP records are no longer linked to user rows.

## Canonical Identity Rule

- `users.country_code` + `users.phone_number` is the **canonical identity anchor** for the entire system.
- All application logic that determines "who is this user", checks account existence, reads profile data, or performs permission checks **must** use the `users` table.
- OTP records (`user_otps`) also store phone fields as snapshots — those are for OTP scoping only, not for identity decisions. See `docs/database/tables/user_otps.md`.
- User records are created in `VerifyOTP` (first successful OTP verification for a phone number). No record is created during OTP trigger (`/auth/send-otp`, `/auth/register`, `/auth/login`).
