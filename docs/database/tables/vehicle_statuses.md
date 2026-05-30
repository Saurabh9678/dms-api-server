# `vehicle_statuses` Table

## Purpose

- Tracks vehicle lifecycle status changes over time.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `vehicle_id`: `BIGINT`, not null, foreign key.
- `status`: `vehicle_status`, enum, not null (`bought`, `garage`, `inspection`, `ready_for_sale`, `sold`). `bought` is the initial status automatically inserted when a vehicle is created.
- `description`: `TEXT`, nullable.
- `started_at`: `TIMESTAMPTZ`, not null.
- `ended_at`: `TIMESTAMPTZ`, nullable.
- `added_by`: `BIGINT`, nullable, foreign key.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `vehicle_id -> vehicles.id`.
- `added_by -> users.id`.

## Indexes

- `idx_vehicle_statuses_vehicle_id` on `vehicle_id`.
