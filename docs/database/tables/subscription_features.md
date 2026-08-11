# `subscription_features` Table

## Purpose

- Defines reusable subscription feature metadata and value type.

## Columns

- `id`: `UUID`, primary key, default `gen_random_uuid()`, not null.
- `key`: `VARCHAR(100)`, not null, unique among non-deleted rows.
- `name`: `VARCHAR(100)`, not null.
- `description`: `TEXT`, nullable.
- `value_type`: `feature_value_type` enum (`BOOLEAN`, `NUMBER`, `STRING`), not null.
- `category`: `VARCHAR(100)`, nullable.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete).

## Keys And Constraints

- Primary key: `id`.
- Partial unique index `idx_subscription_features_key_active` on `key` where `deleted_at IS NULL`.

## Foreign Keys Referencing This Table

- `plan_features.subscription_feature_id -> subscription_features.id`.
