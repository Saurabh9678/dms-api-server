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
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, nullable.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `customer_id -> customers.id`.
- `vehicle_id -> vehicles.id`.

## Indexes

- `idx_customer_vehicle_sales_customer_id` on `customer_id`.
- `idx_customer_vehicle_sales_vehicle_id` on `vehicle_id`.
