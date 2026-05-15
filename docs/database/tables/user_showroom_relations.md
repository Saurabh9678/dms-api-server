# `user_showroom_relations` Table

## Purpose

- Maps users to showrooms with explicit role assignments.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `user_id`: `BIGINT`, not null, foreign key.
- `showroom_id`: `BIGINT`, not null, foreign key.
- `role_id`: `BIGINT`, not null, foreign key.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.
- Unique composite constraint: (`user_id`, `showroom_id`, `role_id`).

## Foreign Keys

- `user_id -> users.id`.
- `showroom_id -> showrooms.id`.
- `role_id -> user_roles.id`.

## Indexes

- `idx_user_showroom_relations_user_id` on `user_id`.
- `idx_user_showroom_relations_showroom_id` on `showroom_id`.
- `idx_user_showroom_relations_role_id` on `role_id`.
