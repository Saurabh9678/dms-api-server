# `user_roles` Table

## Purpose

- Defines allowed showroom role types for users.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `type`: `VARCHAR`, not null, unique.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `type`.

## Seed Data

- Migration inserts default roles: `owner`, `manager`, `employee`.

## Foreign Keys Referencing This Table

- `user_showroom_relations.role_id -> user_roles.id`.
