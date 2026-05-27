package dashboard_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"infiour.local/dms-api-server/internal/modules/dashboard"
)

// fakeRepo is a controllable in-memory implementation of the dashboardRepo interface.
type fakeRepo struct {
	salesResult    *dashboard.SalesQueryResult
	salesErr       error
	inventoryResult *dashboard.InventoryQueryResult
	inventoryErr   error
	expenseResult  *dashboard.ExpenseQueryResult
	expenseErr     error
	topTypesResult []dashboard.VehicleTypeQueryResult
	topTypesErr    error

	capturedSalesParams     dashboard.QueryParams
	capturedInventoryParams dashboard.QueryParams
	capturedExpenseParams   dashboard.QueryParams
	capturedTopTypesParams  dashboard.QueryParams
}

func (f *fakeRepo) FetchSalesSummary(_ context.Context, params dashboard.QueryParams) (*dashboard.SalesQueryResult, error) {
	f.capturedSalesParams = params
	if f.salesErr != nil {
		return nil, f.salesErr
	}
	if f.salesResult != nil {
		return f.salesResult, nil
	}
	return &dashboard.SalesQueryResult{}, nil
}

func (f *fakeRepo) FetchInventorySummary(_ context.Context, params dashboard.QueryParams) (*dashboard.InventoryQueryResult, error) {
	f.capturedInventoryParams = params
	if f.inventoryErr != nil {
		return nil, f.inventoryErr
	}
	if f.inventoryResult != nil {
		return f.inventoryResult, nil
	}
	return &dashboard.InventoryQueryResult{}, nil
}

func (f *fakeRepo) FetchExpenseSummary(_ context.Context, params dashboard.QueryParams) (*dashboard.ExpenseQueryResult, error) {
	f.capturedExpenseParams = params
	if f.expenseErr != nil {
		return nil, f.expenseErr
	}
	if f.expenseResult != nil {
		return f.expenseResult, nil
	}
	return &dashboard.ExpenseQueryResult{}, nil
}

func (f *fakeRepo) FetchTopVehicleTypes(_ context.Context, params dashboard.QueryParams) ([]dashboard.VehicleTypeQueryResult, error) {
	f.capturedTopTypesParams = params
	if f.topTypesErr != nil {
		return nil, f.topTypesErr
	}
	return f.topTypesResult, nil
}

func newService(repo *fakeRepo) dashboard.Service {
	return dashboard.NewService(repo)
}

// --- Duration parsing ---

func TestGetDashboardDefaultsToLifetime(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.capturedSalesParams.From != nil {
		t.Fatalf("expected nil From for lifetime, got %v", repo.capturedSalesParams.From)
	}
}

func TestGetDashboardLifetimeDuration(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.capturedSalesParams.From != nil {
		t.Fatalf("expected nil From for lifetime, got %v", repo.capturedSalesParams.From)
	}
}

func TestGetDashboardDurationWindows(t *testing.T) {
	cases := []struct {
		duration string
		days     int
	}{
		{"1w", 7},
		{"1m", 30},
		{"3m", 90},
		{"6m", 180},
		{"12m", 365},
	}

	for _, tc := range cases {
		t.Run(tc.duration, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := newService(repo)
			before := time.Now()

			_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: tc.duration})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.capturedSalesParams.From == nil {
				t.Fatalf("expected non-nil From for duration %s", tc.duration)
			}

			after := time.Now()
			expectedLow := before.AddDate(0, 0, -tc.days).Add(-time.Second)
			expectedHigh := after.AddDate(0, 0, -tc.days).Add(time.Second)
			if repo.capturedSalesParams.From.Before(expectedLow) || repo.capturedSalesParams.From.After(expectedHigh) {
				t.Fatalf("From out of expected range for duration %s: got %v", tc.duration, repo.capturedSalesParams.From)
			}
		})
	}
}

func TestGetDashboardInvalidDurationReturnsError(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "2y"})
	if !errors.Is(err, dashboard.ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration, got %v", err)
	}
}

// --- Showroom filter passthrough ---

