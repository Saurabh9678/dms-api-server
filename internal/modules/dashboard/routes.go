package dashboard

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler) {
	group.GET("/dashboard", h.GetDashboard)
}
