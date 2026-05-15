# `user_sessions` Table

## Purpose

- Stores refresh-token-backed user sessions by platform/device.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `user_id`: `BIGINT`, not null, foreign key.
- `platform`: `platform_type`, enum, not null (`web`, `ios_mobile`, `android_mobile`, `desktop`).
- `device_id`: `VARCHAR(255)`, nullable.
- `ip_address`: `VARCHAR(45)`, nullable.
- `refresh_token_hash`: `VARCHAR(256)`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `last_used_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `expires_at`: `TIMESTAMPTZ`, nullable.
- `revoked`: `BOOLEAN`, not null, default `false`.
- `compromised`: `BOOLEAN`, not null, default `false`.
- `revoked_reason`: `VARCHAR(255)`, nullable.

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `user_id -> users.id`.

## Indexes

- `idx_user_sessions_user_id` on `user_id`.
- `idx_user_sessions_refresh_token_hash` on `refresh_token_hash`.
