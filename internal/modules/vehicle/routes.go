package vehicle

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler) {
	group.POST("/vehicle", h.CreateVehicle)
}
