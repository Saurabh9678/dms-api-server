package vehicle

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler, showroomRoles gin.HandlerFunc) {
	group.POST("/vehicle", h.CreateVehicle)
	group.GET("/vehicle/listing", h.ListVehicles)
	group.GET("/vehicle/:id", showroomRoles, h.GetVehicle)
}
