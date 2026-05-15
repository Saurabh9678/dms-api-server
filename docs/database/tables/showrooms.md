# `showrooms` Table

## Purpose

- Stores showroom master records.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `name`: `VARCHAR`, not null.
- `showroom_logo`: `TEXT`, nullable.
- `showroom_geolocation`: `JSON`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys Referencing This Table

- `user_showroom_relations.showroom_id -> showrooms.id`.
- `vehicle_showroom_relations.showroom_id -> showrooms.id`.
