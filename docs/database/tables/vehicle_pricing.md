# `vehicle_pricing` Table

## Purpose

- Stores buying and sale-tag pricing metadata for each vehicle.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `vehicle_id`: `BIGINT`, not null, unique, foreign key.
- `buying_price`: `NUMERIC(10,2)`, not null.
- `buying_date`: `DATE`, not null.
- `price_tag`: `NUMERIC(10,2)`, nullable.
- `tagged_at`: `TIMESTAMPTZ`, nullable.
- `currency`: `VARCHAR(10)`, not null, default `'inr'`.
- `remarks`: `TEXT`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, nullable.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `vehicle_id` (one pricing row per vehicle).

## Foreign Keys

- `vehicle_id -> vehicles.id`.
