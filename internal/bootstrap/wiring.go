package bootstrap

import (
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"infiour.local/dms-api-server/pkg/config"
)

// BuildDependencies constructs all application dependencies from the provided
// database connection and configuration. Exported for use in tests and tooling.
func BuildDependencies(cfg *config.Config, db *gorm.DB, log *slog.Logger) *Dependencies {
	return buildDependencies(cfg, db, log)
}

// NewRouter assembles the HTTP engine with all routes and middleware wired up.
// Exported for use in tests and tooling.
func NewRouter(cfg *config.Config, log *slog.Logger, deps *Dependencies, sqlDB *sql.DB) *gin.Engine {
	return newRouter(cfg, log, deps, sqlDB)
}
