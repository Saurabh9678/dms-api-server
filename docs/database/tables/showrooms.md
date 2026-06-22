# `showrooms` Table

## Purpose

- Stores showroom master records.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `name`: `VARCHAR`, not null.
- `showroom_logo`: `TEXT`, nullable. Relative storage path set after upload.
- `showroom_banner`: `TEXT`, nullable. Added via migration `000020_add_showroom_banner_to_showrooms`. Relative storage path set after upload.
- `showroom_geolocation`: `JSON`, nullable. Stores a JSON object: `address`, `city`, `state`, `pincode`, `lat`, `lng`.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys Referencing This Table

- `user_showroom_relations.showroom_id -> showrooms.id`.
- `vehicle_showroom_relations.showroom_id -> showrooms.id`.

## Migration Notes

- `000020_add_showroom_banner_to_showrooms.up.sql`: `ALTER TABLE showrooms ADD COLUMN showroom_banner TEXT;`
- `000020_add_showroom_banner_to_showrooms.down.sql`: `ALTER TABLE showrooms DROP COLUMN showroom_banner;`
