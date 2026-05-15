package router

import (
	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/api/http/handler"
)

type Handlers struct {
	AuthHandler *handler.AuthHandler
}

func SetUpRouter(engine *gin.Engine, handlers *Handlers) {
	api := engine.Group("/api/v1")
	RegisterAuthRoutes(api, handlers.AuthHandler)
}
