# Customer Module

## Responsibility

- Own customer and customer-sales related domain models.
- Sale recording API lives on the vehicle module (`POST /api/v1/vehicle/:id/sale`) and writes `customers` + `customer_vehicle_sales` + sold `vehicle_statuses` in one transaction. Each sale creates a new customer row (not find-or-create).

## Key Components

- Customer model, sales model, and DTOs.

## Boundaries

- Keep customer lifecycle rules in this module.
- Keep inter-module dependencies explicit and documented.
- Vehicle sell endpoint may persist to customer tables via vehicle repository local structs (same pattern as showroom `userRecord`) without importing customer infra.

## Documentation Update Checklist

- Update this file for customer model, workflow, or API contract changes.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
