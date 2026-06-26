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

## Caveats

- `status_id` is nullable in the DB and is always NULL when created via the API — it exists for potential future use only.
- `type` is a free-form VARCHAR in the DB but the API enforces an enum: `repair`, `service`, `insurance`, `tax`, `inspection`, `cleaning`, `documentation`, `other`.
- `paid_to`, `description`, and `date` are all optional; `date` defaults to the time of the API call when not provided.
