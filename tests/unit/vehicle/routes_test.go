package vehicle_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"infiour.local/dms-api-server/internal/modules/vehicle"
)

type mockRoutesService struct {
	mock.Mock
}

func (m *mockRoutesService) CreateVehicle(ctx context.Context, req *vehicle.CreateVehicleRequest, addedBy uint64) (*vehicle.CreateVehicleResponse, error) {
	args := m.Called(ctx, req, addedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.CreateVehicleResponse), args.Error(1)
}

func (m *mockRoutesService) ListVehicles(ctx context.Context, query *vehicle.ListVehiclesQuery) (*vehicle.ListVehiclesResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.ListVehiclesResponse), args.Error(1)
}

func (m *mockRoutesService) GetVehicleByID(ctx context.Context, vehicleID uint64) (*vehicle.VehicleFullDetails, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.VehicleFullDetails), args.Error(1)
}

func (m *mockRoutesService) PublicListVehicles(ctx context.Context, query *vehicle.PublicListVehiclesQuery) (*vehicle.PublicListVehiclesResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.PublicListVehiclesResponse), args.Error(1)
}

func (m *mockRoutesService) GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error) {
	args := m.Called(ctx, vehicleID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockRoutesService) UpdateVehicle(ctx context.Context, vehicleID uint64, req *vehicle.UpdateVehicleRequest) (*vehicle.UpdateVehicleResponse, error) {
	args := m.Called(ctx, vehicleID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.UpdateVehicleResponse), args.Error(1)
}

func (m *mockRoutesService) UpdateVehiclePricing(ctx context.Context, vehicleID uint64, req *vehicle.UpdateVehiclePricingRequest) (*vehicle.UpdateVehiclePricingResponse, error) {
	args := m.Called(ctx, vehicleID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vehicle.UpdateVehiclePricingResponse), args.Error(1)
}

func noopMiddleware(c *gin.Context) { c.Next() }

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := engine.Group("/api/v1")

	mockSvc := new(mockRoutesService)
	handler := vehicle.NewHandler(mockSvc)

	vehicle.RegisterRoutes(router, handler, noopMiddleware)

	routes := engine.Routes()
	assert.NotEmpty(t, routes)

	routeMap := map[string]bool{}
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = true
	}

	assert.True(t, routeMap["POST:/api/v1/vehicle"], "POST /api/v1/vehicle route should be registered")
	assert.True(t, routeMap["GET:/api/v1/vehicle/listing"], "GET /api/v1/vehicle/listing route should be registered")
	assert.True(t, routeMap["GET:/api/v1/vehicle/:id"], "GET /api/v1/vehicle/:id route should be registered")
	assert.True(t, routeMap["PATCH:/api/v1/vehicle/:id"], "PATCH /api/v1/vehicle/:id route should be registered")
	assert.True(t, routeMap["PATCH:/api/v1/vehicle/:id/pricing"], "PATCH /api/v1/vehicle/:id/pricing route should be registered")
}

func TestRegisterRoutes_NoopMiddlewareUsed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := engine.Group("/api/v1")

	mockSvc := new(mockRoutesService)
	handler := vehicle.NewHandler(mockSvc)

	called := false
	testMiddleware := func(c *gin.Context) {
		called = true
		c.AbortWithStatus(http.StatusTeapot)
	}

	vehicle.RegisterRoutes(router, handler, testMiddleware)

	routes := engine.Routes()
	routeMap := map[string]bool{}
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = true
	}
	assert.True(t, routeMap["GET:/api/v1/vehicle/:id"])
	_ = called
}

func TestRegisterPublicRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := engine.Group("/api/v1")

	mockSvc := new(mockRoutesService)
	handler := vehicle.NewHandler(mockSvc)

	vehicle.RegisterPublicRoutes(router, handler)

	routes := engine.Routes()
	routeMap := map[string]bool{}
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = true
	}
	assert.True(t, routeMap["GET:/api/v1/vehicle/public-listing"], "GET /api/v1/vehicle/public-listing route should be registered")
}
