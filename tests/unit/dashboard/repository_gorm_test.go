package dashboard_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/internal/modules/dashboard"
)

func newDashboardMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return gormDB, mock
}

func TestFetchSalesSummarySuccess(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`customer_vehicle_sales`).
		WillReturnRows(sqlmock.NewRows([]string{"vehicles_sold", "total_revenue", "net_profit"}).
			AddRow(int64(5), float64(1000000), float64(200000)))

	result, err := repo.FetchSalesSummary(context.Background(), dashboard.QueryParams{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestFetchSalesSummaryError(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`customer_vehicle_sales`).WillReturnError(gorm.ErrInvalidData)

	_, err := repo.FetchSalesSummary(context.Background(), dashboard.QueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchInventorySummarySuccess(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`inventory_count`).
		WillReturnRows(sqlmock.NewRows([]string{"inventory_count", "inventory_value", "dead_stock_count", "average_inventory_age_days"}).
			AddRow(int64(10), float64(5000000), int64(2), float64(30)))

	result, err := repo.FetchInventorySummary(context.Background(), dashboard.QueryParams{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestFetchInventorySummaryError(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`inventory_count`).WillReturnError(gorm.ErrInvalidData)

	_, err := repo.FetchInventorySummary(context.Background(), dashboard.QueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchExpenseSummarySuccess(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`vehicle_expenses`).
		WillReturnRows(sqlmock.NewRows([]string{"total_expenses"}).AddRow(float64(50000)))

	result, err := repo.FetchExpenseSummary(context.Background(), dashboard.QueryParams{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestFetchExpenseSummaryError(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`vehicle_expenses`).WillReturnError(gorm.ErrInvalidData)

	_, err := repo.FetchExpenseSummary(context.Background(), dashboard.QueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchTopVehicleTypesSuccess(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`vehicle_type`).
		WillReturnRows(sqlmock.NewRows([]string{"vehicle_type", "vehicles_sold", "net_profit"}).
			AddRow("car", int64(3), float64(150000)))

	results, err := repo.FetchTopVehicleTypes(context.Background(), dashboard.QueryParams{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestFetchTopVehicleTypesError(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	mock.ExpectQuery(`vehicle_type`).WillReturnError(gorm.ErrInvalidData)

	_, err := repo.FetchTopVehicleTypes(context.Background(), dashboard.QueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchSalesSummaryWithFilters(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	now := time.Now()
	showroomID := uint64(5)
	mock.ExpectQuery(`customer_vehicle_sales`).
		WillReturnRows(sqlmock.NewRows([]string{"vehicles_sold", "total_revenue", "net_profit"}).
			AddRow(int64(2), float64(500000), float64(100000)))

	result, err := repo.FetchSalesSummary(context.Background(), dashboard.QueryParams{From: &now, ShowroomID: &showroomID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestFetchInventorySummaryWithFilters(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	now := time.Now()
	showroomID := uint64(5)
	mock.ExpectQuery(`inventory_count`).
		WillReturnRows(sqlmock.NewRows([]string{"inventory_count", "inventory_value", "dead_stock_count", "average_inventory_age_days"}).
			AddRow(int64(3), float64(1500000), int64(0), float64(15)))

	result, err := repo.FetchInventorySummary(context.Background(), dashboard.QueryParams{From: &now, ShowroomID: &showroomID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestFetchExpenseSummaryWithFilters(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	now := time.Now()
	showroomID := uint64(5)
	mock.ExpectQuery(`vehicle_expenses`).
		WillReturnRows(sqlmock.NewRows([]string{"total_expenses"}).AddRow(float64(25000)))

	result, err := repo.FetchExpenseSummary(context.Background(), dashboard.QueryParams{From: &now, ShowroomID: &showroomID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestFetchTopVehicleTypesWithFilters(t *testing.T) {
	gormDB, mock := newDashboardMockDB(t)
	repo := dashboard.NewRepository(gormDB)

	now := time.Now()
	showroomID := uint64(5)
	mock.ExpectQuery(`vehicle_type`).
		WillReturnRows(sqlmock.NewRows([]string{"vehicle_type", "vehicles_sold", "net_profit"}).
			AddRow("bike", int64(1), float64(50000)))

	results, err := repo.FetchTopVehicleTypes(context.Background(), dashboard.QueryParams{From: &now, ShowroomID: &showroomID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
