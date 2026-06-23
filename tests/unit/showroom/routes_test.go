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
	nopMW := func(c *gin.Context) { c.Next() }

	showroom.RegisterRoutes(router, handler, nopMW)

	routeMap := map[string]bool{}
	for _, r := range engine.Routes() {
		routeMap[r.Method+":"+r.Path] = true
	}

	assert.True(t, routeMap["POST:/api/v1/showroom"], "POST /api/v1/showroom should be registered")
	assert.True(t, routeMap["POST:/api/v1/showroom/:id/member"], "POST /api/v1/showroom/:id/member should be registered")
	assert.True(t, routeMap["GET:/api/v1/showroom/:id/member"], "GET /api/v1/showroom/:id/member should be registered")
	assert.True(t, routeMap["DELETE:/api/v1/showroom/:id/member/:user_id"], "DELETE /api/v1/showroom/:id/member/:user_id should be registered")
	assert.True(t, routeMap["PATCH:/api/v1/showroom/:id/member/:user_id"], "PATCH /api/v1/showroom/:id/member/:user_id should be registered")
}
