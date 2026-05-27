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
| `showroom_id` | No | Any valid showroom ID (uint) | all showrooms |

**Duration semantics:**
- Applies to sales analytics (`customer_vehicle_sales.sale_date`) and expense analytics (`vehicle_expenses.date`)
- Inventory metrics always reflect current state — not duration-filtered

#### Handler flow

1. Handler reads `duration` (default `lifetime`) and `showroom_id` (optional) from query params
2. If `showroom_id` is non-empty but unparseable → 400 `INVALID_REQUEST`
3. Handler calls `service.GetDashboard(ctx, GetDashboardRequest{Duration, ShowroomID})`
4. On success → 200 with envelope message `"dashboard data fetched"`
5. On error → `response.FromError` (400 for invalid duration, 500 for internal errors)

#### Service flow

1. Empty duration defaults to `"lifetime"`
2. Duration is validated and mapped to a `*time.Time` window start (`nil` = no date filter)
3. Invalid duration → `ErrInvalidDuration` (400)
4. Four parallel repo queries:
   - `FetchSalesSummary`: sales count, total revenue, net profit (duration-filtered by `sale_date`)
   - `FetchInventorySummary`: inventory count/value, dead stock, avg age (no duration filter)
   - `FetchExpenseSummary`: total operational expenses (duration-filtered by `expense.date`)
   - `FetchTopVehicleTypes`: per-type vehicles sold and net profit (duration-filtered by `sale_date`)
5. Computes `average_profit_per_sale = net_profit / vehicles_sold` (0 if no sales)
6. Computes `average_expense_per_vehicle = total_expenses / inventory_count` (0 if no inventory)

#### Business rules

**Sales-anchored profit model:**
- Only SOLD vehicles contribute to revenue and profit
- Buying inventory is asset conversion, not a realized loss
- `net_profit = SUM(sale_price - buying_price - all_vehicle_expenses)` for sold vehicles in period

**Inventory:**
- Unsold vehicles = vehicles NOT present in `customer_vehicle_sales` (active records only)
- `inventory_value = SUM(buying_price)` of unsold vehicles
- `dead_stock_count` = unsold vehicles where age > 90 days (based on `buying_date`)
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
