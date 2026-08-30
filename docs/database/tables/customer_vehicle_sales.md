# `customer_vehicle_sales` Table

## Purpose

- Captures completed sale transactions between customers and vehicles.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `customer_id`: `BIGINT`, not null, foreign key.
- `vehicle_id`: `BIGINT`, not null, foreign key.
- `sale_price`: `NUMERIC(10,2)`, not null.
- `sale_date`: `DATE`, not null.
- `payment_mode`: `VARCHAR`, nullable.
- `receipt_url`: `TEXT`, nullable.
- `remarks`: `TEXT`, nullable.
- `sold_by`: `BIGINT`, nullable, FK → `users.id` (for “vehicles I sold” queries; survives leaving a showroom).
- `sold_by_name`: `VARCHAR`, nullable — seller name snapshot at sale time.
- `sold_by_country_code`: `VARCHAR`, nullable — seller country code snapshot at sale time.
- `sold_by_phone_number`: `VARCHAR`, nullable — seller local phone snapshot at sale time.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, nullable.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `customer_id -> customers.id`.
- `vehicle_id -> vehicles.id`.
- `sold_by -> users.id` (added in migration `000033`).

## Indexes

- `idx_customer_vehicle_sales_customer_id` on `customer_id`.
- `idx_customer_vehicle_sales_vehicle_id` on `vehicle_id`.
- `idx_customer_vehicle_sales_sold_by` on `sold_by`.
