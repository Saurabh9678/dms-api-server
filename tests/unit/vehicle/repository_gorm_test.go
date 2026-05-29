package vehicle_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/internal/modules/vehicle"
)

func newVehicleMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func TestVehicleModelTableNames(t *testing.T) {
	assert.Equal(t, "vehicles", vehicle.Vehicle{}.TableName())
	assert.Equal(t, "vehicle_documents", vehicle.VehicleDocument{}.TableName())
	assert.Equal(t, "vehicle_expenses", vehicle.VehicleExpenses{}.TableName())
	assert.Equal(t, "vehicle_images", vehicle.VehicleImage{}.TableName())
	assert.Equal(t, "vehicle_showroom_relations", vehicle.VehicleShowroom{}.TableName())
	assert.Equal(t, "vehicle_pricing", vehicle.VehiclePricing{}.TableName())
	assert.Equal(t, "vehicle_statuses", vehicle.VehicleStatus{}.TableName())
}

func TestVehicleRepositoryCreateSuccess(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "vehicles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	mock.ExpectCommit()

	v := &vehicle.Vehicle{VehicleType: "sedan"}
	result, err := repo.Create(context.Background(), v)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestVehicleRepositoryCreateError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "vehicles"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), &vehicle.Vehicle{})
	assert.Error(t, err)
}

func TestVehicleRepositoryList_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	cols := []string{
		"id", "vehicle_type", "manufacturer", "model", "variant", "color",
		"year_of_manufacture", "rto_code", "registration_number", "registration_state",
		"usage_km", "fuel_type", "transmission_type", "created_at", "updated_at",
		"vs_status", "vs_started_at", "vp_buying_price", "vp_price_tag", "vp_currency", "vp_tagged_at",
	}

	rows := sqlmock.NewRows(cols).AddRow(
		uint64(1), "car", "Toyota", "Camry", "LE", "Black",
		2020, "KA-01", "KA01AB1234", "Karnataka",
		50000, "petrol", "manual", now, now,
		"ready_for_sale", now, 200000.0, 300000.0, "inr", now,
	)

	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	filter := vehicle.ListFilter{
		Statuses: []vehicle.VehicleStatusType{vehicle.VehicleStatusTypeReadyForSale},
		Page:     1,
		Limit:    20,
	}

	results, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, uint64(1), results[0].ID)
	assert.NotNil(t, results[0].CurrentStatus)
	assert.NotNil(t, results[0].CurrentPricing)
}

func TestVehicleRepositoryList_NoStatusOrPricing(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	cols := []string{
		"id", "vehicle_type", "manufacturer", "model", "variant", "color",
		"year_of_manufacture", "rto_code", "registration_number", "registration_state",
		"usage_km", "fuel_type", "transmission_type", "created_at", "updated_at",
		"vs_status", "vs_started_at", "vp_buying_price", "vp_price_tag", "vp_currency", "vp_tagged_at",
	}

	rows := sqlmock.NewRows(cols).AddRow(
		uint64(2), "bike", "Honda", "CB", "STD", "Red",
		2021, "MH-01", "MH01CD5678", "Maharashtra",
		10000, "petrol", "manual", now, now,
		nil, nil, nil, nil, nil, nil,
	)

	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	filter := vehicle.ListFilter{
		Statuses:     []vehicle.VehicleStatusType{vehicle.VehicleStatusTypeGarage},
		VehicleTypes: []vehicle.VehicleType{vehicle.VehicleTypeBike},
		Page:         1,
		Limit:        10,
	}

	results, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Nil(t, results[0].CurrentStatus)
	assert.Nil(t, results[0].CurrentPricing)
}

func TestVehicleRepositoryList_WithPriceFilter(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	minP := 100000.0
	maxP := 500000.0
	filter := vehicle.ListFilter{
		Statuses: []vehicle.VehicleStatusType{vehicle.VehicleStatusTypeReadyForSale},
		MinPrice: &minP,
		MaxPrice: &maxP,
		Page:     1,
		Limit:    20,
	}

	results, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestVehicleRepositoryList_Error(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrInvalidData)

	filter := vehicle.ListFilter{
		Statuses: []vehicle.VehicleStatusType{vehicle.VehicleStatusTypeReadyForSale},
		Page:     1,
		Limit:    20,
	}

	_, err := repo.List(context.Background(), filter)
	assert.Error(t, err)
}

func TestVehicleRepositoryCountByType_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	rows := sqlmock.NewRows([]string{"vehicle_type", "count"}).
		AddRow("car", int64(5)).
		AddRow("bike", int64(3))

	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	filter := vehicle.ListFilter{
		Statuses: []vehicle.VehicleStatusType{vehicle.VehicleStatusTypeReadyForSale},
		Page:     1,
		Limit:    20,
	}

	counts, err := repo.CountByType(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), counts[vehicle.VehicleTypeCar])
	assert.Equal(t, int64(3), counts[vehicle.VehicleTypeBike])
}

func TestVehicleRepositoryCountByType_Error(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrInvalidData)

	filter := vehicle.ListFilter{
		Statuses: []vehicle.VehicleStatusType{vehicle.VehicleStatusTypeReadyForSale},
		Page:     1,
		Limit:    20,
	}

	_, err := repo.CountByType(context.Background(), filter)
	assert.Error(t, err)
}
