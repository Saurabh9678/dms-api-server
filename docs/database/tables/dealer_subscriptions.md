# `dealer_subscriptions` Table

## Purpose

- Tracks dealer (user) subscription lifecycle against a plan and pricing tier.

## Columns

- `id`: `UUID`, primary key, default `gen_random_uuid()`, not null.
- `dealer_id`: `BIGINT`, not null, foreign key to `users.id`.
- `subscription_plan_id`: `UUID`, not null, foreign key.
- `subscription_pricing_id`: `UUID`, not null, foreign key.
- `status`: `subscription_status_type` enum (`PENDING`, `ACTIVE`, `EXPIRED`, `CANCELLED`, `SUSPENDED`), not null.
- `starts_at`: `TIMESTAMPTZ`, not null.
- `expires_at`: `TIMESTAMPTZ`, not null.
- `auto_renew`: `BOOLEAN`, not null, default `TRUE`.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `dealer_id -> users.id`.
- `subscription_plan_id -> subscription_plans.id`.
- `subscription_pricing_id -> subscription_pricing.id`.

## Indexes

- `idx_dealer_subscriptions_dealer_id` on `dealer_id`.
- `idx_dealer_subscriptions_subscription_plan_id` on `subscription_plan_id`.
- `idx_dealer_subscriptions_subscription_pricing_id` on `subscription_pricing_id`.
- `idx_dealer_subscriptions_status` on `status`.
