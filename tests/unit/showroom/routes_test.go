package showroom_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/showroom"
)

func TestRegisterShowroomRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := engine.Group("/api/v1")

	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	showroom.RegisterRoutes(router, handler)

	routeMap := map[string]bool{}
	for _, r := range engine.Routes() {
		routeMap[r.Method+":"+r.Path] = true
	}

	assert.True(t, routeMap["POST:/api/v1/showroom"], "POST /api/v1/showroom route should be registered")
}
