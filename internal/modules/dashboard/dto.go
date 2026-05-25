package dashboard

type GetDashboardRequest struct {
	Duration   string
	ShowroomID *uint64
}

type SalesSummary struct {
	VehiclesSold         int64   `json:"vehicles_sold"`
	TotalRevenue         float64 `json:"total_revenue"`
	NetProfit            float64 `json:"net_profit"`
	AverageProfitPerSale float64 `json:"average_profit_per_sale"`
}

type InventorySummary struct {
	InventoryCount          int64   `json:"inventory_count"`
	InventoryValue          float64 `json:"inventory_value"`
	DeadStockCount          int64   `json:"dead_stock_count"`
	AverageInventoryAgeDays float64 `json:"average_inventory_age_days"`
}

type ExpenseSummary struct {
	TotalExpenses            float64 `json:"total_expenses"`
	AverageExpensePerVehicle float64 `json:"average_expense_per_vehicle"`
}

type VehicleTypeMetrics struct {
	VehicleType  string  `json:"vehicle_type"`
	VehiclesSold int64   `json:"vehicles_sold"`
	NetProfit    float64 `json:"net_profit"`
}

type DashboardResponse struct {
	SalesSummary     SalesSummary         `json:"sales_summary"`
	InventorySummary InventorySummary     `json:"inventory_summary"`
	ExpenseSummary   ExpenseSummary       `json:"expense_summary"`
	TopVehicleTypes  []VehicleTypeMetrics `json:"top_vehicle_types"`
}
