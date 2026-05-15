package router

import (
	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/api/http/handler"
)

func RegisterAuthRoutes(routerGroup *gin.RouterGroup, authHandler *handler.AuthHandler) {

	auth := routerGroup.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/verify-otp", authHandler.VerifyOTP)
	auth.POST("/refresh-token", authHandler.RefreshToken)
	auth.POST("/logout", authHandler.Logout)
}
