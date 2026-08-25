package vehicle_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func (m *mockVehicleRepo) CountByType(ctx context.Context, f vehicle.ListFilter) (map[vehicle.VehicleType]vehicle.CategoryMetrics, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[vehicle.VehicleType]vehicle.CategoryMetrics), args.Error(1)
}

func metricsCounts(m map[vehicle.VehicleType]int64) map[vehicle.VehicleType]vehicle.CategoryMetrics {
	out := make(map[vehicle.VehicleType]vehicle.CategoryMetrics, len(m))
	for k, v := range m {
		out[k] = vehicle.CategoryMetrics{Total: v}
	}
	return out
}

func (m *mockVehicleRepo) GetByIDWithFullDetails(ctx context.Context, vehicleID uint64) (*vehicle.VehicleFullDetails, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehicleFullDetails), args.Error(1)
}

func (m *mockVehicleRepo) PublicList(ctx context.Context, f vehicle.PublicListFilter) ([]vehicle.VehicleWithDetails, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]vehicle.VehicleWithDetails), args.Error(1)
}

func (m *mockVehicleRepo) GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error) {
	args := m.Called(ctx, vehicleID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockVehicleRepo) GetCurrentStatus(ctx context.Context, vehicleID uint64) (vehicle.VehicleStatusType, error) {
	args := m.Called(ctx, vehicleID)
	return args.Get(0).(vehicle.VehicleStatusType), args.Error(1)
}

func (m *mockVehicleRepo) UpdateVehicleFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*vehicle.Vehicle, error) {
	args := m.Called(ctx, vehicleID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.Vehicle), args.Error(1)
}

func (m *mockVehicleRepo) GetPricingByVehicleID(ctx context.Context, vehicleID uint64) (*vehicle.VehiclePricing, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehiclePricing), args.Error(1)
}

func (m *mockVehicleRepo) CreatePricing(ctx context.Context, pricing *vehicle.VehiclePricing) (*vehicle.VehiclePricing, error) {
	args := m.Called(ctx, pricing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehiclePricing), args.Error(1)
}

func (m *mockVehicleRepo) UpdatePricingFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*vehicle.VehiclePricing, error) {
	args := m.Called(ctx, vehicleID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehiclePricing), args.Error(1)
}

func (m *mockVehicleRepo) PublicCountByType(ctx context.Context, f vehicle.PublicListFilter) (map[vehicle.VehicleType]int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[vehicle.VehicleType]int64), args.Error(1)
}

func (m *mockVehicleRepo) CreateExpense(ctx context.Context, expense *vehicle.VehicleExpenses) (*vehicle.VehicleExpenses, error) {
	args := m.Called(ctx, expense)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehicleExpenses), args.Error(1)
}

func (m *mockVehicleRepo) VehicleExistsByID(ctx context.Context, vehicleID uint64) (bool, error) {
	args := m.Called(ctx, vehicleID)
	return args.Bool(0), args.Error(1)
}

func (m *mockVehicleRepo) AssignShowroom(ctx context.Context, vehicleID, showroomID uint64) (*vehicle.VehicleShowroom, error) {
	args := m.Called(ctx, vehicleID, showroomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehicleShowroom), args.Error(1)
}

func (m *mockVehicleRepo) CreateImage(ctx context.Context, img *vehicle.VehicleImage) (*vehicle.VehicleImage, error) {
	args := m.Called(ctx, img)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehicleImage), args.Error(1)
}

func (m *mockVehicleRepo) SoftDeleteImage(ctx context.Context, vehicleID, imageID uint64) error {
	args := m.Called(ctx, vehicleID, imageID)
	return args.Error(0)
}

func (m *mockVehicleRepo) ListImagesByVehicleIDs(ctx context.Context, vehicleIDs []uint64) (map[uint64][]vehicle.VehicleImage, error) {
	if !mockVehicleRepoHasExpectation(m, "ListImagesByVehicleIDs") {
		return map[uint64][]vehicle.VehicleImage{}, nil
	}
	args := m.Called(ctx, vehicleIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uint64][]vehicle.VehicleImage), args.Error(1)
}

func mockVehicleRepoHasExpectation(m *mockVehicleRepo, method string) bool {
	for _, call := range m.ExpectedCalls {
		if call.Method == method {
			return true
		}
	}
	return false
}

type passthroughStorage struct{}

func (passthroughStorage) Upload(_ context.Context, key string, _ []byte, _ string) (string, error) {
	return key, nil
}

func (passthroughStorage) SignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return key, nil
}

