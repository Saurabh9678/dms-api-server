package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(routerGroup *gin.RouterGroup, authHandler *Handler) {
	auth := routerGroup.Group("/auth")
	auth.POST("/send-otp", authHandler.SendOTP)
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/verify-otp", authHandler.VerifyOTP)
	auth.POST("/refresh-token", authHandler.RefreshToken)
	auth.POST("/logout", authHandler.Logout)
}
