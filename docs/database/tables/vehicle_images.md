# `vehicle_images` Table

## Purpose

- Stores uploaded image metadata for vehicles.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `vehicle_id`: `BIGINT`, not null, foreign key.
- `image_url`: `TEXT`, not null.
- `label`: `vehicle_image_label`, enum, nullable (`front`, `interior`, `exterior`, `back`, `wheel`).
- `uploaded_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `uploaded_by`: `BIGINT`, nullable, foreign key.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `vehicle_id -> vehicles.id`.
- `uploaded_by -> users.id`.

## Indexes

- `idx_vehicle_images_vehicle_id` on `vehicle_id`.
