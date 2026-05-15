# Schema Overview

## Purpose

- Summarize major tables, ownership by module, and key relationships.
- Provide a stable index to per-table documentation in `docs/database/tables/`.

## Module Ownership

- `auth`: `user_otps`, `user_sessions`
- `user`: `users`, `user_roles`, `moderators`, `user_showroom_relations`
- `showroom`: `showrooms`, `vehicle_showroom_relations`
- `vehicle`: `vehicles`, `vehicle_images`, `vehicle_documents`, `vehicle_pricing`, `vehicle_statuses`, `vehicle_expenses`
- `customer`: `customers`, `customer_vehicle_sales`

## Auth Schema Notes

- `user_otps.request_id` is an 8-character unique, non-null identifier used by OTP verification APIs.

## Table Documentation Index

- `docs/database/tables/users.md`
- `docs/database/tables/showrooms.md`
- `docs/database/tables/user_roles.md`
- `docs/database/tables/moderators.md`
- `docs/database/tables/user_showroom_relations.md`
- `docs/database/tables/vehicles.md`
- `docs/database/tables/vehicle_showroom_relations.md`
- `docs/database/tables/vehicle_images.md`
- `docs/database/tables/vehicle_documents.md`
- `docs/database/tables/vehicle_pricing.md`
- `docs/database/tables/customers.md`
- `docs/database/tables/customer_vehicle_sales.md`
- `docs/database/tables/vehicle_statuses.md`
- `docs/database/tables/vehicle_expenses.md`
- `docs/database/tables/user_otps.md`
- `docs/database/tables/user_sessions.md`

## Update Checklist

- Update this file whenever schema structure or ownership changes.
- Keep per-table docs synchronized in `docs/database/tables/` (one file per table).
