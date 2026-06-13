# `user_otps` Table

## Purpose

- Stores OTP challenge records used by auth OTP trigger and verification flows.
- Phone identity fields (`country_code`, `phone_number`) are **immutable snapshots** captured at OTP generation time. They are **not** the authoritative identity source — `users.country_code` + `users.phone_number` is canonical for all identity lookups.
- Rate limiting (cooldown, daily cap) is enforced by querying this table by phone; no user lookup required.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `country_code`: `VARCHAR`, not null. Phone identity snapshot at OTP generation time.
- `phone_number`: `VARCHAR`, not null. Phone identity snapshot at OTP generation time.
- `request_id`: `VARCHAR(8)`, not null (added in migration `000017`).
- `otp_code`: `VARCHAR(6)`, not null.
- `platform`: `platform_type`, enum, not null (`web`, `ios_mobile`, `android_mobile`, `desktop`).
- `otp_for`: `otp_for_type`, enum, not null (`mobile`, `email`).
- `device_id`: `VARCHAR(255)`, nullable.
- `attempt_count`: `INT`, not null, default `0`.
- `resend_count`: `INT`, not null, default `0`.
- `is_used`: `BOOLEAN`, not null, default `false`.
- `expires_at`: `TIMESTAMPTZ`, not null.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `verified_at`: `TIMESTAMPTZ`, nullable.

> **Note**: `user_id` was removed in migration `000019`. OTP records are now independent of the `users` table, anchored only by phone identity snapshot columns.

## Keys And Constraints

- Primary key: `id`.
- Unique index: `idx_user_otps_request_id` on `request_id`.

## Foreign Keys

- None. `user_id` was dropped in migration `000019`; OTP records are phone-anchored, not user-anchored.

## Indexes

- `idx_user_otps_request_id` on `request_id` (unique) — supports `FindLatestActiveByRequestIDAndPlatform`.
- `idx_user_otps_phone_created` on `(country_code, phone_number, created_at)` — supports `FindLatestByPhone` (ORDER BY created_at DESC LIMIT 1) and `CountRecentByPhone` (WHERE created_at >= ?).
- `idx_user_otps_otp_code` on `otp_code`.

## Data Ownership Rule

- `user_otps.country_code` + `user_otps.phone_number` are **snapshots** used for: OTP verification lookup, cooldown enforcement, daily rate limiting, and immutable historical audit.
- They are **NOT** used for: determining user identity, account existence checks, profile lookups, or any business logic beyond OTP scoping.
- Any logic that needs to know "who is the user" must query the `users` table.
