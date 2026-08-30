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
4. Repository: in one transaction, GORM `Create` on `vehicles`, then inserts `vehicle_statuses` with `status = garage` and `started_at = now` (`added_by`, `description`, and `ended_at` left null)
5. Response: `201 Created` with vehicle fields

---

### GET /api/v1/vehicle/listing — List Vehicles by Category

**Flow:**
1. `GET /api/v1/vehicle/listing` → `RequireDeviceContext` → `RequireAuth` → `RequireShowroomRoles` → `vehicle.Handler.ListVehicles`
2. Handler: `ShouldBindQuery` → requires `showroom_id` > 0 → caller must be a member of that showroom (any role) → calls `service.ListVehicles`
3. Service:
   - Validates query (`showroom_id` > 0, page ≥ 1, limit 1–100, valid status/type enums, min_price ≤ max_price)
   - Omits `status` filter when empty (all current statuses); otherwise filters to requested statuses
   - Calls `repo.CountByType` → per-category totals + dead-stock (same filters as the page query, including status)
   - Calls `repo.CountStatusBreakdownByType` → `available_count` / `repair_count` / `sold_count` (**status filter ignored**; still scoped by showroom/type/price). Mapping: available=`ready_for_sale`, repair=`garage`+`inspection`, sold=`sold`
   - Calls `repo.List` → paginated vehicles with current status + pricing (`LIMIT`/`OFFSET` on the combined result ordered by `v.id DESC`)
   - Batch-loads `vehicle_images` via `repo.ListImagesByVehicleIDs` for the page’s vehicle IDs, signs object keys (1-hour TTL), and attaches an `images` section on each list item
   - Groups results by vehicle_type (car/bike/scooty)
   - Returns `ListVehiclesResponse` with only requested categories (omits unmatched when type filter applied)
4. Repository:
   - JOINs `vehicle_showroom_relations` on `showroom_id = ?` to return only vehicles assigned to that showroom
   - Uses LATERAL JOIN to get latest `vehicle_statuses` row (by `id DESC`) as current status
   - Uses LATERAL JOIN to get latest `vehicle_pricing` row (by `id DESC`) as current pricing
   - Applies optional filters: `vs.status IN (statuses)` when status is provided, `v.type IN (types)` (aliased as `vehicle_type` in responses), price range on `price_tag`. GORM expands slice args inside `(?)`, so listing SQL uses `IN (?)` rather than PostgreSQL `ANY(?)`.
   - Paginates with `LIMIT/OFFSET` (defaults: page 1, limit 20, max 100). `total` is the full matching count per type; `vehicles[]` is the current page only.
5. Response: `200 OK` with grouped response — `cars`, `bikes`, `scooties` each having `total`, `available_count`, `repair_count`, `sold_count`, `dead_stock_count`, `page`, `limit`, `vehicles[]`. Each vehicle includes `current_status` as `{ "status", "started_at" }`, `images`, and `is_dead_stock`. Empty photos → `"images": {}`. Dead-stock uses shared dashboard rule via `pkg/inventory`.

**Query Parameters:**
| Param | Default | Notes |
|---|---|---|
| `showroom_id` | — | **Required**, numeric showroom ID; caller must be a member |
| `status` (repeatable) | all statuses | garage, inspection, ready_for_sale, sold |
| `type` (repeatable) | all types | car, bike, scooty |
| `min_price` | — | filters on `price_tag` |
| `max_price` | — | filters on `price_tag` |
| `page` | 1 | ≥ 1 |
| `limit` | 20 | 1–100 |

**Errors:**
| Condition | Status | Code |
|---|---|---|
| Missing/zero `showroom_id` or invalid query | 400 | `INVALID_REQUEST` |
| Caller is not a member of the showroom | 403 | `FORBIDDEN` |

**Response Shape:**
```json
{
  "success": true,
  "message": "vehicle listing",
  "data": {
    "cars": {
      "total": 5,
      "available_count": 2,
      "repair_count": 2,
      "sold_count": 1,
      "dead_stock_count": 1,
      "page": 1,
      "limit": 20,
      "vehicles": [{ "id": 1, "current_status": { "status": "ready_for_sale", "started_at": "2024-01-01T00:00:00Z" }, "is_dead_stock": true }]
    },
    "bikes": { "total": 3, "available_count": 1, "repair_count": 1, "sold_count": 1, "dead_stock_count": 0, "page": 1, "limit": 20, "vehicles": [] },
    "scooties": { "total": 2, "available_count": 0, "repair_count": 1, "sold_count": 1, "dead_stock_count": 0, "page": 1, "limit": 20, "vehicles": [] }
  }
}
```

---

### GET /api/v1/vehicle/:id — Get Vehicle Details

