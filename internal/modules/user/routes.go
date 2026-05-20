package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler) {
	group.PATCH("/user/me", h.UpdateProfile)
}
