# `vehicle_documents` Table

## Purpose

- Stores uploaded compliance and ownership document metadata for vehicles.

## Columns

- `id`: `BIGSERIAL`, primary key, auto-increment, not null.
- `vehicle_id`: `BIGINT`, not null, foreign key.
- `document_type`: `vehicle_document_type`, enum, nullable (`registration_certificate`, `insurance`, `pollution`).
- `document_url`: `TEXT`, not null.
- `valid_from`: `DATE`, nullable.
- `valid_till`: `DATE`, nullable.
- `remarks`: `TEXT`, nullable.
- `uploaded_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `uploaded_by`: `BIGINT`, nullable, foreign key.
- `created_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `updated_at`: `TIMESTAMPTZ`, not null, default `CURRENT_TIMESTAMP`.
- `deleted_at`: `TIMESTAMPTZ`, nullable (soft delete marker).

## Keys And Constraints

- Primary key: `id`.

## Foreign Keys

- `vehicle_id -> vehicles.id`.
- `uploaded_by -> users.id`.

## Indexes

- `idx_vehicle_documents_vehicle_id` on `vehicle_id`.
