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

func TestGetByIDWithFullDetails_NotFound(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetByIDWithFullDetails(context.Background(), 999)
	assert.ErrorIs(t, err, vehicle.ErrVehicleNotFound)
}

func TestGetByIDWithFullDetails_DBError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetByIDWithFullDetails(context.Background(), 1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, vehicle.ErrVehicleNotFound)
}

func TestGetByIDWithFullDetails_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	vehicleCols := []string{
		"id", "vehicle_type", "manufacturer", "model", "variant", "color",
		"year_of_manufacture", "rto_code", "registration_number", "registration_state",
		"usage_km", "fuel_type", "transmission_type", "created_at", "updated_at", "deleted_at",
	}
	vehicleRows := sqlmock.NewRows(vehicleCols).
		AddRow(uint64(1), "car", "Toyota", "Camry", "LE", "Black", 2020, "KA-01", "KA01AB1234", "Karnataka", 50000, "petrol", "manual", now, now, nil)
	mock.ExpectQuery(`SELECT \* FROM "vehicles"`).WillReturnRows(vehicleRows)

	pricingCols := []string{"id", "vehicle_id", "buying_price", "buying_date", "price_tag", "tagged_at", "currency", "remarks", "created_at", "updated_at", "deleted_at"}
	pricingRows := sqlmock.NewRows(pricingCols).
		AddRow(uint64(1), uint64(1), 200000.0, now, 300000.0, now, "inr", "", now, now, nil)
	mock.ExpectQuery(`SELECT \* FROM "vehicle_pricing"`).WillReturnRows(pricingRows)

	statusCols := []string{"id", "vehicle_id", "status", "description", "started_at", "ended_at", "added_by", "created_at", "updated_at", "deleted_at"}
	statusRows := sqlmock.NewRows(statusCols).
		AddRow(uint64(1), uint64(1), "ready_for_sale", "", now, now, uint64(1), now, now, nil)
	mock.ExpectQuery(`SELECT \* FROM "vehicle_statuses"`).WillReturnRows(statusRows)

	docCols := []string{"id", "vehicle_id", "document_type", "document_url", "valid_from", "valid_till", "remarks", "uploaded_at", "uploaded_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(`SELECT \* FROM "vehicle_documents"`).WillReturnRows(sqlmock.NewRows(docCols))

	expCols := []string{"id", "vehicle_id", "status_id", "type", "amount", "paid_to", "description", "date", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(`SELECT \* FROM "vehicle_expenses"`).WillReturnRows(sqlmock.NewRows(expCols))

	imgCols := []string{"id", "vehicle_id", "image_url", "label", "uploaded_at", "uploaded_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(`SELECT \* FROM "vehicle_images"`).WillReturnRows(sqlmock.NewRows(imgCols))

	showroomCols := []string{"id", "vehicle_id", "showroom_id", "created_at", "updated_at", "deleted_at"}
	showroomRows := sqlmock.NewRows(showroomCols).AddRow(uint64(1), uint64(1), uint64(10), now, now, nil)
	mock.ExpectQuery(`SELECT \* FROM "vehicle_showroom_relations"`).WillReturnRows(showroomRows)

	saleCols := []string{
		"sale_price", "sale_date", "payment_mode", "receipt_url", "remarks",
		"customer_first_name", "customer_last_name", "customer_email",
		"customer_phone", "customer_address", "customer_city", "customer_state",
	}
	mock.ExpectQuery(`customer_vehicle_sales`).WillReturnRows(sqlmock.NewRows(saleCols))

	details, err := repo.GetByIDWithFullDetails(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, details)
	assert.Equal(t, uint64(1), details.Vehicle.ID)
	assert.NotNil(t, details.Pricing)
	assert.Equal(t, 1, len(details.Statuses))
	assert.Equal(t, uint64(10), details.ShowroomID)
	assert.Nil(t, details.SaleInfo)
}

func TestVehicleRepositoryPublicList_Success(t *testing.T) {
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
		"ready_for_sale", now, 200000.0, 350000.0, "inr", now,
	)
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	filter := vehicle.PublicListFilter{
		ShowroomID: 1,
		Page:       1,
		Limit:      20,
		SortBy:     "price_asc",
	}

	results, err := repo.PublicList(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, uint64(1), results[0].ID)
	assert.NotNil(t, results[0].CurrentPricing)
	assert.Equal(t, 350000.0, results[0].CurrentPricing.PriceTag)
}

func TestVehicleRepositoryPublicList_NilStatusAndPricing(t *testing.T) {
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

	filter := vehicle.PublicListFilter{ShowroomID: 1, Page: 1, Limit: 10, SortBy: "price_desc"}
	results, err := repo.PublicList(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Nil(t, results[0].CurrentStatus)
	assert.Nil(t, results[0].CurrentPricing)
}

func TestVehicleRepositoryPublicList_WithTypeFilter(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	filter := vehicle.PublicListFilter{
		ShowroomID:   1,
		VehicleTypes: []vehicle.VehicleType{vehicle.VehicleTypeCar},
		Page:         1,
		Limit:        20,
		SortBy:       "price_asc",
	}
	results, err := repo.PublicList(context.Background(), filter)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestVehicleRepositoryPublicList_WithPriceFilter(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	minP, maxP := 100000.0, 500000.0
	filter := vehicle.PublicListFilter{
		ShowroomID: 1,
		MinPrice:   &minP,
		MaxPrice:   &maxP,
		Page:       1,
		Limit:      20,
		SortBy:     "price_desc",
	}
	results, err := repo.PublicList(context.Background(), filter)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestVehicleRepositoryPublicList_Error(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrInvalidData)

	filter := vehicle.PublicListFilter{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}
	_, err := repo.PublicList(context.Background(), filter)
	assert.Error(t, err)
}

func TestVehicleRepositoryPublicCountByType_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	rows := sqlmock.NewRows([]string{"vehicle_type", "count"}).
		AddRow("car", int64(3)).
		AddRow("bike", int64(1))
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	filter := vehicle.PublicListFilter{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}
	counts, err := repo.PublicCountByType(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), counts[vehicle.VehicleTypeCar])
	assert.Equal(t, int64(1), counts[vehicle.VehicleTypeBike])
}

func TestVehicleRepositoryPublicCountByType_Error(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrInvalidData)

	filter := vehicle.PublicListFilter{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}
	_, err := repo.PublicCountByType(context.Background(), filter)
	assert.Error(t, err)
}