func newTestService(repo *mockVehicleRepo, opts ...vehicle.ServiceOption) vehicle.Service {
	return vehicle.NewService(repo, passthroughStorage{}, opts...)
}

func TestCreateVehicle_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

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
			svc := newTestService(mockRepo)

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
			svc := newTestService(mockRepo)

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
			svc := newTestService(mockRepo)

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
	svc := newTestService(mockRepo)

	resp, err := svc.CreateVehicle(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestValidateRequest_ValidRequest(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

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

func TestListVehicles_DefaultAllStatuses(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}

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
		return f.ShowroomID == 1 && len(f.Statuses) == 0 && f.Page == 1 && f.Limit == 20
	})).Return(metricsCounts(counts), nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.ShowroomID == 1 && len(f.Statuses) == 0 && f.Page == 1 && f.Limit == 20
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Cars)
	assert.Equal(t, int64(2), resp.Cars.Total)
	assert.Equal(t, 1, resp.Cars.Page)
	assert.Equal(t, 20, resp.Cars.Limit)
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
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{
		ShowroomID: 1,
		Statuses:   []string{"garage", "inspection"},
		Page:       1,
		Limit:      20,
	}

	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 1}
	vehicles := []vehicle.VehicleWithDetails{{ID: 1, VehicleType: vehicle.VehicleTypeCar}}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.ShowroomID == 1 && len(f.Statuses) == 2
	})).Return(metricsCounts(counts), nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.ShowroomID == 1 && len(f.Statuses) == 2
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestListVehicles_TypeFilter(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{
		ShowroomID:   1,
		VehicleTypes: []string{"car"},
		Page:         1,
		Limit:        20,
	}

	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 1}
	vehicles := []vehicle.VehicleWithDetails{{ID: 1, VehicleType: vehicle.VehicleTypeCar}}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.ShowroomID == 1 && len(f.VehicleTypes) == 1 && f.VehicleTypes[0] == vehicle.VehicleTypeCar && len(f.Statuses) == 0
	})).Return(metricsCounts(counts), nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.ShowroomID == 1 && len(f.VehicleTypes) == 1 && f.VehicleTypes[0] == vehicle.VehicleTypeCar && len(f.Statuses) == 0
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
	svc := newTestService(mockRepo)

	minP := 100000.0
	maxP := 500000.0
	query := &vehicle.ListVehiclesQuery{
		ShowroomID: 1,
		MinPrice:   &minP,
		MaxPrice:   &maxP,
		Page:       1,
		Limit:      20,
	}

	counts := map[vehicle.VehicleType]int64{}
	vehicles := []vehicle.VehicleWithDetails{}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.MinPrice != nil && *f.MinPrice == minP && f.MaxPrice != nil && *f.MaxPrice == maxP
	})).Return(metricsCounts(counts), nil)
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
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(metricsCounts(map[vehicle.VehicleType]int64{}), nil)
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
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}

	st := vehicle.VehicleStatus{Status: vehicle.VehicleStatusTypeReadyForSale}
	pr := vehicle.VehiclePricing{BuyingPrice: 200000, PriceTag: 300000, Currency: vehicle.CurrencyINR}
	vehicles := []vehicle.VehicleWithDetails{
		{ID: 1, VehicleType: vehicle.VehicleTypeCar, CurrentStatus: &st, CurrentPricing: &pr},
	}

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(metricsCounts(map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 1}), nil)
	mockRepo.On("List", mock.Anything, mock.Anything).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Cars.Vehicles, 1)
	assert.NotNil(t, resp.Cars.Vehicles[0].CurrentStatus)
	assert.Equal(t, "ready_for_sale", resp.Cars.Vehicles[0].CurrentStatus.Status)
	assert.NotNil(t, resp.Cars.Vehicles[0].Pricing)
	assert.Equal(t, 300000.0, resp.Cars.Vehicles[0].Pricing.PriceTag)
	assert.NotNil(t, resp.Cars.Vehicles[0].Images)
	assert.Empty(t, resp.Cars.Vehicles[0].Images)
	assert.False(t, resp.Cars.Vehicles[0].IsDeadStock)
	assert.Equal(t, int64(0), resp.Cars.DeadStockCount)
}

