package showroom

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *Handler, showroomRolesMW gin.HandlerFunc) {
	sg := group.Group("/showroom")
	sg.POST("", h.CreateShowroom)
	sg.GET("", h.ListShowrooms)

	mg := sg.Group("")
	mg.Use(showroomRolesMW)
	mg.PATCH("/:id", h.UpdateShowroom)
	mg.POST("/:id/member", h.AddMember)
	mg.GET("/:id/member", h.ListMembers)
	mg.DELETE("/:id/member/:user_id", h.RemoveMember)
	mg.PATCH("/:id/member/:user_id", h.UpdateMemberRole)
}
