package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler) {
	userGroup := group.Group("/user")
	userGroup.PATCH("/me", h.UpdateProfile)
	userGroup.GET("/me", h.GetProfile)
}
