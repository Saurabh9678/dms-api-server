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

func (m *mockHandlerService) GetVehicleByID(ctx context.Context, vehicleID uint64) (*vehicle.VehicleFullDetails, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehicleFullDetails), args.Error(1)
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

func setupGetVehicleContext(t *testing.T, idParam string) (*gin.Context, *httptest.ResponseRecorder, *mockHandlerService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockHandlerService)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/api/v1/vehicle/"+idParam, nil)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: idParam}}
	return ctx, w, mockSvc
}

func TestHandler_GetVehicle_InvalidID(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "abc")
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "GetVehicleByID")
}

func TestHandler_GetVehicle_ZeroID(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "0")
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "GetVehicleByID")
}

func TestHandler_GetVehicle_NotFound(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "99")
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(99)).Return(nil, vehicle.ErrVehicleNotFound)
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_GetVehicle_ServiceError(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "1")
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(1)).Return(nil, &mockError{})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_GetVehicle_MissingShowroomRoles(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "1")
	details := &vehicle.VehicleFullDetails{Vehicle: vehicle.Vehicle{ID: 1}, ShowroomID: 5}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(1)).Return(details, nil)
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_GetVehicle_WrongShowroomRolesType(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "1")
	details := &vehicle.VehicleFullDetails{Vehicle: vehicle.Vehicle{ID: 1}, ShowroomID: 5}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(1)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, "not-a-map")
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_GetVehicle_UserNotInShowroom(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "1")
	details := &vehicle.VehicleFullDetails{Vehicle: vehicle.Vehicle{ID: 1}, ShowroomID: 5}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(1)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, map[uint64]string{99: "owner"})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_GetVehicle_OwnerGetsAdminResponse(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "1")
	details := &vehicle.VehicleFullDetails{
		Vehicle:    vehicle.Vehicle{ID: 1, VehicleType: vehicle.VehicleTypeCar, Manufacturer: "Toyota"},
		ShowroomID: 5,
		Statuses:   []vehicle.VehicleStatus{},
		Expenses:   []vehicle.VehicleExpenses{},
		Documents:  []vehicle.VehicleDocument{},
		Images:     []vehicle.VehicleImage{},
	}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(1)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, map[uint64]string{5: "owner"})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "basic")
	assert.Contains(t, data, "status")
	assert.Contains(t, data, "expenses")
	assert.Contains(t, data, "documents")
	assert.Contains(t, data, "images")
}

func TestHandler_GetVehicle_ManagerGetsBasicResponse(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "1")
	details := &vehicle.VehicleFullDetails{
		Vehicle:    vehicle.Vehicle{ID: 1, VehicleType: vehicle.VehicleTypeCar},
		ShowroomID: 5,
		Statuses:   []vehicle.VehicleStatus{},
	}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(1)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, map[uint64]string{5: "manager"})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "basic")
	assert.NotContains(t, data, "expenses")
	assert.NotContains(t, data, "buying_details")
}

func TestHandler_GetVehicle_EmployeeGetsBasicResponse(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "1")
	details := &vehicle.VehicleFullDetails{
		Vehicle:    vehicle.Vehicle{ID: 1, VehicleType: vehicle.VehicleTypeCar},
		ShowroomID: 5,
		Statuses:   []vehicle.VehicleStatus{},
	}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(1)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, map[uint64]string{5: "employee"})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "basic")
	assert.NotContains(t, data, "buying_details")
}

func TestHandler_GetVehicle_OwnerWithPricingAndSale(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "2")
	now := "2024-01-01T00:00:00Z"
	_ = now
	pricing := &vehicle.VehiclePricing{BuyingPrice: 200000, PriceTag: 300000, Currency: vehicle.CurrencyINR}
	saleInfo := &vehicle.VehicleSaleInfo{SalePrice: 280000, CustomerFirstName: "John", CustomerLastName: "Doe"}
	details := &vehicle.VehicleFullDetails{
		Vehicle:    vehicle.Vehicle{ID: 2, VehicleType: vehicle.VehicleTypeCar},
		ShowroomID: 5,
		Statuses:   []vehicle.VehicleStatus{},
		Expenses:   []vehicle.VehicleExpenses{},
		Documents:  []vehicle.VehicleDocument{},
		Images:     []vehicle.VehicleImage{},
		Pricing:    pricing,
		SaleInfo:   saleInfo,
	}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(2)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, map[uint64]string{5: "owner"})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "buying_details")
	assert.Contains(t, data, "pricing")
	assert.Contains(t, data, "selling")
}

func TestHandler_GetVehicle_NonOwnerWithPricingAndSale(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "2")
	pricing := &vehicle.VehiclePricing{PriceTag: 300000, Currency: vehicle.CurrencyINR}
	saleInfo := &vehicle.VehicleSaleInfo{SalePrice: 280000}
	details := &vehicle.VehicleFullDetails{
		Vehicle:    vehicle.Vehicle{ID: 2, VehicleType: vehicle.VehicleTypeCar},
		ShowroomID: 5,
		Statuses:   []vehicle.VehicleStatus{},
		Pricing:    pricing,
		SaleInfo:   saleInfo,
	}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(2)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, map[uint64]string{5: "manager"})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "pricing")
	assert.Contains(t, data, "selling")
	selling := data["selling"].(map[string]interface{})
	assert.Equal(t, 280000.0, selling["sale_price"])
}

func TestHandler_GetVehicle_OwnerWithNonEmptyCollections(t *testing.T) {
	ctx, w, mockSvc := setupGetVehicleContext(t, "3")
	status := vehicle.VehicleStatus{
		ID:     1,
		Status: vehicle.VehicleStatusTypeGarage,
	}
	details := &vehicle.VehicleFullDetails{
		Vehicle:    vehicle.Vehicle{ID: 3, VehicleType: vehicle.VehicleTypeCar},
		ShowroomID: 5,
		Statuses: []vehicle.VehicleStatus{
			status,
			{ID: 2, Status: vehicle.VehicleStatusTypeInspection},
		},
		Expenses: []vehicle.VehicleExpenses{
			{ID: 1, Type: "repair", Amount: 5000, PaidTo: "garage", Description: "fix"},
		},
		Documents: []vehicle.VehicleDocument{
			{ID: 1, DocumentType: vehicle.VehicleDocumentTypeInsurance, DocumentURL: "http://example.com/doc"},
		},
		Images: []vehicle.VehicleImage{
			{ID: 1, Label: vehicle.VehicleImageLabelFront, ImageURL: "http://example.com/img"},
		},
	}
	mockSvc.On("GetVehicleByID", mock.Anything, uint64(3)).Return(details, nil)
	ctx.Set(middleware.ContextKeyShowroomRoles, map[uint64]string{5: "owner"})
	vehicle.NewHandler(mockSvc).GetVehicle(ctx)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "basic")
	assert.Contains(t, data, "status")
	assert.Contains(t, data, "expenses")
	assert.Contains(t, data, "documents")
	assert.Contains(t, data, "images")
	expenses := data["expenses"].([]interface{})
	assert.Len(t, expenses, 1)
	documents := data["documents"].([]interface{})
	assert.Len(t, documents, 1)
	images := data["images"].([]interface{})
	assert.Len(t, images, 1)
}