func TestListVehicles_DeadStockFlags(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}
	old := time.Now().Add(-100 * 24 * time.Hour)
	recent := time.Now().Add(-10 * 24 * time.Hour)
	vehicles := []vehicle.VehicleWithDetails{
		{ID: 1, VehicleType: vehicle.VehicleTypeCar, BuyingDate: &old, HasActiveSale: false},
		{ID: 2, VehicleType: vehicle.VehicleTypeCar, BuyingDate: &old, HasActiveSale: true},
		{ID: 3, VehicleType: vehicle.VehicleTypeCar, BuyingDate: &recent, HasActiveSale: false},
	}
	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]vehicle.CategoryMetrics{
		vehicle.VehicleTypeCar: {Total: 3, DeadStockCount: 1},
	}, nil)
	mockRepo.On("List", mock.Anything, mock.Anything).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	require.NotNil(t, resp.Cars)
	assert.Equal(t, int64(1), resp.Cars.DeadStockCount)
	require.Len(t, resp.Cars.Vehicles, 3)
	assert.True(t, resp.Cars.Vehicles[0].IsDeadStock)
	assert.False(t, resp.Cars.Vehicles[1].IsDeadStock)
	assert.False(t, resp.Cars.Vehicles[2].IsDeadStock)
}

func TestListVehicles_ZeroShowroomID(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 0, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "CountByType")
}

func TestListVehicles_InvalidPage(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 0, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_InvalidLimit(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

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
			resp, err := svc.ListVehicles(context.Background(), &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: tt.limit})
			assert.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

func TestListVehicles_InvalidStatusEnum(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Statuses: []string{"unknown"}, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_InvalidTypeEnum(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, VehicleTypes: []string{"truck"}, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_MinPriceGreaterThanMaxPrice(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	min := 500000.0
	max := 100000.0
	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, MinPrice: &min, MaxPrice: &max, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_NilQuery(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	resp, err := svc.ListVehicles(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_CountByTypeError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetVehicleByID_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	details := &vehicle.VehicleFullDetails{
		Vehicle: vehicle.Vehicle{ID: 5, VehicleType: vehicle.VehicleTypeCar},
	}
	mockRepo.On("GetByIDWithFullDetails", mock.Anything, uint64(5)).Return(details, nil)

	result, err := svc.GetVehicleByID(context.Background(), 5)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(5), result.Vehicle.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetVehicleByID_NotFound(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	mockRepo.On("GetByIDWithFullDetails", mock.Anything, uint64(99)).Return(nil, vehicle.ErrVehicleNotFound)

	result, err := svc.GetVehicleByID(context.Background(), 99)
	assert.ErrorIs(t, err, vehicle.ErrVehicleNotFound)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetVehicleByID_RepoError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	mockRepo.On("GetByIDWithFullDetails", mock.Anything, uint64(1)).Return(nil, errors.New("db error"))

	result, err := svc.GetVehicleByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestListVehicles_ListError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	mockRepo.On("CountByType", mock.Anything, mock.Anything).Return(metricsCounts(map[vehicle.VehicleType]int64{}), nil)
	mockRepo.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}
	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListVehicles_Pagination(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 2, Limit: 10}

	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 25}
	vehicles := []vehicle.VehicleWithDetails{{ID: 11, VehicleType: vehicle.VehicleTypeCar}}

	mockRepo.On("CountByType", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.ShowroomID == 1 && f.Page == 2 && f.Limit == 10 && len(f.Statuses) == 0
	})).Return(metricsCounts(counts), nil)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f vehicle.ListFilter) bool {
		return f.ShowroomID == 1 && f.Page == 2 && f.Limit == 10 && len(f.Statuses) == 0
	})).Return(vehicles, nil)

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(25), resp.Cars.Total)
	assert.Equal(t, 2, resp.Cars.Page)
	assert.Equal(t, 10, resp.Cars.Limit)
}

