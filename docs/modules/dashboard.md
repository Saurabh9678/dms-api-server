# Dashboard Module

Executive overview dashboard for dealership health metrics.

## Purpose

- Sales performance summary
- Inventory visibility
- Expense visibility
- Vehicle category insights (top vehicle types)

## Endpoint

### `GET /api/v1/dashboard`

Protected endpoint. Requires `Authorization: Bearer <accessToken>`, `X-Platform`, and `X-Device-Id` headers.

#### Query Parameters

| Param | Required | Values | Default |
|---|---|---|---|
| `duration` | No | `1w`, `1m`, `3m`, `6m`, `12m`, `lifetime` | `lifetime` |
| `showroom_id` | No | Numeric showroom ID (uint) | lowest-ID showroom where caller is `owner` or `manager` |

**Access:**
- Caller must be `owner` or `manager` of the resolved showroom; otherwise `403 FORBIDDEN`
- Employees cannot view the dashboard
- If `showroom_id` is omitted, the service selects the lowest numeric showroom ID among the caller's owner/manager memberships
- If the caller has no owner/manager showroom, `403 FORBIDDEN`

**Duration semantics:**
- Applies to sales analytics (`customer_vehicle_sales.sale_date`) and expense analytics (`vehicle_expenses.date`)
- Inventory metrics always reflect current state — not duration-filtered

#### Handler flow

1. `RequireDeviceContext` validates `X-Platform` and `X-Device-Id`.
2. `RequireAuth` validates the bearer token and sets user ID in context.
3. `RequireShowroomRoles` loads all active showroom roles for the authenticated user.
4. Handler reads `duration` (default `lifetime`) and `showroom_id` (optional) from query params.
5. If `showroom_id` is non-empty but unparseable → 400 `INVALID_REQUEST`.
6. Handler calls `service.GetDashboard(ctx, GetDashboardRequest{Duration, ShowroomID, ShowroomRoles})`.
7. On success → 200 with envelope message `"dashboard data fetched"`.
8. On error → `response.FromError` (400 for invalid duration, 403 for access, 500 for internal errors).

#### Service flow

1. Empty duration defaults to `"lifetime"`
2. Duration is validated and mapped to a `*time.Time` window start (`nil` = no date filter)
3. Invalid duration → `ErrInvalidDuration` (400)
4. Resolve showroom: explicit `showroom_id` if provided, else lowest owner/manager showroom ID; enforce owner/manager → else `ErrForbidden` (403)
5. Four parallel repo queries (all scoped to the resolved showroom):
   - `FetchSalesSummary`: sales count, total revenue, net profit (duration-filtered by `sale_date`)
   - `FetchInventorySummary`: inventory count/value, dead stock, avg age (no duration filter)
   - `FetchExpenseSummary`: total operational expenses (duration-filtered by `expense.date`)
   - `FetchTopVehicleTypes`: per-type vehicles sold and net profit (duration-filtered by `sale_date`)
6. Computes `average_profit_per_sale = net_profit / vehicles_sold` (0 if no sales)
7. Computes `average_expense_per_vehicle = total_expenses / inventory_count` (0 if no inventory)

#### Business rules

**Sales-anchored profit model:**
- Only SOLD vehicles contribute to revenue and profit
- Buying inventory is asset conversion, not a realized loss
- `net_profit = SUM(sale_price - buying_price - all_vehicle_expenses)` for sold vehicles in period

**Inventory:**
- Scoped via `INNER JOIN vehicle_showroom_relations` on the resolved `showroom_id`
- Unsold vehicles = vehicles NOT present in `customer_vehicle_sales` (active records only)
- Pricing uses latest `vehicle_pricing` row (`LATERAL … ORDER BY id DESC LIMIT 1`)
- `inventory_value = SUM(buying_price)` of unsold showroom vehicles (0 when no pricing rows)
- `dead_stock_count` = unsold vehicles where age > 90 days (based on `buying_date`), using shared helper `pkg/inventory` (`DeadStockCaseSQL` / `IsDeadStock`, threshold `DeadStockThresholdDays = 90`)
- `average_inventory_age_days` = AVG age of unsold vehicles with known buying dates

**Expenses:**
- Independent from sales — covers all operational costs during the period
- Includes repair, servicing, washing, transportation, accessories, maintenance

#### Response structure

```json
{
  "success": true,
  "message": "dashboard data fetched",
  "data": {
    "sales_summary": {
      "vehicles_sold": 15,
      "total_revenue": 3000000,
      "net_profit": 930000,
      "average_profit_per_sale": 62000
    },
    "inventory_summary": {
      "inventory_count": 45,
      "inventory_value": 12000000,
      "dead_stock_count": 6,
      "average_inventory_age_days": 38
    },
    "expense_summary": {
      "total_expenses": 70000,
      "average_expense_per_vehicle": 1555.56
    },
    "top_vehicle_types": [
      { "vehicle_type": "car",   "vehicles_sold": 8, "net_profit": 500000 },
      { "vehicle_type": "bike",  "vehicles_sold": 5, "net_profit": 300000 }
    ]
  }
}
```

`top_vehicle_types` contains only types with at least one sale, ordered by `vehicles_sold DESC`.

#### Error responses

| Condition | HTTP | Code |
|---|---|---|
| Invalid duration value | 400 | `INVALID_REQUEST` |
| Unparseable `showroom_id` | 400 | `INVALID_REQUEST` |
| Not owner/manager of resolved showroom (or no managed showroom) | 403 | `FORBIDDEN` |
| Internal/DB error | 500 | `INTERNAL_ERROR` |

## Tables Used

| Table | Purpose | Date Filter Column |
|---|---|---|
| `vehicles` | Join for type grouping and unsold check | — |
| `vehicle_pricing` | Buying price, buying date (inventory age) | `buying_date` |
| `vehicle_expenses` | Operational expenses | `date` |
| `customer_vehicle_sales` | Revenue, profit, sold status | `sale_date` |
| `vehicle_showroom_relations` | Showroom scope filter | — |
| `showrooms` | Showroom existence | — |
