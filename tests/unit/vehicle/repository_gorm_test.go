package vehicle_test

import (
	"context"
	"testing"

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
