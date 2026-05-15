# `moderators` Table

## Purpose

- Stores moderator-level access metadata linked to users.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `user_id`: `BIGINT`, not null, unique, foreign key.
- `access_level`: `VARCHAR`, not null.
- `remarks`: `TEXT`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `user_id` (one moderator row per user).

## Foreign Keys

- `user_id -> users.id`.
