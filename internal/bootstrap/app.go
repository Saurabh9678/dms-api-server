package bootstrap

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	infradb "infiour.local/dms-api-server/internal/infra/database"
	"infiour.local/dms-api-server/pkg/config"
	"infiour.local/dms-api-server/pkg/logger"
)

type App struct {
	cfg    *config.Config
	log    *slog.Logger
	db     *gorm.DB
	engine *gin.Engine
}

func NewApp() *App {
	cfg := config.MustLoad()
	log := logger.New(cfg.Env)

	db, err := infradb.Connect(cfg.Database.URL)
	if err != nil {
		log.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("failed to access sql db", "error", err)
		os.Exit(1)
	}

	deps := buildDependencies(cfg, db, log)
	engine := newRouter(cfg, log, deps, sqlDB)

	return &App{
		cfg:    cfg,
		log:    log,
		db:     db,
		engine: engine,
	}
}

func (a *App) Run() {
	addr := ":" + a.cfg.Server.Port
	a.log.Info("server starting", "addr", addr)
	if err := a.engine.Run(addr); err != nil {
		a.log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
