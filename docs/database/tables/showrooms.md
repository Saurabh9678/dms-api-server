# `showrooms` Table

## Purpose

- Stores showroom master records.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `showroom_id`: `VARCHAR(8)`, not null. Unique external showroom identifier, generated internally as uppercase alphanumeric (`A-Z`, `0-9`).
- `name`: `VARCHAR`, not null.
- `showroom_logo`: `TEXT`, nullable. Object storage key set after upload (not a public URL). Example: `{userID}/showroom/{externalShowroomID}/{YYYYMMDDHHmmss}.jpg`.
- `showroom_banner`: `TEXT`, nullable. Added via migration `000020_add_showroom_banner_to_showrooms`. Same object-key convention as `showroom_logo`.
- `showroom_geolocation`: `JSON`, nullable. Stores a JSON object: `address`, `city`, `state`, `pincode`, `lat`, `lng`.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.
- Unique index: `idx_showrooms_showroom_id` on `showroom_id`.
- Check constraint: `chk_showrooms_showroom_id_format` enforces exactly 8 uppercase alphanumeric characters.

## Foreign Keys Referencing This Table

- `user_showroom_relations.showroom_id -> showrooms.id`.
- `vehicle_showroom_relations.showroom_id -> showrooms.id`.

## Migration Notes

- `000020_add_showroom_banner_to_showrooms.up.sql`: `ALTER TABLE showrooms ADD COLUMN showroom_banner TEXT;`
- `000020_add_showroom_banner_to_showrooms.down.sql`: `ALTER TABLE showrooms DROP COLUMN showroom_banner;`
- `000030_add_showroom_id_to_showrooms.up.sql`: adds `showroom_id`, backfills existing rows from numeric `id` using left-padded base36, enforces format, sets not null, and adds a unique index.
- `000030_add_showroom_id_to_showrooms.down.sql`: drops the unique index, check constraint, and `showroom_id` column.