**Flow:**
1. `GET /api/v1/vehicle/:id` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` middleware → `vehicle.Handler.GetVehicle`
2. Handler: parse `:id` → calls `service.GetVehicleByID`
3. Check `middleware.ContextKeyShowroomRoles` (map[uint64]string) — if vehicle's showroom not in map → 404
4. Role `owner` → `buildAdminResponse` (full details including buying price, expenses, documents, images).
5. Role `manager`/`employee` → `buildBasicResponse` (basic fields + price_tag only, no buying price). Images are included for every showroom role.
6. `images` is a section grouped by label: `{ "front": [{ "id", "url" }, ...], "interior": [...] }`. `url` values are 1-hour signed URLs resolved from stored object keys. Images that fail to sign, and labels with no photos, are omitted. Multiple photos per label are returned as an array. `id` is the `vehicle_images.id` used by `DELETE /api/v1/vehicle/:id/image/:image_id`.

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
   - Batch-loads `vehicle_images` via `repo.ListImagesByVehicleIDs`, signs object keys, and attaches `images` on each public list item (same shape as auth listing)
   - Groups results by vehicle_type; only requested types appear in response
5. Repository:
   - JOINs `vehicle_showroom_relations` on `showroom_id = ?` to scope to the showroom
   - Uses LATERAL JOIN to get latest `vehicle_statuses` row — hardcoded to `ready_for_sale`
   - Uses LATERAL JOIN (inner) to get latest `vehicle_pricing` row where `price_tag IS NOT NULL`
   - Applies optional `vehicle_type` (`v.type IN (?)`), `min_price`, `max_price` filters
   - Orders by `vp.price_tag ASC` or `DESC` based on `sort_by`
   - Paginates with `LIMIT/OFFSET`
6. Response: `200 OK` — grouped as `cars`, `bikes`, `scooties`, each with `total`, `page`, `limit`, `vehicles[]`. Each vehicle includes `price_tag`, `currency`, and `images` (label → `[{id,url}]`, empty object when none) but **no buying price**.

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

### POST /api/v1/vehicle/:id/expense — Add Vehicle Expense

**Flow:**
1. `POST /api/v1/vehicle/:id/expense` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.AddExpense`
2. Handler:
   - Parse `:id` (uint64, must be > 0)
   - `ShouldBindJSON` → `AddExpenseRequest` (`type` and `amount` are required)
   - Extract `middleware.ContextKeyShowroomRoles` from context
   - Call `service.GetVehicleShowroomID(ctx, id)` → `SELECT showroom_id FROM vehicle_showroom_relations`
   - If showroom not in roles map → 404 `VEHICLE_NOT_FOUND`
   - Call `service.AddExpense(ctx, id, req)` → 201 on success
3. Service:
   - Validates `type` against known expense types
   - Validates `amount` > 0
   - If `date` provided: parses RFC3339; if invalid → 400
   - If `date` not provided: defaults to `time.Now()`
   - `repo.CreateExpense(ctx, &VehicleExpenses{...})` → GORM `Create` on `vehicle_expenses` table
4. Response: `201 Created` with `AddExpenseResponse`

**Request Body:**
| Field | Required | Type | Notes |
|-------|----------|------|-------|
| `type` | Yes | string | `repair`, `service`, `insurance`, `tax`, `inspection`, `cleaning`, `documentation`, `other` |
| `amount` | Yes | float64 | Must be > 0 |
| `paid_to` | No | string | Recipient of payment |
| `description` | No | string | Reason/why the expense was made |
| `date` | No | string | RFC3339 format; defaults to current time |

**Error Codes:**
| Scenario | HTTP | Code |
|----------|------|------|
| Vehicle not found | 404 | `VEHICLE_NOT_FOUND` |
| Not showroom member | 404 | `VEHICLE_NOT_FOUND` |
| Invalid/missing type or amount | 400 | `INVALID_REQUEST` |
| Invalid date format | 400 | `INVALID_REQUEST` |

---

### POST /api/v1/vehicle/:id/status — Update Vehicle Status

