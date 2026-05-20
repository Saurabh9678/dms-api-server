package vehicle

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRoutesService struct {
	mock.Mock
}

func (m *mockRoutesService) CreateVehicle(ctx context.Context, req *CreateVehicleRequest) (*CreateVehicleResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CreateVehicleResponse), args.Error(1)
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := engine.Group("/api/v1")

	mockSvc := new(mockRoutesService)
	handler := NewHandler(mockSvc)

	RegisterRoutes(router, handler)

	routes := engine.Routes()
	assert.NotEmpty(t, routes)

	found := false
	for _, route := range routes {
		if route.Path == "/api/v1/vehicle" && route.Method == "POST" {
			found = true
			break
		}
	}
	assert.True(t, found, "POST /api/v1/vehicle route should be registered")
}
