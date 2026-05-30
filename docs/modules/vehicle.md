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

### GET /api/v1/vehicle/public-listing — Public Showroom Vehicle Listing

**Flow:**
1. `GET /api/v1/vehicle/public-listing` → `RequireDeviceContext` → `vehicle.Handler.PublicListVehicles`
2. No `RequireAuth` — endpoint is publicly accessible.
3. Handler: `ShouldBindQuery` → calls `service.PublicListVehicles`
4. Service:
   - Validates `showroom_id` > 0 (required), page ≥ 1, limit 1–100, sort_by ∈ {price_asc, price_desc}, valid type enums, min_price ≤ max_price
   - Calls `repo.PublicCountByType` → per-category totals scoped to showroom
   - Calls `repo.PublicList` → paginated vehicles with current status + pricing
   - Groups results by vehicle_type; only requested types appear in response
5. Repository:
   - JOINs `vehicle_showroom_relations` on `showroom_id = ?` to scope to the showroom
   - Uses LATERAL JOIN to get latest `vehicle_statuses` row — hardcoded to `ready_for_sale`
   - Uses LATERAL JOIN (inner) to get latest `vehicle_pricing` row where `price_tag IS NOT NULL`
   - Applies optional `vehicle_type`, `min_price`, `max_price` filters
   - Orders by `vp.price_tag ASC` or `DESC` based on `sort_by`
   - Paginates with `LIMIT/OFFSET`
6. Response: `200 OK` — grouped as `cars`, `bikes`, `scooties`, each with `total`, `page`, `limit`, `vehicles[]`. Each vehicle includes `price_tag` and `currency` but **no buying price**.

**Query Parameters:**
| Param | Default | Notes |
|---|---|---|
| `showroom_id` | — | **Required**, must be > 0 |
| `type` (repeatable) | all | car, bike, scooty |
| `min_price` | — | filters on `price_tag` |
| `max_price` | — | filters on `price_tag` |
| `sort_by` | `price_asc` | price_asc, price_desc |
| `page` | 1 | ≥ 1 |
| `limit` | 20 | 1–100 |

**Response Shape:**
```json
{
  "success": true,
  "message": "vehicle listing",
  "data": {
    "cars":     { "total": 2, "page": 1, "limit": 20, "vehicles": [{ "id": 1, "price_tag": 350000, "currency": "inr", ... }] },
    "bikes":    { "total": 0, "page": 1, "limit": 20, "vehicles": [] },
    "scooties": null
  }
}
```

---

## Documentation Update Checklist

- Update this file when vehicle behavior, schema assumptions, or APIs change.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
