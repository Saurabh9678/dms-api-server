package wire

import (
	"github.com/gin-gonic/gin"
	httprouter "infiour.local/dms-api-server/internal/api/http/router"
	"infiour.local/dms-api-server/internal/infra/config"
	"infiour.local/dms-api-server/internal/infra/db"
)

func BuildServer() (*gin.Engine, error) {
	dbConfig := config.LoadDBConfig()

	dbProvider := db.NewPostgresProvider(dbConfig.URL)
	database, err := db.Connect(dbProvider)
	if err != nil {
		return nil, err
	}

	authHandler, err := BuildAuthHandler(database)
	if err != nil {
		return nil, err
	}

	engine := gin.Default()
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "DMS API server is running",
		})
	})

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	httprouter.SetUpRouter(engine, &httprouter.Handlers{
		AuthHandler: authHandler,
	})

	return engine, nil
}
