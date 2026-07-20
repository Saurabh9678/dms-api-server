# `subscription_features` Table

## Purpose

- Defines reusable subscription feature metadata and value type.

## Columns

- `id`: `UUID`, primary key, default `gen_random_uuid()`, not null.
- `key`: `VARCHAR(100)`, not null, unique.
- `name`: `VARCHAR(100)`, not null.
- `description`: `TEXT`, nullable.
- `value_type`: `feature_value_type` enum (`BOOLEAN`, `NUMBER`, `STRING`), not null.
- `category`: `VARCHAR(100)`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.

## Keys And Constraints

- Primary key: `id`.
- Unique constraint: `key`.

## Foreign Keys Referencing This Table

- `plan_features.subscription_feature_id -> subscription_features.id`.
