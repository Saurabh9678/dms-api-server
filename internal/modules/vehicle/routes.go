package vehicle

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler, showroomRoles gin.HandlerFunc) {
	group.POST("/vehicle", h.CreateVehicle)
	group.GET("/vehicle/listing", h.ListVehicles)
	group.GET("/vehicle/:id", showroomRoles, h.GetVehicle)
	group.PATCH("/vehicle/:id", showroomRoles, h.UpdateVehicle)
	group.PATCH("/vehicle/:id/pricing", showroomRoles, h.UpdateVehiclePricing)
	group.POST("/vehicle/:id/expense", showroomRoles, h.AddExpense)
	group.POST("/vehicle/:id/showroom", showroomRoles, h.AssignShowroom)
}

func RegisterPublicRoutes(group *gin.RouterGroup, h *Handler) {
	group.GET("/vehicle/public-listing", h.PublicListVehicles)
}
