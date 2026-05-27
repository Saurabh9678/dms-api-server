package dashboard

import (
	"context"
	"time"
)

// Service defines the dashboard use-case contract.
type Service interface {
	GetDashboard(ctx context.Context, req GetDashboardRequest) (*DashboardResponse, error)
}

type service struct {
	repo  dashboardRepo
	nowFn func() time.Time
}

func NewService(repo dashboardRepo) Service {
	return &service{repo: repo, nowFn: time.Now}
}

var durationDays = map[string]int{
	"1w":       7,
	"1m":       30,
	"3m":       90,
	"6m":       180,
	"12m":      365,
	"lifetime": 0,
}

func (s *service) parseDuration(d string) (*time.Time, error) {
	days, ok := durationDays[d]
	if !ok {
		return nil, ErrInvalidDuration
	}
	if days == 0 {
		return nil, nil
	}
	t := s.nowFn().AddDate(0, 0, -days)
	return &t, nil
}

func (s *service) GetDashboard(ctx context.Context, req GetDashboardRequest) (*DashboardResponse, error) {
	if req.Duration == "" {
		req.Duration = "lifetime"
	}

	from, err := s.parseDuration(req.Duration)
	if err != nil {
		return nil, err
	}

	params := QueryParams{From: from, ShowroomID: req.ShowroomID}

	sales, err := s.repo.FetchSalesSummary(ctx, params)
	if err != nil {
		return nil, err
	}

	inventory, err := s.repo.FetchInventorySummary(ctx, params)
	if err != nil {
		return nil, err
	}

	expenses, err := s.repo.FetchExpenseSummary(ctx, params)
	if err != nil {
		return nil, err
	}

	topTypes, err := s.repo.FetchTopVehicleTypes(ctx, params)
	if err != nil {
		return nil, err
	}

	var avgProfitPerSale float64
	if sales.VehiclesSold > 0 {
		avgProfitPerSale = sales.NetProfit / float64(sales.VehiclesSold)
	}

	var avgExpensePerVehicle float64
	if inventory.InventoryCount > 0 {
		avgExpensePerVehicle = expenses.TotalExpenses / float64(inventory.InventoryCount)
	}

	vehicleTypeMetrics := make([]VehicleTypeMetrics, len(topTypes))
	for i, t := range topTypes {
		vehicleTypeMetrics[i] = VehicleTypeMetrics(t)
	}

	return &DashboardResponse{
		SalesSummary: SalesSummary{
			VehiclesSold:         sales.VehiclesSold,
			TotalRevenue:         sales.TotalRevenue,
			NetProfit:            sales.NetProfit,
			AverageProfitPerSale: avgProfitPerSale,
		},
		InventorySummary: InventorySummary{
			InventoryCount:          inventory.InventoryCount,
			InventoryValue:          inventory.InventoryValue,
			DeadStockCount:          inventory.DeadStockCount,
			AverageInventoryAgeDays: inventory.AverageInventoryAgeDays,
		},
		ExpenseSummary: ExpenseSummary{
			TotalExpenses:            expenses.TotalExpenses,
			AverageExpensePerVehicle: avgExpensePerVehicle,
		},
		TopVehicleTypes: vehicleTypeMetrics,
	}, nil
}
