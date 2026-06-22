package showroom

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler) {
	showroomGroup := group.Group("/showroom")
	showroomGroup.POST("", h.CreateShowroom)
}
