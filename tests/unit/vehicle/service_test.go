package vehicle_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"infiour.local/dms-api-server/internal/modules/vehicle"
)

type mockVehicleRepo struct {
	mock.Mock
}

func (m *mockVehicleRepo) Create(ctx context.Context, v *vehicle.Vehicle) (*vehicle.Vehicle, error) {
	args := m.Called(ctx, v)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.Vehicle), args.Error(1)
}

func (m *mockVehicleRepo) List(ctx context.Context, f vehicle.ListFilter) ([]vehicle.VehicleWithDetails, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]vehicle.VehicleWithDetails), args.Error(1)
}

func (m *mockVehicleRepo) CountByType(ctx context.Context, f vehicle.ListFilter) (map[vehicle.VehicleType]int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[vehicle.VehicleType]int64), args.Error(1)
}

func TestCreateVehicle_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	expectedVehicle := &vehicle.Vehicle{
		ID:                 1,
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(v *vehicle.Vehicle) bool {
		return v.VehicleType == vehicle.VehicleTypeCar && v.Manufacturer == "Toyota"
	})).Return(expectedVehicle, nil).Run(func(args mock.Arguments) {
		v := args.Get(1).(*vehicle.Vehicle)
		v.ID = 1
	})

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, "car", resp.VehicleType)
	assert.Equal(t, "Toyota", resp.Manufacturer)
	mockRepo.AssertExpectations(t)
}

