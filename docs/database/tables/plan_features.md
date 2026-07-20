# `plan_features` Table

## Purpose

- Maps subscription features to plans with typed values.

## Columns

- `id`: `UUID`, primary key, default `gen_random_uuid()`, not null.
- `subscription_plan_id`: `UUID`, not null, foreign key.
- `subscription_feature_id`: `UUID`, not null, foreign key.
- `bool_value`: `BOOLEAN`, nullable.
- `numeric_value`: `NUMERIC(18,2)`, nullable.
- `string_value`: `TEXT`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `(subscription_plan_id, subscription_feature_id)`.

## Foreign Keys

- `subscription_plan_id -> subscription_plans.id`.
- `subscription_feature_id -> subscription_features.id`.

## Indexes

- `idx_plan_features_subscription_plan_id` on `subscription_plan_id`.
- `idx_plan_features_subscription_feature_id` on `subscription_feature_id`.
