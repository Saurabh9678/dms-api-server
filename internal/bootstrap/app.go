package bootstrap

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	infradb "infiour.local/dms-api-server/internal/infra/database"
	"infiour.local/dms-api-server/pkg/config"
	"infiour.local/dms-api-server/pkg/logger"
)

// Package-level vars allow injection in tests without changing function signatures.
var (
	osExit    = os.Exit
	connectDB = func(dsn string) (*gorm.DB, error) { return infradb.Connect(dsn) }
	getSQLDB  = func(db *gorm.DB) (*sql.DB, error) { return db.DB() }
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

	db, err := connectDB(cfg.Database.URL)
	if err != nil {
		log.Error("failed to connect database", "error", err)
		osExit(1)
		return nil
	}

	sqlDB, err := getSQLDB(db)
	if err != nil {
		log.Error("failed to access sql db", "error", err)
		osExit(1)
		return nil
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
		osExit(1)
	}
}
