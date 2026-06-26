package bootstrap

import (
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/modules/auth"
	"infiour.local/dms-api-server/internal/modules/dashboard"
	"infiour.local/dms-api-server/internal/modules/showroom"
	"infiour.local/dms-api-server/internal/modules/user"
	"infiour.local/dms-api-server/internal/modules/vehicle"
	"infiour.local/dms-api-server/pkg/config"
	"infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/middleware"
	"infiour.local/dms-api-server/pkg/response"
)

func newRouter(cfg *config.Config, log *slog.Logger, deps *Dependencies, sqlDB *sql.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.RequestLog(log),
		middleware.Recovery(log),
	)

	engine.GET("/health", func(c *gin.Context) {
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			response.Error(c, 500, errors.CodeInternal, "database unavailable")
			return
		}
		response.OK(c, "ok", map[string]any{"status": "ok"})
	})

	api := engine.Group("/api/v1")
	api.Use(middleware.RequireDeviceContext())
	auth.RegisterRoutes(api, deps.AuthHandler)
	vehicle.RegisterPublicRoutes(api, deps.VehicleHandler)

	protected := api.Group("")
	protected.Use(middleware.RequireAuth(deps.TokenProvider))
	user.RegisterRoutes(protected, deps.UserHandler)
	vehicle.RegisterRoutes(protected, deps.VehicleHandler, deps.ShowroomRolesMiddleware)
	dashboard.RegisterRoutes(protected, deps.DashboardHandler)
	showroom.RegisterRoutes(protected, deps.ShowroomHandler, deps.ShowroomRolesMiddleware)

	return engine
}