func TestRepo_GetVehicleShowroomID_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "vehicle_id", "showroom_id", "created_at", "updated_at", "deleted_at"}).
		AddRow(uint64(1), uint64(10), uint64(5), now, now, nil)
	mock.ExpectQuery(`SELECT \* FROM "vehicle_showroom_relations"`).WillReturnRows(rows)

	showroomID, err := repo.GetVehicleShowroomID(context.Background(), 10)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), showroomID)
}

func TestRepo_GetVehicleShowroomID_NotFound(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "vehicle_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetVehicleShowroomID(context.Background(), 999)
	assert.ErrorIs(t, err, vehicle.ErrVehicleNotFound)
}

func TestRepo_GetVehicleShowroomID_DBError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "vehicle_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetVehicleShowroomID(context.Background(), 1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, vehicle.ErrVehicleNotFound)
}

func TestRepo_GetCurrentStatus_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	rows := sqlmock.NewRows([]string{"status"}).AddRow("ready_for_sale")
	mock.ExpectQuery(`SELECT status FROM vehicle_statuses`).WillReturnRows(rows)

	status, err := repo.GetCurrentStatus(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, vehicle.VehicleStatusTypeReadyForSale, status)
}

func TestRepo_GetCurrentStatus_NotFound(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT status FROM vehicle_statuses`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}))

	_, err := repo.GetCurrentStatus(context.Background(), 999)
	assert.ErrorIs(t, err, vehicle.ErrVehicleNotFound)
}

func TestRepo_GetCurrentStatus_DBError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT status FROM vehicle_statuses`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetCurrentStatus(context.Background(), 1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, vehicle.ErrVehicleNotFound)
}

func TestRepo_UpdateVehicleFields_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "vehicles"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	vehicleCols := []string{
		"id", "vehicle_type", "manufacturer", "model", "variant", "color",
		"year_of_manufacture", "rto_code", "registration_number", "registration_state",
		"usage_km", "fuel_type", "transmission_type", "created_at", "updated_at", "deleted_at",
	}
	mock.ExpectQuery(`SELECT \* FROM "vehicles"`).
		WillReturnRows(sqlmock.NewRows(vehicleCols).AddRow(
			uint64(1), "car", "Honda", "City", "V", "White", 2021, "KA-01", "KA01AB", "Karnataka", 30000, "petrol", "manual", now, now, nil,
		))

	result, err := repo.UpdateVehicleFields(context.Background(), 1, map[string]interface{}{"manufacturer": "Honda"})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Honda", result.Manufacturer)
}

func TestRepo_UpdateVehicleFields_NotFound(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "vehicles"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	_, err := repo.UpdateVehicleFields(context.Background(), 999, map[string]interface{}{"manufacturer": "Honda"})
	assert.ErrorIs(t, err, vehicle.ErrVehicleNotFound)
}

func TestRepo_UpdateVehicleFields_DBError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "vehicles"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.UpdateVehicleFields(context.Background(), 1, map[string]interface{}{"manufacturer": "Honda"})
	assert.Error(t, err)
}

func TestRepo_UpdateVehicleFields_FetchAfterUpdateError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "vehicles"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "vehicles"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.UpdateVehicleFields(context.Background(), 1, map[string]interface{}{"manufacturer": "Honda"})
	assert.Error(t, err)
}

