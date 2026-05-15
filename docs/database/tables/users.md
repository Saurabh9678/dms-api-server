# `users` Table

## Purpose

- Stores core user identity and profile primitives shared across modules.

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
- No explicit unique constraint on `phone_number` in current migration.

## Foreign Keys Referencing This Table

- `moderators.user_id -> users.id`.
- `user_showroom_relations.user_id -> users.id`.
- `vehicle_images.uploaded_by -> users.id`.
- `vehicle_documents.uploaded_by -> users.id`.
- `vehicle_statuses.added_by -> users.id`.
- `user_otps.user_id -> users.id`.
- `user_sessions.user_id -> users.id`.
