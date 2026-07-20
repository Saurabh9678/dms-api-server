# `moderator_roles` Table

## Purpose

- Defines CMS moderator roles independently from showroom `user_roles`.
- Supports current roles (`super_admin`, `admin`) and future CMS permission expansion.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `code`: `VARCHAR(40)`, not null, unique; must match `^[a-z][a-z0-9_]*$`.
- `name`: `VARCHAR(80)`, not null.
- `description`: `TEXT`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `code`.
- Check constraint: `moderator_roles_code_format` enforces lowercase role codes with underscores.

## Seed Data

- Migration inserts default roles: `super_admin`, `admin`.

## Foreign Keys Referencing This Table

- `moderators.role_id -> moderator_roles.id`.
