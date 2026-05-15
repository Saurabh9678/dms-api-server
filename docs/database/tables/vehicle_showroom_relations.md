# `vehicle_showroom_relations` Table

## Purpose

- Maps vehicles to showrooms.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `showroom_id`: `BIGINT`, not null, foreign key.
- `vehicle_id`: `BIGINT`, not null, foreign key.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.
- Unique composite constraint: (`showroom_id`, `vehicle_id`).

## Foreign Keys

- `showroom_id -> showrooms.id`.
- `vehicle_id -> vehicles.id`.

## Indexes

- `idx_vehicle_showroom_relations_vehicle_id` on `vehicle_id`.
- `idx_vehicle_showroom_relations_showroom_id` on `showroom_id`.