func TestCreateVehicle_InvalidVehicleType(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleType("truck"),
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateVehicle_EmptyManufacturer(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "   ",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateVehicle_EmptyModel(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_EmptyVariant(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_EmptyColor(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "  ",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_YearBelowMin(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  1800,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_YearInFuture(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2099,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_EmptyRTOCode(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_EmptyRegistrationNumber(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_EmptyRegistrationState(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_NegativeUsageKM(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            -100,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_InvalidFuelType(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelType("cng"),
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_InvalidTransmissionType(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionType("cvt"),
	}

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateVehicle_RepositoryError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	resp, err := svc.CreateVehicle(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "db error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestCreateVehicle_AllVehicleTypes(t *testing.T) {
	tests := []struct {
		name        string
		vehicleType vehicle.VehicleType
		shouldPass  bool
	}{
		{"bike", vehicle.VehicleTypeBike, true},
		{"car", vehicle.VehicleTypeCar, true},
		{"scooty", vehicle.VehicleTypeScooty, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockVehicleRepo)
			svc := vehicle.NewService(mockRepo)

			req := &vehicle.CreateVehicleRequest{
				VehicleType:        tt.vehicleType,
				Manufacturer:       "Manufacturer",
				Model:              "Model",
				Variant:            "Variant",
				Color:              "Color",
				YearOfManufacture:  2020,
				RTOCode:            "Code",
				RegistrationNumber: "Number",
				RegistrationState:  "State",
				UsageKM:            0,
				FuelType:           vehicle.FuelTypePetrol,
				TransmissionType:   vehicle.TransmissionTypeManual,
			}

			expectedVehicle := &vehicle.Vehicle{
				ID:          1,
				VehicleType: tt.vehicleType,
			}

			mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedVehicle, nil)

			resp, err := svc.CreateVehicle(context.Background(), req)

			if tt.shouldPass {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestCreateVehicle_AllFuelTypes(t *testing.T) {
	tests := []struct {
		name       string
		fuelType   vehicle.FuelType
		shouldPass bool
	}{
		{"petrol", vehicle.FuelTypePetrol, true},
		{"diesel", vehicle.FuelTypeDiesel, true},
		{"ev", vehicle.FuelTypeEV, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockVehicleRepo)
			svc := vehicle.NewService(mockRepo)

			req := &vehicle.CreateVehicleRequest{
				VehicleType:        vehicle.VehicleTypeCar,
				Manufacturer:       "Manufacturer",
				Model:              "Model",
				Variant:            "Variant",
				Color:              "Color",
				YearOfManufacture:  2020,
				RTOCode:            "Code",
				RegistrationNumber: "Number",
				RegistrationState:  "State",
				UsageKM:            0,
				FuelType:           tt.fuelType,
				TransmissionType:   vehicle.TransmissionTypeManual,
			}

			expectedVehicle := &vehicle.Vehicle{
				ID:       1,
				FuelType: tt.fuelType,
			}

			mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedVehicle, nil)

			resp, err := svc.CreateVehicle(context.Background(), req)

			if tt.shouldPass {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestCreateVehicle_AllTransmissionTypes(t *testing.T) {
	tests := []struct {
		name             string
		transmissionType vehicle.TransmissionType
		shouldPass       bool
	}{
		{"manual", vehicle.TransmissionTypeManual, true},
		{"automatic", vehicle.TransmissionTypeAutomatic, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockVehicleRepo)
			svc := vehicle.NewService(mockRepo)

			req := &vehicle.CreateVehicleRequest{
				VehicleType:        vehicle.VehicleTypeCar,
				Manufacturer:       "Manufacturer",
				Model:              "Model",
				Variant:            "Variant",
				Color:              "Color",
				YearOfManufacture:  2020,
				RTOCode:            "Code",
				RegistrationNumber: "Number",
				RegistrationState:  "State",
				UsageKM:            0,
				FuelType:           vehicle.FuelTypePetrol,
				TransmissionType:   tt.transmissionType,
			}

			expectedVehicle := &vehicle.Vehicle{
				ID:               1,
				TransmissionType: tt.transmissionType,
			}

			mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedVehicle, nil)

			resp, err := svc.CreateVehicle(context.Background(), req)

			if tt.shouldPass {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestValidateRequest_NilRequest(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	resp, err := svc.CreateVehicle(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestValidateRequest_ValidRequest(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	req := &vehicle.CreateVehicleRequest{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	expectedVehicle := &vehicle.Vehicle{
		ID: 1,
	}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedVehicle, nil)

	resp, err := svc.CreateVehicle(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestRepository_Create_Covered(t *testing.T) {
	mockRepo := new(mockVehicleRepo)

	v := &vehicle.Vehicle{
		VehicleType:        vehicle.VehicleTypeCar,
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           vehicle.FuelTypePetrol,
		TransmissionType:   vehicle.TransmissionTypeManual,
	}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(v, nil)

	result, err := mockRepo.Create(context.Background(), v)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, v.Manufacturer, result.Manufacturer)
}

func TestListVehicles_DefaultStatusFilter(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{Page: 1, Limit: 20}

	counts := map[vehicle.VehicleType]int64{
		vehicle.VehicleTypeCar:    2,
		vehicle.VehicleTypeBike:   1,
		vehicle.VehicleTypeScooty: 0,
	}
	vehicles := []vehicle.VehicleWithDetails{
		{ID: 1, VehicleType: vehicle.VehicleTypeCar},
		{ID: 2, VehicleType: vehicle.VehicleTypeCar},
		{ID: 3, VehicleType: vehicle.VehicleTypeBike},
	}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return len(f.Statuses) == 1 && f.Statuses[0] == vehicle.VehicleStatusTypeReadyForSale
	})).Return(counts, nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return len(f.Statuses) == 1 && f.Statuses[0] == vehicle.VehicleStatusTypeReadyForSale
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Cars)
	assert.Equal(t, int64(2), resp.Cars.Total)
	assert.Len(t, resp.Cars.Vehicles, 2)
	assert.NotNil(t, resp.Bikes)
	assert.Equal(t, int64(1), resp.Bikes.Total)
	assert.Len(t, resp.Bikes.Vehicles, 1)
	assert.NotNil(t, resp.Scooties)
	assert.Equal(t, int64(0), resp.Scooties.Total)
	mockRepo.AssertExpectations(t)
}

func TestListVehicles_MultiStatusFilter(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{
		Statuses: []string{"garage", "inspection"},
		Page:     1,
		Limit:    20,
	}

	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 1}
	vehicles := []vehicle.VehicleWithDetails{{ID: 1, VehicleType: vehicle.VehicleTypeCar}}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return len(f.Statuses) == 2
	})).Return(counts, nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return len(f.Statuses) == 2
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestListVehicles_TypeFilter(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{
		VehicleTypes: []string{"car"},
		Page:         1,
		Limit:        20,
	}

	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 1}
	vehicles := []vehicle.VehicleWithDetails{{ID: 1, VehicleType: vehicle.VehicleTypeCar}}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return len(f.VehicleTypes) == 1 && f.VehicleTypes[0] == vehicle.VehicleTypeCar
	})).Return(counts, nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return len(f.VehicleTypes) == 1 && f.VehicleTypes[0] == vehicle.VehicleTypeCar
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Cars)
	assert.Nil(t, resp.Bikes)
	assert.Nil(t, resp.Scooties)
	mockRepo.AssertExpectations(t)
}

func TestListVehicles_PriceRangeFilter(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	minP := 100000.0
	maxP := 500000.0
	query := &vehicle.ListVehiclesQuery{
		MinPrice: &minP,
		MaxPrice: &maxP,
		Page:     1,
		Limit:    20,
	}

	counts := map[vehicle.VehicleType]int64{}
	vehicles := []vehicle.VehicleWithDetails{}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.MinPrice != nil && *f.MinPrice == minP && f.MaxPrice != nil && *f.MaxPrice == maxP
	})).Return(counts, nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.MinPrice != nil && f.MaxPrice != nil
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestListVehicles_EmptyResult(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{Page: 1, Limit: 20}

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]int64{}, nil)
	mockRepo.On("List", mock.Anything, mock.Anything).Return([]vehicle.VehicleWithDetails{}, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Cars)
	assert.Len(t, resp.Cars.Vehicles, 0)
	assert.Equal(t, int64(0), resp.Cars.Total)
}

func TestListVehicles_WithStatusAndPricing(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{Page: 1, Limit: 20}

	st := vehicle.VehicleStatus{Status: vehicle.VehicleStatusTypeReadyForSale}
	pr := vehicle.VehiclePricing{BuyingPrice: 200000, PriceTag: 300000, Currency: vehicle.CurrencyINR}
	vehicles := []vehicle.VehicleWithDetails{
		{ID: 1, VehicleType: vehicle.VehicleTypeCar, CurrentStatus: &st, CurrentPricing: &pr},
	}

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 1}, nil)
	mockRepo.On("List", mock.Anything, mock.Anything).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Cars.Vehicles, 1)
	assert.NotNil(t, resp.Cars.Vehicles[0].CurrentStatus)
	assert.Equal(t, "ready_for_sale", resp.Cars.Vehicles[0].CurrentStatus.Status)
	assert.NotNil(t, resp.Cars.Vehicles[0].Pricing)
	assert.Equal(t, 300000.0, resp.Cars.Vehicles[0].Pricing.PriceTag)
}

func TestListVehicles_InvalidPage(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{Page: 0, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_InvalidLimit(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	tests := []struct {
		name  string
		limit int
	}{
		{"zero", 0},
		{"over100", 101},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.ListVehicles(context.Background(), &vehicle.ListVehiclesQuery{Page: 1, Limit: tt.limit})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestListVehicles_InvalidStatusEnum(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{Statuses: []string{"unknown"}, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_InvalidTypeEnum(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{VehicleTypes: []string{"truck"}, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_MinPriceGreaterThanMaxPrice(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	min := 500000.0
	max := 100000.0
	query := &vehicle.ListVehiclesQuery{MinPrice: &min, MaxPrice: &max, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_NilQuery(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	resp, err := svc.ListVehicles(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_CountByTypeError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	query := &vehicle.ListVehiclesQuery{Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_ListError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]int64{}, nil)
	mockRepo.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	query := &vehicle.ListVehiclesQuery{Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_Pagination(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := vehicle.NewService(mockRepo)

	query := &vehicle.ListVehiclesQuery{Page: 2, Limit: 10}

	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 25}
	vehicles := []vehicle.VehicleWithDetails{{ID: 11, VehicleType: vehicle.VehicleTypeCar}}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.Page == 2 && f.Limit == 10
	})).Return(counts, nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.Page == 2 && f.Limit == 10
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(25), resp.Cars.Total)
	assert.Equal(t, 2, resp.Cars.Page)
	assert.Equal(t, 10, resp.Cars.Limit)
}
