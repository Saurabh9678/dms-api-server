# `vehicle_expenses` Table

## Purpose

- Stores per-vehicle expense entries, optionally tied to a vehicle status stage.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `vehicle_id`: `BIGINT`, not null, foreign key.
- `status_id`: `BIGINT`, nullable, foreign key.
- `type`: `VARCHAR`, not null.
- `amount`: `NUMERIC(10,2)`, not null.
- `paid_to`: `VARCHAR`, nullable.
- `description`: `TEXT`, nullable.
- `date`: `TIMESTAMPTZ`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `vehicle_id -> vehicles.id`.
- `status_id -> vehicle_statuses.id`.

## Indexes

- `idx_vehicle_expenses_vehicle_id` on `vehicle_id`.
- `idx_vehicle_expenses_status_id` on `status_id`.
