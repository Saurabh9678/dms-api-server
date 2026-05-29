# Vehicle Module

## Responsibility

- Own vehicle domain models, inventory management, and listing business flows.

## Key Components

- Vehicle models, DTOs, repository, service, handler.

## Boundaries

- Keep vehicle-specific rules and persistence in this module.
- Avoid leaking vehicle logic into unrelated modules.

---

## Endpoints

### POST /api/v1/vehicle — Create Vehicle

**Flow:**
1. `POST /api/v1/vehicle` → `RequireDeviceContext` → `RequireAuth` → `vehicle.Handler.CreateVehicle`
2. Handler: `ShouldBindJSON` → calls `service.CreateVehicle`
3. Service: validates all fields (type, manufacturer, model, variant, color, year, RTO, registration, state, usageKM, fuel, transmission) → calls `repo.Create`
4. Repository: GORM `Create` on `vehicles` table
5. Response: `201 Created` with vehicle fields

---

### GET /api/v1/vehicle/listing — List Vehicles by Category

**Flow:**
1. `GET /api/v1/vehicle/listing` → `RequireDeviceContext` → `RequireAuth` → `vehicle.Handler.ListVehicles`
2. Handler: `ShouldBindQuery` → calls `service.ListVehicles`
3. Service:
   - Validates query (page ≥ 1, limit 1–100, valid status/type enums, min_price ≤ max_price)
   - Defaults `status` to `ready_for_sale` when empty
   - Calls `repo.CountByType` → per-category totals
   - Calls `repo.List` → paginated vehicles with current status + pricing
   - Groups results by vehicle_type (car/bike/scooty)
   - Returns `ListVehiclesResponse` with only requested categories (omits unmatched when type filter applied)
4. Repository:
   - Uses LATERAL JOIN to get latest `vehicle_statuses` row (by `id DESC`) as current status
   - Uses LATERAL JOIN to get latest `vehicle_pricing` row (by `id DESC`) as current pricing
   - Applies filters: `vs.status = ANY(statuses)`, `v.vehicle_type = ANY(types)`, price range on `price_tag`
   - Paginates with `LIMIT/OFFSET`
5. Response: `200 OK` with grouped response — `cars`, `bikes`, `scooties` each having `total`, `page`, `limit`, `vehicles[]`

**Query Parameters:**
| Param | Default | Notes |
|---|---|---|
| `status` (repeatable) | `ready_for_sale` | garage, inspection, ready_for_sale, sold |
| `type` (repeatable) | all types | car, bike, scooty |
| `min_price` | — | filters on `price_tag` |
| `max_price` | — | filters on `price_tag` |
| `page` | 1 | ≥ 1 |
| `limit` | 20 | 1–100 |

**Response Shape:**
```json
{
  "success": true,
  "message": "vehicle listing",
  "data": {
    "cars":     { "total": 5, "page": 1, "limit": 20, "vehicles": [...] },
    "bikes":    { "total": 3, "page": 1, "limit": 20, "vehicles": [...] },
    "scooties": { "total": 2, "page": 1, "limit": 20, "vehicles": [...] }
  }
}
```

---

## Documentation Update Checklist

- Update this file when vehicle behavior, schema assumptions, or APIs change.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