func TestGetDashboardPassesShowroomIDToRepo(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)
	id := uint64(42)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime", ShowroomID: &id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.capturedSalesParams.ShowroomID == nil || *repo.capturedSalesParams.ShowroomID != 42 {
		t.Fatalf("expected ShowroomID 42 to be passed through, got %v", repo.capturedSalesParams.ShowroomID)
	}
}

func TestGetDashboardNilShowroomIDPassedThrough(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.capturedSalesParams.ShowroomID != nil {
		t.Fatalf("expected nil ShowroomID, got %v", repo.capturedSalesParams.ShowroomID)
	}
}

// --- Average calculations ---

func TestGetDashboardAverageProfitPerSale(t *testing.T) {
	repo := &fakeRepo{
		salesResult: &dashboard.SalesQueryResult{VehiclesSold: 4, TotalRevenue: 400000, NetProfit: 80000},
	}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SalesSummary.AverageProfitPerSale != 20000 {
		t.Fatalf("expected 20000, got %v", resp.SalesSummary.AverageProfitPerSale)
	}
}

func TestGetDashboardAverageProfitPerSaleZeroWhenNoSales(t *testing.T) {
	repo := &fakeRepo{
		salesResult: &dashboard.SalesQueryResult{VehiclesSold: 0, TotalRevenue: 0, NetProfit: 0},
	}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SalesSummary.AverageProfitPerSale != 0 {
		t.Fatalf("expected 0, got %v", resp.SalesSummary.AverageProfitPerSale)
	}
}

func TestGetDashboardAverageExpensePerVehicle(t *testing.T) {
	repo := &fakeRepo{
		inventoryResult: &dashboard.InventoryQueryResult{InventoryCount: 5},
		expenseResult:   &dashboard.ExpenseQueryResult{TotalExpenses: 50000},
	}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExpenseSummary.AverageExpensePerVehicle != 10000 {
		t.Fatalf("expected 10000, got %v", resp.ExpenseSummary.AverageExpensePerVehicle)
	}
}

func TestGetDashboardAverageExpensePerVehicleZeroWhenNoInventory(t *testing.T) {
	repo := &fakeRepo{
		inventoryResult: &dashboard.InventoryQueryResult{InventoryCount: 0},
		expenseResult:   &dashboard.ExpenseQueryResult{TotalExpenses: 5000},
	}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExpenseSummary.AverageExpensePerVehicle != 0 {
		t.Fatalf("expected 0, got %v", resp.ExpenseSummary.AverageExpensePerVehicle)
	}
}

// --- Zero data handling ---

func TestGetDashboardZeroDataReturnsZeroValues(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SalesSummary.VehiclesSold != 0 {
		t.Fatalf("expected 0 vehicles sold")
	}
	if resp.SalesSummary.TotalRevenue != 0 {
		t.Fatalf("expected 0 total revenue")
	}
	if resp.SalesSummary.NetProfit != 0 {
		t.Fatalf("expected 0 net profit")
	}
	if resp.SalesSummary.AverageProfitPerSale != 0 {
		t.Fatalf("expected 0 average profit per sale")
	}
	if resp.InventorySummary.InventoryCount != 0 {
		t.Fatalf("expected 0 inventory count")
	}
	if resp.ExpenseSummary.TotalExpenses != 0 {
		t.Fatalf("expected 0 total expenses")
	}
	if len(resp.TopVehicleTypes) != 0 {
		t.Fatalf("expected empty top vehicle types")
	}
}

// --- Top vehicle types ---

func TestGetDashboardTopVehicleTypesOrdered(t *testing.T) {
	repo := &fakeRepo{
		topTypesResult: []dashboard.VehicleTypeQueryResult{
			{VehicleType: "car", VehiclesSold: 8, NetProfit: 500000},
			{VehicleType: "bike", VehiclesSold: 5, NetProfit: 300000},
		},
	}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.TopVehicleTypes) != 2 {
		t.Fatalf("expected 2 vehicle types, got %d", len(resp.TopVehicleTypes))
	}
	if resp.TopVehicleTypes[0].VehicleType != "car" {
		t.Fatalf("expected car first, got %s", resp.TopVehicleTypes[0].VehicleType)
	}
	if resp.TopVehicleTypes[0].VehiclesSold != 8 {
		t.Fatalf("expected 8, got %d", resp.TopVehicleTypes[0].VehiclesSold)
	}
	if resp.TopVehicleTypes[1].VehicleType != "bike" {
		t.Fatalf("expected bike second, got %s", resp.TopVehicleTypes[1].VehicleType)
	}
}