**Flow:**
1. `POST /api/v1/vehicle/:id/status` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.UpdateVehicleStatus`
2. Handler:
   - Parse `:id` (uint64, must be > 0)
   - `ShouldBindJSON` → `UpdateVehicleStatusRequest`
   - Extract showroom roles; resolve vehicle showroom; non-member → 404 `VEHICLE_NOT_FOUND`
   - Any showroom role may update status
   - `added_by` = authenticated user id
   - Call `service.UpdateVehicleStatus` → 201 on success
3. Service:
   - Validates `status` ∈ `garage` | `inspection` | `ready_for_sale` (`sold` not allowed here — use sell API)
   - Optional `description` (trimmed)
   - Calls `repo.UpdateStatus` (single transaction)
4. Repository transaction:
   - Load latest `vehicle_statuses` row; missing → `ErrVehicleNotFound`
   - Current `sold` → `ErrVehicleSold` (422 `VEHICLE_UPDATE_FORBIDDEN`)
   - Same status as requested → `ErrVehicleStatusUnchanged` (400 `INVALID_REQUEST`)
   - Set previous row `ended_at = now`
   - Insert new status row with `started_at = now`, `added_by`, optional description
5. Response: `201 Created` with `{ id, vehicle_id, status, description, started_at, added_by }`

**Request Body:**
| Field | Required | Type | Notes |
|-------|----------|------|-------|
| `status` | Yes | string | `garage`, `inspection`, or `ready_for_sale` |
| `description` | No | string | Free-text note |

**Errors:**
| Scenario | HTTP | Code |
|----------|------|------|
| Invalid/missing status (incl. `sold`) or same as current | 400 | `INVALID_REQUEST` |
| Vehicle not found / not member | 404 | `VEHICLE_NOT_FOUND` |
| Vehicle already sold | 422 | `VEHICLE_UPDATE_FORBIDDEN` |

---

### POST /api/v1/vehicle/:id/sale — Sell Vehicle

**Flow:**
1. `POST /api/v1/vehicle/:id/sale` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.SellVehicle`
2. Handler:
   - Parse `:id` (uint64, must be > 0)
   - `ShouldBindJSON` → `SellVehicleRequest`
   - Extract showroom roles; resolve vehicle showroom; non-member → 404 `VEHICLE_NOT_FOUND`
   - Any showroom role (owner/manager/employee) may sell
   - Call `service.SellVehicle` → 201 on success
3. Service:
   - Validates `sale_price` > 0
   - Validates `payment_mode` enum: `cash`, `cheque`, `bank_transfer`, `online`, `credit`, `debit`, `other`
   - Requires trimmed non-empty customer `first_name`, `last_name`, `phone_number`, `address`
   - `sale_date` optional (`YYYY-MM-DD`); defaults to today (UTC date)
   - Calls `repo.SellVehicle` (single transaction)
4. Repository transaction:
   - Load current status; missing vehicle → `ErrVehicleNotFound`; status `sold` → `ErrVehicleAlreadySold`
   - Reject if an active `customer_vehicle_sales` row exists → `ErrVehicleAlreadySold`
   - Load seller from `users` by authenticated user id; snapshot name/country_code/phone_number
   - Always **create a new** `customers` row (sale snapshot; phone uniqueness removed in migration `000032`)
   - Insert `customer_vehicle_sales` with `sold_by` + seller snapshot columns (migration `000033`)
   - Insert `vehicle_statuses` with `status = sold` and `added_by = seller`
5. Response: `201 Created` with sale id, vehicle id, sale fields, customer snapshot, and `sold_by` `{ user_id, name, country_code, phone_number }`
6. Owner `GET /vehicle/:id` admin `selling` section also returns `sold_by` snapshot when present.

**Request Body:**
| Field | Required | Type | Notes |
|-------|----------|------|-------|
| `sale_price` | Yes | float64 | Must be > 0 |
| `sale_date` | No | string | `YYYY-MM-DD`; defaults to today UTC |
| `payment_mode` | Yes | string | Enum above |
| `remarks` | No | string | Optional notes |
| `customer.first_name` | Yes | string | |
| `customer.last_name` | Yes | string | |
| `customer.phone_number` | Yes | string | |
| `customer.address` | Yes | string | |
| `customer.email` / `city` / `state` / `pincode` | No | string | |

**Error Codes:**
| Scenario | HTTP | Code |
|----------|------|------|
| Vehicle not found / not a member | 404 | `VEHICLE_NOT_FOUND` |
| Already sold (status or active sale) | 409 | `VEHICLE_ALREADY_SOLD` |
| Invalid body / price / payment / customer | 400 | `INVALID_REQUEST` |

---

### POST /api/v1/vehicle/:id/showroom — Assign Vehicle to Showroom

