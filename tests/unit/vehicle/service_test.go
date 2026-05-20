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
		name               string
		transmissionType   vehicle.TransmissionType
		shouldPass         bool
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
