# `vehicles` Table

## Purpose

- Stores primary vehicle inventory records.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `type`: `vehicle_type`, enum, not null (`bike`, `car`, `scooty`).
- `manufacturer`: `VARCHAR`, not null.
- `model`: `VARCHAR`, not null.
- `variant`: `VARCHAR`, not null.
- `color`: `VARCHAR`, not null.
- `year_of_manufacture`: `INT`, not null.
- `rto_code`: `VARCHAR`, not null.
- `registration_number`: `VARCHAR`, not null.
- `registration_state`: `VARCHAR`, not null.
- `usage_km`: `INT`, not null.
- `fuel_type`: `fuel_type`, enum, not null (`petrol`, `diesel`, `ev`).
- `transmission_type`: `transmission_type`, enum, not null (`manual`, `automatic`).
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys Referencing This Table

- `vehicle_showroom_relations.vehicle_id -> vehicles.id`.
- `vehicle_images.vehicle_id -> vehicles.id`.
- `vehicle_documents.vehicle_id -> vehicles.id`.
- `vehicle_pricing.vehicle_id -> vehicles.id`.
- `customer_vehicle_sales.vehicle_id -> vehicles.id`.
- `vehicle_statuses.vehicle_id -> vehicles.id`.
- `vehicle_expenses.vehicle_id -> vehicles.id`.