func TestGetDashboardTopVehicleTypesEmptyWhenNoSales(t *testing.T) {
	repo := &fakeRepo{topTypesResult: []dashboard.VehicleTypeQueryResult{}}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.TopVehicleTypes) != 0 {
		t.Fatalf("expected empty top vehicle types")
	}
}

// --- Repository error propagation ---

func TestGetDashboardPropagatesSalesError(t *testing.T) {
	sentinel := errors.New("sales db error")
	repo := &fakeRepo{salesErr: sentinel}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestGetDashboardPropagatesInventoryError(t *testing.T) {
	sentinel := errors.New("inventory db error")
	repo := &fakeRepo{inventoryErr: sentinel}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestGetDashboardPropagatesExpenseError(t *testing.T) {
	sentinel := errors.New("expense db error")
	repo := &fakeRepo{expenseErr: sentinel}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestGetDashboardPropagatesTopTypesError(t *testing.T) {
	sentinel := errors.New("top types db error")
	repo := &fakeRepo{topTypesErr: sentinel}
	svc := newService(repo)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "lifetime"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

// --- Response field mapping ---

func TestGetDashboardResponseFieldsMapCorrectly(t *testing.T) {
	repo := &fakeRepo{
		salesResult:     &dashboard.SalesQueryResult{VehiclesSold: 10, TotalRevenue: 1000000, NetProfit: 200000},
		inventoryResult: &dashboard.InventoryQueryResult{InventoryCount: 20, InventoryValue: 5000000, DeadStockCount: 3, AverageInventoryAgeDays: 45.5},
		expenseResult:   &dashboard.ExpenseQueryResult{TotalExpenses: 60000},
		topTypesResult:  []dashboard.VehicleTypeQueryResult{{VehicleType: "car", VehiclesSold: 10, NetProfit: 200000}},
	}
	svc := newService(repo)

	resp, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardRequest{Duration: "1m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.SalesSummary.VehiclesSold != 10 {
		t.Errorf("VehiclesSold: expected 10, got %d", resp.SalesSummary.VehiclesSold)
	}
	if resp.SalesSummary.TotalRevenue != 1000000 {
		t.Errorf("TotalRevenue: expected 1000000, got %v", resp.SalesSummary.TotalRevenue)
	}
	if resp.SalesSummary.NetProfit != 200000 {
		t.Errorf("NetProfit: expected 200000, got %v", resp.SalesSummary.NetProfit)
	}
	if resp.SalesSummary.AverageProfitPerSale != 20000 {
		t.Errorf("AverageProfitPerSale: expected 20000, got %v", resp.SalesSummary.AverageProfitPerSale)
	}
	if resp.InventorySummary.InventoryCount != 20 {
		t.Errorf("InventoryCount: expected 20, got %d", resp.InventorySummary.InventoryCount)
	}
	if resp.InventorySummary.InventoryValue != 5000000 {
		t.Errorf("InventoryValue: expected 5000000, got %v", resp.InventorySummary.InventoryValue)
	}
	if resp.InventorySummary.DeadStockCount != 3 {
		t.Errorf("DeadStockCount: expected 3, got %d", resp.InventorySummary.DeadStockCount)
	}
	if resp.InventorySummary.AverageInventoryAgeDays != 45.5 {
		t.Errorf("AverageInventoryAgeDays: expected 45.5, got %v", resp.InventorySummary.AverageInventoryAgeDays)
	}
	if resp.ExpenseSummary.TotalExpenses != 60000 {
		t.Errorf("TotalExpenses: expected 60000, got %v", resp.ExpenseSummary.TotalExpenses)
	}
	if resp.ExpenseSummary.AverageExpensePerVehicle != 3000 {
		t.Errorf("AverageExpensePerVehicle: expected 3000, got %v", resp.ExpenseSummary.AverageExpensePerVehicle)
	}
}