func TestPublicListVehicles_NilQuery(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	resp, err := svc.PublicListVehicles(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_ZeroShowroomID(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 0, Page: 1, Limit: 20, SortBy: "price_asc"}
	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_InvalidPage(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 0, Limit: 20, SortBy: "price_asc"}
	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_InvalidLimit(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	tests := []struct{ limit int }{
		{0}, {101}, {-1},
	}
	for _, tt := range tests {
		q := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: tt.limit, SortBy: "price_asc"}
		resp, err := svc.PublicListVehicles(context.Background(), q)
		assert.Error(t, err)
		assert.Nil(t, resp)
	}
}

func TestPublicListVehicles_InvalidSortBy(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "invalid_sort"}
	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_InvalidTypeEnum(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, VehicleTypes: []string{"truck"}, Page: 1, Limit: 20, SortBy: "price_asc"}
	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_MinPriceGreaterThanMaxPrice(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	min, max := 500000.0, 100000.0
	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, MinPrice: &min, MaxPrice: &max, Page: 1, Limit: 20, SortBy: "price_asc"}
	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_CountByTypeError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	mockRepo.On("PublicCountByType", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}
	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_ListError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	mockRepo.On("PublicCountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]int64{}, nil)
	mockRepo.On("PublicList", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}
	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_Success_AllTypes(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}

	pr := vehicle.VehiclePricing{PriceTag: 300000, Currency: vehicle.CurrencyINR}
	vehicles := []vehicle.VehicleWithDetails{
		{ID: 1, VehicleType: vehicle.VehicleTypeCar, CurrentPricing: &pr},
		{ID: 2, VehicleType: vehicle.VehicleTypeBike},
		{ID: 3, VehicleType: vehicle.VehicleTypeScooty},
	}
	counts := map[vehicle.VehicleType]int64{
		vehicle.VehicleTypeCar:    1,
		vehicle.VehicleTypeBike:   1,
		vehicle.VehicleTypeScooty: 1,
	}

	mockRepo.On("PublicCountByType", mock.Anything, mock.MatchedBy(func(f vehicle.PublicListFilter) bool {
		return f.ShowroomID == 1 && f.SortBy == "price_asc"
	})).Return(counts, nil)
	mockRepo.On("PublicList", mock.Anything, mock.MatchedBy(func(f vehicle.PublicListFilter) bool {
		return f.ShowroomID == 1 && f.SortBy == "price_asc"
	})).Return(vehicles, nil)

	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Cars)
	assert.NotNil(t, resp.Bikes)
	assert.NotNil(t, resp.Scooties)
	assert.Len(t, resp.Cars.Vehicles, 1)
	assert.Equal(t, 300000.0, resp.Cars.Vehicles[0].PriceTag)
	assert.Equal(t, "inr", resp.Cars.Vehicles[0].Currency)
	assert.NotNil(t, resp.Cars.Vehicles[0].Images)
	assert.Empty(t, resp.Cars.Vehicles[0].Images)
	mockRepo.AssertExpectations(t)
}

func TestPublicListVehicles_TypeFilter(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.PublicListVehiclesQuery{
		ShowroomID:   1,
		VehicleTypes: []string{"car"},
		Page:         1,
		Limit:        20,
		SortBy:       "price_desc",
	}

	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 2}
	vehicles := []vehicle.VehicleWithDetails{
		{ID: 1, VehicleType: vehicle.VehicleTypeCar},
		{ID: 2, VehicleType: vehicle.VehicleTypeCar},
	}

	mockRepo.On("PublicCountByType", mock.Anything, mock.MatchedBy(func(f vehicle.PublicListFilter) bool {
		return f.SortBy == "price_desc" && len(f.VehicleTypes) == 1
	})).Return(counts, nil)
	mockRepo.On("PublicList", mock.Anything, mock.MatchedBy(func(f vehicle.PublicListFilter) bool {
		return f.SortBy == "price_desc" && len(f.VehicleTypes) == 1
	})).Return(vehicles, nil)

	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Cars)
	assert.Nil(t, resp.Bikes)
	assert.Nil(t, resp.Scooties)
	mockRepo.AssertExpectations(t)
}

func TestPublicListVehicles_PriceRangeFilter(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	min, max := 100000.0, 500000.0
	query := &vehicle.PublicListVehiclesQuery{
		ShowroomID: 1, MinPrice: &min, MaxPrice: &max, Page: 1, Limit: 20, SortBy: "price_asc",
	}

	mockRepo.On("PublicCountByType", mock.Anything, mock.MatchedBy(func(f vehicle.PublicListFilter) bool {
		return f.MinPrice != nil && *f.MinPrice == min && f.MaxPrice != nil && *f.MaxPrice == max
	})).Return(map[vehicle.VehicleType]int64{}, nil)
	mockRepo.On("PublicList", mock.Anything, mock.MatchedBy(func(f vehicle.PublicListFilter) bool {
		return f.MinPrice != nil && f.MaxPrice != nil
	})).Return([]vehicle.VehicleWithDetails{}, nil)

	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestPublicListVehicles_NoPricingOnItem(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}

	vehicles := []vehicle.VehicleWithDetails{
		{ID: 1, VehicleType: vehicle.VehicleTypeCar, CurrentPricing: nil},
	}
	counts := map[vehicle.VehicleType]int64{vehicle.VehicleTypeCar: 1}

	mockRepo.On("PublicCountByType", mock.Anything, mock.Anything).Return(counts, nil)
	mockRepo.On("PublicList", mock.Anything, mock.Anything).Return(vehicles, nil)

	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Cars.Vehicles, 1)
	assert.Equal(t, 0.0, resp.Cars.Vehicles[0].PriceTag)
	assert.Equal(t, "", resp.Cars.Vehicles[0].Currency)
}

