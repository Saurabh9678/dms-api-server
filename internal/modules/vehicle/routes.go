package vehicle

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler, showroomRoles gin.HandlerFunc) {
	group.POST("/vehicle", h.CreateVehicle)
	group.GET("/vehicle/listing", showroomRoles, h.ListVehicles)
	group.GET("/vehicle/:id", showroomRoles, h.GetVehicle)
	group.PATCH("/vehicle/:id", showroomRoles, h.UpdateVehicle)
	group.PATCH("/vehicle/:id/pricing", showroomRoles, h.UpdateVehiclePricing)
	group.POST("/vehicle/:id/expense", showroomRoles, h.AddExpense)
	group.POST("/vehicle/:id/status", showroomRoles, h.UpdateVehicleStatus)
	group.POST("/vehicle/:id/sale", showroomRoles, h.SellVehicle)
	group.POST("/vehicle/:id/showroom", showroomRoles, h.AssignShowroom)
	group.POST("/vehicle/:id/image", showroomRoles, h.AddVehicleImage)
	group.DELETE("/vehicle/:id/image/:image_id", showroomRoles, h.DeleteVehicleImage)
}

func RegisterPublicRoutes(group *gin.RouterGroup, h *Handler) {
	group.GET("/vehicle/public-listing", h.PublicListVehicles)
}