**Flow:**
1. `POST /api/v1/vehicle/:id/showroom` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.AssignShowroom`
2. Handler:
   - Parse `:id` (uint64, must be > 0)
   - `ShouldBindJSON` → `AssignShowroomRequest` (`showroom_id` is required, must be > 0)
   - Extract `middleware.ContextKeyShowroomRoles` from context
   - Authorisation: caller must be a member of `req.ShowroomID` **and** must not be `employee` role → 403 `FORBIDDEN` otherwise
   - Call `service.AssignVehicleToShowroom(ctx, id, req.ShowroomID)` → 201 on success
3. Service:
   - `repo.VehicleExistsByID(ctx, vehicleID)` — `SELECT COUNT(*) FROM vehicles WHERE id = ? AND deleted_at IS NULL`; if 0 → `ErrVehicleNotFound`
   - `repo.GetVehicleShowroomID(ctx, vehicleID)` — if it returns `nil` error the vehicle already has a showroom → `ErrVehicleAlreadyInShowroom`; if it returns `ErrVehicleNotFound` proceed; any other error propagates
   - `repo.AssignShowroom(ctx, vehicleID, showroomID)` — GORM `Create` on `vehicle_showroom_relations`
4. Response: `201 Created` with `AssignShowroomResponse`

**Authorization:**
- Only `owner` and `manager` roles can assign a vehicle to their showroom.
- `employee` role → 403 `FORBIDDEN`.
- User must be a member of the target showroom (not any showroom).

**Request Body:**
| Field | Required | Type | Notes |
|-------|----------|------|-------|
| `showroom_id` | Yes | uint64 | Must be > 0; user must be owner/manager of this showroom |

**Error Codes:**
| Scenario | HTTP | Code |
|----------|------|------|
| Zero or invalid vehicle ID | 400 | `INVALID_REQUEST` |
| Zero `showroom_id` | 400 | `INVALID_REQUEST` |
| Caller not a member of target showroom | 403 | `FORBIDDEN` |
| Caller is `employee` role | 403 | `FORBIDDEN` |
| Vehicle does not exist | 404 | `VEHICLE_NOT_FOUND` |
| Vehicle already assigned to a showroom | 409 | `VEHICLE_ALREADY_IN_SHOWROOM` |

---

### POST /api/v1/vehicle/:id/image — Upload Vehicle Photo

**Flow:**
1. `POST /api/v1/vehicle/:id/image` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.AddVehicleImage`
2. Handler:
   - Extracts caller `userID`, parses `:id`, loads showroom roles
   - `GetVehicleShowroomID` + membership check (any role: owner/manager/employee). Non-member → 404 `VEHICLE_NOT_FOUND`
   - Parses `multipart/form-data` (20 MB form limit). Required file field `photo`; required form field `label`
3. Service:
   - Validates `label` ∈ `front`, `interior`, `exterior`, `back`, `wheel`
   - Validates file: `.jpg`/`.jpeg`/`.png`, size ≤ 15 MB
   - `GetCurrentStatus` — sold → 422 `VEHICLE_UPDATE_FORBIDDEN`
   - Uploads via `storage.Provider`. Object key: `{userID}/vehicle/{vehicleID}/{YYYYMMDDHHmmss}{ext}`
   - Persists `vehicle_images` with the **object key** in `image_url` and `uploaded_by` = caller. Upload failure fails the request (no DB row).
   - Response `url` is a signed GET URL (TTL from `STORAGE_SIGNED_URL_TTL_SECONDS`, default 1 hour). Signing failure returns an empty `url`.
4. Multiple photos with the same label are allowed (no unique constraint).
5. Response: `201 Created` with `AddVehicleImageResponse`

**Error Codes:**
| Scenario | HTTP | Code |
|----------|------|------|
| Missing/invalid file, label, or multipart body | 400 | `INVALID_REQUEST` |
| File larger than 15 MB | 400 | `FILE_TOO_LARGE` |
| Extension not jpg/jpeg/png | 400 | `INVALID_FILE_TYPE` |
| Vehicle not found / not a showroom member | 404 | `VEHICLE_NOT_FOUND` |
| Vehicle is sold | 422 | `VEHICLE_UPDATE_FORBIDDEN` |
| Storage upload failure | 500 | `INTERNAL` |

---

### DELETE /api/v1/vehicle/:id/image/:image_id — Soft-delete Vehicle Photo

**Flow:**
1. `DELETE /api/v1/vehicle/:id/image/:image_id` → `RequireDeviceContext` → `RequireAuth` → `ShowroomRoles` → `vehicle.Handler.DeleteVehicleImage`
2. Handler: same membership gate as upload (any showroom member).
3. Service:
   - `GetCurrentStatus` — sold → 422 `VEHICLE_UPDATE_FORBIDDEN`
   - `repo.SoftDeleteImage` sets `deleted_at` where `id` and `vehicle_id` match. No GCS object delete.
   - Missing row → 404 `VEHICLE_IMAGE_NOT_FOUND`
4. Response: `200 OK` with message `vehicle image deleted`

**Error Codes:**
| Scenario | HTTP | Code |
|----------|------|------|
| Invalid vehicle or image id | 400 | `INVALID_REQUEST` |
| Vehicle not found / not a showroom member | 404 | `VEHICLE_NOT_FOUND` |
| Image not found for this vehicle | 404 | `VEHICLE_IMAGE_NOT_FOUND` |
| Vehicle is sold | 422 | `VEHICLE_UPDATE_FORBIDDEN` |

---

## Documentation Update Checklist

- Update this file when vehicle behavior, schema assumptions, or APIs change.
- For API or function behavior changes, add/update flow details: route entry, middleware, handler/service path, and response outcomes.