func TestRepo_GetPricingByVehicleID_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	cols := []string{"id", "vehicle_id", "buying_price", "buying_date", "price_tag", "tagged_at", "currency", "remarks", "created_at", "updated_at", "deleted_at"}
	rows := sqlmock.NewRows(cols).AddRow(uint64(1), uint64(10), 200000.0, now, 300000.0, now, "inr", "", now, now, nil)
	mock.ExpectQuery(`SELECT \* FROM "vehicle_pricing"`).WillReturnRows(rows)

	pricing, err := repo.GetPricingByVehicleID(context.Background(), 10)
	assert.NoError(t, err)
	assert.NotNil(t, pricing)
	assert.Equal(t, 200000.0, pricing.BuyingPrice)
}

func TestRepo_GetPricingByVehicleID_NotFound(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "vehicle_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	pricing, err := repo.GetPricingByVehicleID(context.Background(), 999)
	assert.NoError(t, err)
	assert.Nil(t, pricing)
}

func TestRepo_GetPricingByVehicleID_DBError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "vehicle_pricing"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetPricingByVehicleID(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_CreatePricing_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "vehicle_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	mock.ExpectCommit()

	now := time.Now()
	pricing := &vehicle.VehiclePricing{
		VehicleID:   10,
		BuyingPrice: 200000.0,
		BuyingDate:  now,
		Currency:    vehicle.CurrencyINR,
	}
	result, err := repo.CreatePricing(context.Background(), pricing)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200000.0, result.BuyingPrice)
}

func TestRepo_CreatePricing_DBError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "vehicle_pricing"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreatePricing(context.Background(), &vehicle.VehiclePricing{})
	assert.Error(t, err)
}

func TestRepo_UpdatePricingFields_Success(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "vehicle_pricing"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	pricingCols := []string{"id", "vehicle_id", "buying_price", "buying_date", "price_tag", "tagged_at", "currency", "remarks", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(`SELECT \* FROM "vehicle_pricing"`).
		WillReturnRows(sqlmock.NewRows(pricingCols).AddRow(
			uint64(1), uint64(5), 200000.0, now, 350000.0, now, "inr", "", now, now, nil,
		))

	result, err := repo.UpdatePricingFields(context.Background(), 5, map[string]interface{}{"price_tag": 350000.0})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 350000.0, result.PriceTag)
}

func TestRepo_UpdatePricingFields_DBError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "vehicle_pricing"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.UpdatePricingFields(context.Background(), 5, map[string]interface{}{"price_tag": 350000.0})
	assert.Error(t, err)
}

func TestRepo_UpdatePricingFields_FetchAfterUpdateError(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "vehicle_pricing"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "vehicle_pricing"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.UpdatePricingFields(context.Background(), 5, map[string]interface{}{"price_tag": 350000.0})
	assert.Error(t, err)
}

func TestGetByIDWithFullDetails_WithSaleInfo(t *testing.T) {
	gormDB, mock := newVehicleMockDB(t)
	repo := vehicle.NewRepository(gormDB)

	now := time.Now()
	vehicleCols := []string{
		"id", "vehicle_type", "manufacturer", "model", "variant", "color",
		"year_of_manufacture", "rto_code", "registration_number", "registration_state",
		"usage_km", "fuel_type", "transmission_type", "created_at", "updated_at", "deleted_at",
	}
	vehicleRows := sqlmock.NewRows(vehicleCols).
		AddRow(uint64(2), "car", "Honda", "City", "V", "White", 2021, "DL-01", "DL01XY5678", "Delhi", 30000, "petrol", "automatic", now, now, nil)
	mock.ExpectQuery(`SELECT \* FROM "vehicles"`).WillReturnRows(vehicleRows)

	mock.ExpectQuery(`SELECT \* FROM "vehicle_pricing"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "vehicle_statuses"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "vehicle_documents"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "vehicle_expenses"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "vehicle_images"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery(`SELECT \* FROM "vehicle_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	saleCols := []string{
		"sale_price", "sale_date", "payment_mode", "receipt_url", "remarks",
		"customer_first_name", "customer_last_name", "customer_email",
		"customer_phone", "customer_address", "customer_city", "customer_state",
	}
	saleRows := sqlmock.NewRows(saleCols).
		AddRow(500000.0, now, "cash", "https://receipt.url", "sold", "John", "Doe", "john@example.com", "9876543210", "123 Main St", "Mumbai", "Maharashtra")
	mock.ExpectQuery(`customer_vehicle_sales`).WillReturnRows(saleRows)

	details, err := repo.GetByIDWithFullDetails(context.Background(), 2)
	assert.NoError(t, err)
	assert.NotNil(t, details)
	assert.NotNil(t, details.SaleInfo)
	assert.Equal(t, 500000.0, details.SaleInfo.SalePrice)
	assert.Equal(t, "John", details.SaleInfo.CustomerFirstName)
}
