# `subscription_plans` Table

## Purpose

- Stores subscription plan catalog entries (global and customer-specific).

## Columns

- `id`: `UUID`, primary key, default `gen_random_uuid()`, not null.
- `code`: `VARCHAR(50)`, not null, unique.
- `name`: `VARCHAR(100)`, not null.
- `description`: `TEXT`, nullable.
- `is_free`: `BOOLEAN`, not null, default `FALSE`.
- `is_global`: `BOOLEAN`, not null, default `TRUE`.
- `customer_id`: `UUID`, nullable (custom plan scope; no FK yet).
- `is_active`: `BOOLEAN`, not null, default `TRUE`.
- `display_order`: `INT`, not null, default `0`.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `code`.

## Foreign Keys Referencing This Table

- `subscription_pricing.subscription_plan_id -> subscription_plans.id`.
- `plan_features.subscription_plan_id -> subscription_plans.id`.
- `dealer_subscriptions.subscription_plan_id -> subscription_plans.id`.
