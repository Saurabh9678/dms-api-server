# `user_otps` Table

## Purpose

- Stores OTP challenge records used by auth register/login verification flows.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `user_id`: `BIGINT`, not null, foreign key.
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
- `request_id`: `VARCHAR(8)`, not null (added in migration `000017`).

## Keys And Constraints

- Primary key: `id`.
- Unique index: `idx_user_otps_request_id` on `request_id`.

## Foreign Keys

- `user_id -> users.id`.

## Indexes

- `idx_user_otps_user_id` on `user_id`.
- `idx_user_otps_otp_code` on `otp_code`.
- `idx_user_otps_request_id` on `request_id` (unique).
