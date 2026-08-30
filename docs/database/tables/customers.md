# `customers` Table

## Purpose

- Stores customer profile and KYC/basic contact details.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `first_name`: `VARCHAR`, not null.
- `last_name`: `VARCHAR`, nullable.
- `email`: `VARCHAR`, nullable.
- `phone_number`: `VARCHAR`, not null (duplicates allowed; each sale creates its own customer snapshot).
- `alt_phone_number`: `VARCHAR`, nullable.
- `address`: `TEXT`, nullable.
- `city`: `VARCHAR`, nullable.
- `state`: `VARCHAR`, nullable.
- `pincode`: `VARCHAR`, nullable.
- `id_proof_type`: `VARCHAR`, nullable.
- `id_proof_number`: `VARCHAR`, nullable.
- `id_proof_url`: `TEXT`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, nullable.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.
- No unique constraint on `phone_number` (dropped in migration `000032_drop_customers_phone_unique` so duplicate phones can exist across sale records).

## Foreign Keys Referencing This Table

- `customer_vehicle_sales.customer_id -> customers.id`.
