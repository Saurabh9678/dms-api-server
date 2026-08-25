package dashboard

import (
	"context"
	"time"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/pkg/inventory"
)

// QueryParams carries filter values shared across all repo queries.
type QueryParams struct {
	From       *time.Time // nil = no date filter (lifetime)
	ShowroomID uint64     // required; inventory/sales/expenses scoped to this showroom
}

// SalesQueryResult holds raw aggregate sales data returned by the repository.
type SalesQueryResult struct {
	VehiclesSold int64
	TotalRevenue float64
	NetProfit    float64
}

// InventoryQueryResult holds raw aggregate inventory data returned by the repository.
type InventoryQueryResult struct {
	InventoryCount          int64
	InventoryValue          float64
	DeadStockCount          int64
	AverageInventoryAgeDays float64
}

// ExpenseQueryResult holds raw aggregate expense data returned by the repository.
type ExpenseQueryResult struct {
	TotalExpenses float64
}

// VehicleTypeQueryResult holds per-vehicle-type aggregate data returned by the repository.
type VehicleTypeQueryResult struct {
	VehicleType  string
	VehiclesSold int64
	NetProfit    float64
}

type dashboardRepo interface {
	FetchSalesSummary(ctx context.Context, params QueryParams) (*SalesQueryResult, error)
	FetchInventorySummary(ctx context.Context, params QueryParams) (*InventoryQueryResult, error)
	FetchExpenseSummary(ctx context.Context, params QueryParams) (*ExpenseQueryResult, error)
	FetchTopVehicleTypes(ctx context.Context, params QueryParams) ([]VehicleTypeQueryResult, error)
}

// Repository implements dashboardRepo using raw SQL aggregations.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func showroomScopeSQL() string {
	return ` AND v.id IN (SELECT vehicle_id FROM vehicle_showroom_relations WHERE showroom_id = ? AND deleted_at IS NULL)`
}

func (r *Repository) FetchSalesSummary(ctx context.Context, params QueryParams) (*SalesQueryResult, error) {
	q := `
		SELECT
			COUNT(cvs.id) AS vehicles_sold,
			COALESCE(SUM(cvs.sale_price), 0) AS total_revenue,
			COALESCE(SUM(cvs.sale_price - COALESCE(vp.buying_price, 0) - COALESCE(exp_totals.total_exp, 0)), 0) AS net_profit
		FROM customer_vehicle_sales cvs
		INNER JOIN vehicles v ON v.id = cvs.vehicle_id AND v.deleted_at IS NULL
		LEFT JOIN vehicle_pricing vp ON vp.vehicle_id = cvs.vehicle_id AND vp.deleted_at IS NULL
		LEFT JOIN (
			SELECT vehicle_id, SUM(amount) AS total_exp
			FROM vehicle_expenses
			WHERE deleted_at IS NULL
			GROUP BY vehicle_id
		) exp_totals ON exp_totals.vehicle_id = cvs.vehicle_id
		WHERE cvs.deleted_at IS NULL` + showroomScopeSQL()

	args := []interface{}{params.ShowroomID}
	if params.From != nil {
		q += ` AND cvs.sale_date >= ?`
		args = append(args, params.From)
	}

	var result SalesQueryResult
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) FetchInventorySummary(ctx context.Context, params QueryParams) (*InventoryQueryResult, error) {
	q := `
		SELECT
			COUNT(v.id) AS inventory_count,
			COALESCE(SUM(vp.buying_price), 0) AS inventory_value,
			COALESCE(SUM(` + inventory.DeadStockCaseSQL("vp.buying_date") + `), 0) AS dead_stock_count,
			COALESCE(AVG(CASE WHEN vp.buying_date IS NOT NULL THEN EXTRACT(EPOCH FROM (NOW() - vp.buying_date)) / 86400 ELSE NULL END), 0) AS average_inventory_age_days
		FROM vehicles v
		INNER JOIN vehicle_showroom_relations vsr
			ON vsr.vehicle_id = v.id
			AND vsr.showroom_id = ?
			AND vsr.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT buying_price, buying_date
			FROM vehicle_pricing
			WHERE vehicle_id = v.id AND deleted_at IS NULL
			ORDER BY id DESC
			LIMIT 1
		) vp ON true
		WHERE v.deleted_at IS NULL
		  AND v.id NOT IN (SELECT vehicle_id FROM customer_vehicle_sales WHERE deleted_at IS NULL)`

	args := []interface{}{params.ShowroomID}

	var result InventoryQueryResult
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) FetchExpenseSummary(ctx context.Context, params QueryParams) (*ExpenseQueryResult, error) {
	q := `
		SELECT
			COALESCE(SUM(ve.amount), 0) AS total_expenses
		FROM vehicle_expenses ve
		INNER JOIN vehicles v ON v.id = ve.vehicle_id AND v.deleted_at IS NULL
		WHERE ve.deleted_at IS NULL` + showroomScopeSQL()

	args := []interface{}{params.ShowroomID}
	if params.From != nil {
		q += ` AND ve.date >= ?`
		args = append(args, params.From)
	}

	var result ExpenseQueryResult
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) FetchTopVehicleTypes(ctx context.Context, params QueryParams) ([]VehicleTypeQueryResult, error) {
	q := `
		SELECT
			v.type AS vehicle_type,
			COUNT(cvs.id) AS vehicles_sold,
			COALESCE(SUM(cvs.sale_price - COALESCE(vp.buying_price, 0) - COALESCE(exp_totals.total_exp, 0)), 0) AS net_profit
		FROM customer_vehicle_sales cvs
		INNER JOIN vehicles v ON v.id = cvs.vehicle_id AND v.deleted_at IS NULL
		LEFT JOIN vehicle_pricing vp ON vp.vehicle_id = cvs.vehicle_id AND vp.deleted_at IS NULL
		LEFT JOIN (
			SELECT vehicle_id, SUM(amount) AS total_exp
			FROM vehicle_expenses
			WHERE deleted_at IS NULL
			GROUP BY vehicle_id
		) exp_totals ON exp_totals.vehicle_id = cvs.vehicle_id
		WHERE cvs.deleted_at IS NULL` + showroomScopeSQL()

	args := []interface{}{params.ShowroomID}
	if params.From != nil {
		q += ` AND cvs.sale_date >= ?`
		args = append(args, params.From)
	}

	q += ` GROUP BY v.type ORDER BY vehicles_sold DESC`

	var results []VehicleTypeQueryResult
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
