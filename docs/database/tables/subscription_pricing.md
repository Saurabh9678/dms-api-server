# `subscription_pricing` Table

## Purpose

- Stores billing-cycle pricing for each subscription plan.

## Columns

- `id`: `UUID`, primary key, default `gen_random_uuid()`, not null.
- `subscription_plan_id`: `UUID`, not null, foreign key.
- `billing_cycle`: `billing_cycle_type` enum (`MONTHLY`, `QUARTERLY`, `HALF_YEARLY`, `YEARLY`), not null.
- `duration_days`: `INT`, not null.
- `price`: `NUMERIC(10,2)`, not null.
- `currency`: `CHAR(3)`, not null, default `'INR'`.
- `is_active`: `BOOLEAN`, not null, default `TRUE`.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `(subscription_plan_id, billing_cycle)`.

## Foreign Keys

- `subscription_plan_id -> subscription_plans.id`.

## Indexes

- `idx_subscription_pricing_subscription_plan_id` on `subscription_plan_id`.

## Foreign Keys Referencing This Table

- `dealer_subscriptions.subscription_pricing_id -> subscription_pricing.id`.
