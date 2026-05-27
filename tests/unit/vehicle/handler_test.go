package vehicle_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"infiour.local/dms-api-server/internal/modules/vehicle"
	"infiour.local/dms-api-server/pkg/middleware"
)

type mockHandlerService struct {
	mock.Mock
}

func (m *mockHandlerService) CreateVehicle(ctx context.Context, req *vehicle.CreateVehicleRequest) (*vehicle.CreateVehicleResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.CreateVehicleResponse), args.Error(1)
}

func (m *mockHandlerService) ListVehicles(ctx context.Context, query *vehicle.ListVehiclesQuery) (*vehicle.ListVehiclesResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.ListVehiclesResponse), args.Error(1)
}

func TestHandler_CreateVehicle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	handler := vehicle.NewHandler(mockSvc)

	reqBody := vehicle.CreateVehicleRequest{
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

	respData := &vehicle.CreateVehicleResponse{
		ID:                 1,
		VehicleType:        "car",
		Manufacturer:       "Toyota",
		Model:              "Camry",
		Variant:            "LE",
		Color:              "Black",
		YearOfManufacture:  2020,
		RTOCode:            "KA-01",
		RegistrationNumber: "KA01AB1234",
		RegistrationState:  "Karnataka",
		UsageKM:            50000,
		FuelType:           "petrol",
		TransmissionType:   "manual",
		CreatedAt:          "2024-01-01T00:00:00Z",
		UpdatedAt:          "2024-01-01T00:00:00Z",
	}

	mockSvc.On("CreateVehicle", mock.Anything, mock.Anything).Return(respData, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/vehicle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set(middleware.ContextKeyUserID, uint64(1))

	handler.CreateVehicle(ctx)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_CreateVehicle_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	handler := vehicle.NewHandler(mockSvc)

	req := httptest.NewRequest("POST", "/api/v1/vehicle", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set(middleware.ContextKeyUserID, uint64(1))

	handler.CreateVehicle(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "CreateVehicle")
}

func TestHandler_CreateVehicle_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	handler := vehicle.NewHandler(mockSvc)

	reqBody := vehicle.CreateVehicleRequest{
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

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/vehicle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.CreateVehicle(ctx)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertNotCalled(t, "CreateVehicle")
}

func TestHandler_CreateVehicle_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	handler := vehicle.NewHandler(mockSvc)

	reqBody := vehicle.CreateVehicleRequest{
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

	mockSvc.On("CreateVehicle", mock.Anything, mock.Anything).Return(nil, &mockError{})

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/vehicle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set(middleware.ContextKeyUserID, uint64(1))

	handler.CreateVehicle(ctx)

	mockSvc.AssertExpectations(t)
}

type mockError struct{}

func (e *mockError) Error() string {
	return "test error"
}

func TestHandler_ListVehicles_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	handler := vehicle.NewHandler(mockSvc)

	respData := &vehicle.ListVehiclesResponse{
		Cars: &vehicle.CategoryListing{
			Total:    1,
			Page:     1,
			Limit:    20,
			Vehicles: []vehicle.VehicleListItem{{ID: 1, VehicleType: "car"}},
		},
	}

	mockSvc.On("ListVehicles", mock.Anything, mock.Anything).Return(respData, nil)

	req := httptest.NewRequest("GET", "/api/v1/vehicle/listing", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.ListVehicles(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_ListVehicles_InvalidQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	handler := vehicle.NewHandler(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/vehicle/listing?page=abc", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.ListVehicles(ctx)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "ListVehicles")
}

func TestHandler_ListVehicles_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	handler := vehicle.NewHandler(mockSvc)

	mockSvc.On("ListVehicles", mock.Anything, mock.Anything).Return(nil, &mockError{})

	req := httptest.NewRequest("GET", "/api/v1/vehicle/listing", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	handler.ListVehicles(ctx)

	mockSvc.AssertExpectations(t)
}