// ---- GetVehicleShowroomID ----

func TestGetVehicleShowroomID_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("GetVehicleShowroomID", mock.Anything, uint64(5)).Return(uint64(10), nil)
	id, err := svc.GetVehicleShowroomID(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), id)
	mockRepo.AssertExpectations(t)
}

func TestGetVehicleShowroomID_NotFound(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("GetVehicleShowroomID", mock.Anything, uint64(99)).Return(uint64(0), vehicle.ErrVehicleNotFound)
	id, err := svc.GetVehicleShowroomID(context.Background(), 99)
	assert.ErrorIs(t, err, vehicle.ErrVehicleNotFound)
	assert.Equal(t, uint64(0), id)
	mockRepo.AssertExpectations(t)
}

func TestGetVehicleShowroomID_RepoError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("GetVehicleShowroomID", mock.Anything, uint64(1)).Return(uint64(0), errors.New("db error"))
	_, err := svc.GetVehicleShowroomID(context.Background(), 1)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// ---- UpdateVehicle ----

func TestUpdateVehicle_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	manufacturer := "Honda"
	color := "Red"
	req := &vehicle.UpdateVehicleRequest{Manufacturer: &manufacturer, Color: &color}

	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	updatedVehicle := &vehicle.Vehicle{
		ID: 1, VehicleType: vehicle.VehicleTypeCar, Manufacturer: "Honda", Color: "Red",
		RegistrationNumber: "KA01AB1234",
	}
	mockRepo.On("UpdateVehicleFields", mock.Anything, uint64(1), mock.Anything).Return(updatedVehicle, nil)

	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, "Honda", resp.Manufacturer)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehicle_NilRequest(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	resp, err := svc.UpdateVehicle(context.Background(), 1, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "GetCurrentStatus")
}

