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

### GET /api/v1/vehicle/:id — Get Vehicle Details

**Flow:**
1. `GET /api/v1/vehicle/:id` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` middleware → `vehicle.Handler.GetVehicle`
2. Handler: parse `:id` → calls `service.GetVehicleByID`
3. Check `middleware.ContextKeyShowroomRoles` (map[uint64]string) — if vehicle's showroom not in map → 404
4. Role `owner` → `buildAdminResponse` (full details including buying price, expenses, documents, images)
5. Role `manager`/`employee` → `buildBasicResponse` (basic fields + price_tag only, no buying price)

---

### PATCH /api/v1/vehicle/:id — Update Vehicle Core Fields

**Flow:**
1. `PATCH /api/v1/vehicle/:id` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.UpdateVehicle`
2. Handler:
   - Parse `:id` (uint64, must be > 0)
   - `ShouldBindJSON` → `UpdateVehicleRequest` (all pointer fields; nil = skip update)
   - Extract `middleware.ContextKeyShowroomRoles` from context
   - Call `service.GetVehicleShowroomID(ctx, id)` → `SELECT showroom_id FROM vehicle_showroom_relations`
   - If showroom not in roles map → 404 `VEHICLE_NOT_FOUND`
   - Call `service.UpdateVehicle(ctx, id, req)` → 200 on success
3. Service:
   - `GetCurrentStatus(ctx, id)` → `SELECT status FROM vehicle_statuses WHERE vehicle_id = ? ORDER BY id DESC LIMIT 1`
   - If status == `sold` → 422 `VEHICLE_UPDATE_FORBIDDEN`
   - `buildVehicleUpdates(req)` — validates each non-nil field, builds `map[string]interface{}`
   - If map empty → 400 `INVALID_REQUEST`
   - `repo.UpdateVehicleFields(ctx, id, updates)` → GORM `Model().Where().Updates(map)` + re-fetch
4. Response: `200 OK` with `UpdateVehicleResponse` (vehicle fields, no `registration_number` in request — immutable)

**Validation Rules:**
| Field | Rule |
|-------|------|
| `vehicle_type` | `bike`, `car`, or `scooty` |
| string fields | TrimSpace; must not be empty if provided |
| `year_of_manufacture` | 1900–current year inclusive |
| `usage_km` | ≥ 0 |
| `fuel_type` | `petrol`, `diesel`, or `ev` |
| `transmission_type` | `manual` or `automatic` |

**Error Codes:**
| Scenario | HTTP | Code |
|----------|------|------|
| Vehicle not found | 404 | `VEHICLE_NOT_FOUND` |
| Not showroom member | 404 | `VEHICLE_NOT_FOUND` |
| Vehicle is sold | 422 | `VEHICLE_UPDATE_FORBIDDEN` |
| No fields / invalid value | 400 | `INVALID_REQUEST` |

---

### PATCH /api/v1/vehicle/:id/pricing — Update Vehicle Pricing

**Flow:**
1. `PATCH /api/v1/vehicle/:id/pricing` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.UpdateVehiclePricing`
2. Handler:
   - Same membership gate as UpdateVehicle (GetVehicleShowroomID + roles check)
   - Call `service.UpdateVehiclePricing(ctx, id, req)` → 200 on success
3. Service:
   - `GetCurrentStatus` → sold check (422)
   - `GetPricingByVehicleID(ctx, id)` → returns `*VehiclePricing` or nil (not-found treated as nil)
   - **Create branch** (no pricing record): `buying_price` > 0 required, `buying_date` required; `tagged_at` defaults to now, `currency` defaults to `inr` → `CreatePricing`
   - **Update branch** (pricing exists): `buildPricingUpdates(req)` → map; if empty → 400; `UpdatePricingFields`
4. Response: `200 OK` with `UpdateVehiclePricingResponse`

**Validation Rules:**
| Field | Rule |
|-------|------|
| `buying_price` | > 0 if provided; **required** when no pricing record exists |
| `buying_date` | valid `2006-01-02` format; **required** when no pricing record exists |
| `price_tag` | ≥ 0 if provided |
| `tagged_at` | valid RFC3339 if provided; defaults to `time.Now()` on create |
| `currency` | `inr` or `usd`; defaults to `inr` on create |

**DB Queries Per Call:**
- 3 queries: showroom ID lookup, current status, create/update pricing
- 4 queries when pricing record exists: showroom ID, current status, get pricing, update

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