func TestUpdateVehicle_GetCurrentStatusError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	manufacturer := "Honda"
	req := &vehicle.UpdateVehicleRequest{Manufacturer: &manufacturer}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusType(""), errors.New("db error"))
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehicle_VehicleSold(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	manufacturer := "Honda"
	req := &vehicle.UpdateVehicleRequest{Manufacturer: &manufacturer}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeSold, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.ErrorIs(t, err, vehicle.ErrVehicleSold)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehicle_InvalidVehicleType(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	vt := vehicle.VehicleType("truck")
	req := &vehicle.UpdateVehicleRequest{VehicleType: &vt}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_EmptyManufacturer(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	empty := "  "
	req := &vehicle.UpdateVehicleRequest{Manufacturer: &empty}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_EmptyModel(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	empty := ""
	req := &vehicle.UpdateVehicleRequest{Model: &empty}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_EmptyVariant(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	empty := ""
	req := &vehicle.UpdateVehicleRequest{Variant: &empty}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_EmptyColor(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	empty := "  "
	req := &vehicle.UpdateVehicleRequest{Color: &empty}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_YearTooLow(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	year := 1800
	req := &vehicle.UpdateVehicleRequest{YearOfManufacture: &year}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_YearTooHigh(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	year := 2099
	req := &vehicle.UpdateVehicleRequest{YearOfManufacture: &year}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_EmptyRTOCode(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	empty := ""
	req := &vehicle.UpdateVehicleRequest{RTOCode: &empty}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_EmptyRegistrationState(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	empty := ""
	req := &vehicle.UpdateVehicleRequest{RegistrationState: &empty}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_NegativeUsageKM(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	km := -1
	req := &vehicle.UpdateVehicleRequest{UsageKM: &km}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_InvalidFuelType(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	ft := vehicle.FuelType("cng")
	req := &vehicle.UpdateVehicleRequest{FuelType: &ft}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_InvalidTransmissionType(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	tt := vehicle.TransmissionType("cvt")
	req := &vehicle.UpdateVehicleRequest{TransmissionType: &tt}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehicle_AllFieldsNil(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	req := &vehicle.UpdateVehicleRequest{}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "UpdateVehicleFields")
}

func TestUpdateVehicle_UpdateFieldsError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	manufacturer := "Honda"
	req := &vehicle.UpdateVehicleRequest{Manufacturer: &manufacturer}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("UpdateVehicleFields", mock.Anything, uint64(1), mock.Anything).Return(nil, errors.New("db error"))
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehicle_ValidUsageKMZero(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	km := 0
	req := &vehicle.UpdateVehicleRequest{UsageKM: &km}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeReadyForSale, nil)
	updatedVehicle := &vehicle.Vehicle{ID: 1}
	mockRepo.On("UpdateVehicleFields", mock.Anything, uint64(1), mock.Anything).Return(updatedVehicle, nil)
	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestUpdateVehicle_AllValidFields(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	vt := vehicle.VehicleTypeBike
	mfr := "Yamaha"
	model := "R15"
	variant := "V4"
	color := "Blue"
	year := 2022
	rto := "KA-02"
	state := "Karnataka"
	km := 5000
	ft := vehicle.FuelTypePetrol
	tt := vehicle.TransmissionTypeManual

	req := &vehicle.UpdateVehicleRequest{
		VehicleType: &vt, Manufacturer: &mfr, Model: &model, Variant: &variant,
		Color: &color, YearOfManufacture: &year, RTOCode: &rto, RegistrationState: &state,
		UsageKM: &km, FuelType: &ft, TransmissionType: &tt,
	}

	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeInspection, nil)
	updatedVehicle := &vehicle.Vehicle{ID: 1, VehicleType: vt, Manufacturer: mfr}
	mockRepo.On("UpdateVehicleFields", mock.Anything, uint64(1), mock.Anything).Return(updatedVehicle, nil)

	resp, err := svc.UpdateVehicle(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

// ---- UpdateVehiclePricing ----

func TestUpdateVehiclePricing_NilRequest(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "GetCurrentStatus")
}

func TestUpdateVehiclePricing_GetCurrentStatusError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "2024-01-01"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusType(""), errors.New("db error"))
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_VehicleSold(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "2024-01-01"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeSold, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.ErrorIs(t, err, vehicle.ErrVehicleSold)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_GetPricingError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "2024-01-01"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, errors.New("db error"))
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehiclePricing_CreateNew_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	bp := 100000.0
	bd := "2024-01-01"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd}

	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	createdPricing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1, BuyingPrice: 100000, Currency: vehicle.CurrencyINR}
	mockRepo.On("CreatePricing", mock.Anything, mock.Anything).Return(createdPricing, nil)

	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.VehicleID)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehiclePricing_CreateNew_AllFields(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	bp := 100000.0
	bd := "2024-01-01"
	pt := 150000.0
	taggedAt := "2024-01-02T10:00:00Z"
	cur := "usd"
	remarks := "test"
	req := &vehicle.UpdateVehiclePricingRequest{
		BuyingPrice: &bp, BuyingDate: &bd, PriceTag: &pt, TaggedAt: &taggedAt, Currency: &cur, Remarks: &remarks,
	}

	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeReadyForSale, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	createdPricing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1, BuyingPrice: bp, PriceTag: pt, Currency: vehicle.CurrencyUSD, Remarks: remarks}
	mockRepo.On("CreatePricing", mock.Anything, mock.Anything).Return(createdPricing, nil)

	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehiclePricing_CreateNew_NilBuyingPrice(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bd := "2024-01-01"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingDate: &bd}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_CreateNew_ZeroBuyingPrice(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 0.0
	bd := "2024-01-01"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_CreateNew_NilBuyingDate(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_CreateNew_InvalidBuyingDate(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "not-a-date"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_CreateNew_NegativePriceTag(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "2024-01-01"
	pt := -1.0
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd, PriceTag: &pt}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_CreateNew_InvalidTaggedAt(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "2024-01-01"
	ta := "not-a-timestamp"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd, TaggedAt: &ta}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_CreateNew_InvalidCurrency(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "2024-01-01"
	cur := "eur"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd, Currency: &cur}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_CreateNew_CreatePricingError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := 100000.0
	bd := "2024-01-01"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp, BuyingDate: &bd}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(nil, nil)
	mockRepo.On("CreatePricing", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_UpdateExisting_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	pt := 200000.0
	req := &vehicle.UpdateVehiclePricingRequest{PriceTag: &pt}

	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1, BuyingPrice: 100000, Currency: vehicle.CurrencyINR}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeReadyForSale, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	updated := &vehicle.VehiclePricing{ID: 1, VehicleID: 1, BuyingPrice: 100000, PriceTag: 200000, Currency: vehicle.CurrencyINR}
	mockRepo.On("UpdatePricingFields", mock.Anything, uint64(1), mock.Anything).Return(updated, nil)

	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200000.0, resp.PriceTag)
	mockRepo.AssertExpectations(t)
}

func TestUpdateVehiclePricing_UpdateExisting_AllFieldsNil(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	req := &vehicle.UpdateVehiclePricingRequest{}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "UpdatePricingFields")
}

func TestUpdateVehiclePricing_UpdateExisting_InvalidBuyingPrice(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bp := -1.0
	req := &vehicle.UpdateVehiclePricingRequest{BuyingPrice: &bp}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_UpdateExisting_InvalidBuyingDate(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	bd := "bad-date"
	req := &vehicle.UpdateVehiclePricingRequest{BuyingDate: &bd}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_UpdateExisting_NegativePriceTag(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	pt := -1.0
	req := &vehicle.UpdateVehiclePricingRequest{PriceTag: &pt}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_UpdateExisting_InvalidTaggedAt(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	ta := "not-rfc3339"
	req := &vehicle.UpdateVehiclePricingRequest{TaggedAt: &ta}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_UpdateExisting_InvalidCurrency(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	cur := "gbp"
	req := &vehicle.UpdateVehiclePricingRequest{Currency: &cur}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_UpdateExisting_UpdateFieldsError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	pt := 200000.0
	req := &vehicle.UpdateVehiclePricingRequest{PriceTag: &pt}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	mockRepo.On("UpdatePricingFields", mock.Anything, uint64(1), mock.Anything).Return(nil, errors.New("db error"))
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUpdateVehiclePricing_UpdateExisting_Remarks(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	remarks := "updated"
	req := &vehicle.UpdateVehiclePricingRequest{Remarks: &remarks}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	updated := &vehicle.VehiclePricing{ID: 1, VehicleID: 1, Remarks: "updated", Currency: vehicle.CurrencyINR}
	mockRepo.On("UpdatePricingFields", mock.Anything, uint64(1), mock.Anything).Return(updated, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "updated", resp.Remarks)
}

func TestUpdateVehiclePricing_UpdateExisting_AllValidFields(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	bp := 150000.0
	bd := "2023-06-15"
	pt := 250000.0
	ta := "2023-06-15T10:00:00Z"
	cur := "usd"
	remarks := "all fields"
	req := &vehicle.UpdateVehiclePricingRequest{
		BuyingPrice: &bp,
		BuyingDate:  &bd,
		PriceTag:    &pt,
		TaggedAt:    &ta,
		Currency:    &cur,
		Remarks:     &remarks,
	}
	existing := &vehicle.VehiclePricing{ID: 1, VehicleID: 1}
	mockRepo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	mockRepo.On("GetPricingByVehicleID", mock.Anything, uint64(1)).Return(existing, nil)
	updated := &vehicle.VehiclePricing{
		ID: 1, VehicleID: 1,
		BuyingPrice: 150000, PriceTag: 250000,
		Currency: vehicle.CurrencyUSD, Remarks: "all fields",
	}
	mockRepo.On("UpdatePricingFields", mock.Anything, uint64(1), mock.Anything).Return(updated, nil)
	resp, err := svc.UpdateVehiclePricing(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 250000.0, resp.PriceTag)
	mockRepo.AssertExpectations(t)
}

func TestAddExpense_NilRequest(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	resp, err := svc.AddExpense(context.Background(), 1, nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "CreateExpense")
}

func TestAddExpense_InvalidType(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	req := &vehicle.AddExpenseRequest{Type: "invalid_type", Amount: 1000}
	resp, err := svc.AddExpense(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "CreateExpense")
}

func TestAddExpense_ZeroAmount(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	req := &vehicle.AddExpenseRequest{Type: "repair", Amount: 0}
	resp, err := svc.AddExpense(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "CreateExpense")
}

func TestAddExpense_NegativeAmount(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	req := &vehicle.AddExpenseRequest{Type: "repair", Amount: -500}
	resp, err := svc.AddExpense(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "CreateExpense")
}

func TestAddExpense_InvalidDateFormat(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	badDate := "not-a-date"
	req := &vehicle.AddExpenseRequest{Type: "repair", Amount: 1000, Date: &badDate}
	resp, err := svc.AddExpense(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertNotCalled(t, "CreateExpense")
}

func TestAddExpense_WithValidDate(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	date := "2024-03-15T10:00:00Z"
	req := &vehicle.AddExpenseRequest{Type: "service", Amount: 2500, PaidTo: "Mechanic", Description: "Routine service", Date: &date}
	created := &vehicle.VehicleExpenses{ID: 1, VehicleID: 1, Type: vehicle.VehicleExpensesTypeService, Amount: 2500}
	mockRepo.On("CreateExpense", mock.Anything, mock.Anything).Return(created, nil)

	resp, err := svc.AddExpense(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.ID)
	assert.Equal(t, "service", resp.Type)
	mockRepo.AssertExpectations(t)
}

func TestAddExpense_WithoutDate_UsesNow(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	req := &vehicle.AddExpenseRequest{Type: "repair", Amount: 3000}
	created := &vehicle.VehicleExpenses{ID: 2, VehicleID: 1, Type: vehicle.VehicleExpensesTypeRepair, Amount: 3000}
	mockRepo.On("CreateExpense", mock.Anything, mock.Anything).Return(created, nil)

	resp, err := svc.AddExpense(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(2), resp.ID)
	mockRepo.AssertExpectations(t)
}

func TestAddExpense_RepoError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)

	req := &vehicle.AddExpenseRequest{Type: "insurance", Amount: 5000}
	mockRepo.On("CreateExpense", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	resp, err := svc.AddExpense(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "db error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestAddExpense_AllValidTypes(t *testing.T) {
	types := []string{"repair", "service", "insurance", "tax", "inspection", "cleaning", "documentation", "other"}
	for _, expType := range types {
		t.Run(expType, func(t *testing.T) {
			mockRepo := new(mockVehicleRepo)
			svc := newTestService(mockRepo)

			req := &vehicle.AddExpenseRequest{Type: expType, Amount: 1000}
			created := &vehicle.VehicleExpenses{ID: 1, VehicleID: 1, Type: vehicle.VehicleExpensesType(expType), Amount: 1000}
			mockRepo.On("CreateExpense", mock.Anything, mock.Anything).Return(created, nil)

			resp, err := svc.AddExpense(context.Background(), 1, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, expType, resp.Type)
		})
	}
}

// ---- AssignVehicleToShowroom ----

func TestAssignVehicleToShowroom_VehicleNotFound(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("VehicleExistsByID", mock.Anything, uint64(1)).Return(false, nil)
	resp, err := svc.AssignVehicleToShowroom(context.Background(), 1, 10)
	assert.ErrorIs(t, err, vehicle.ErrVehicleNotFound)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAssignVehicleToShowroom_VehicleExistsError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("VehicleExistsByID", mock.Anything, uint64(1)).Return(false, errors.New("db error"))
	resp, err := svc.AssignVehicleToShowroom(context.Background(), 1, 10)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAssignVehicleToShowroom_AlreadyAssigned(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("VehicleExistsByID", mock.Anything, uint64(1)).Return(true, nil)
	mockRepo.On("GetVehicleShowroomID", mock.Anything, uint64(1)).Return(uint64(5), nil)
	resp, err := svc.AssignVehicleToShowroom(context.Background(), 1, 10)
	assert.ErrorIs(t, err, vehicle.ErrVehicleAlreadyInShowroom)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAssignVehicleToShowroom_GetShowroomIDUnexpectedError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("VehicleExistsByID", mock.Anything, uint64(1)).Return(true, nil)
	mockRepo.On("GetVehicleShowroomID", mock.Anything, uint64(1)).Return(uint64(0), errors.New("db error"))
	resp, err := svc.AssignVehicleToShowroom(context.Background(), 1, 10)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAssignVehicleToShowroom_AssignShowroomError(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("VehicleExistsByID", mock.Anything, uint64(1)).Return(true, nil)
	mockRepo.On("GetVehicleShowroomID", mock.Anything, uint64(1)).Return(uint64(0), vehicle.ErrVehicleNotFound)
	mockRepo.On("AssignShowroom", mock.Anything, uint64(1), uint64(10)).Return(nil, errors.New("db error"))
	resp, err := svc.AssignVehicleToShowroom(context.Background(), 1, 10)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAssignVehicleToShowroom_Success(t *testing.T) {
	mockRepo := new(mockVehicleRepo)
	svc := newTestService(mockRepo)
	mockRepo.On("VehicleExistsByID", mock.Anything, uint64(1)).Return(true, nil)
	mockRepo.On("GetVehicleShowroomID", mock.Anything, uint64(1)).Return(uint64(0), vehicle.ErrVehicleNotFound)
	rel := &vehicle.VehicleShowroom{VehicleID: 1, ShowroomID: 10}
	mockRepo.On("AssignShowroom", mock.Anything, uint64(1), uint64(10)).Return(rel, nil)
	resp, err := svc.AssignVehicleToShowroom(context.Background(), 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.VehicleID)
	assert.Equal(t, uint64(10), resp.ShowroomID)
	mockRepo.AssertExpectations(t)
}
