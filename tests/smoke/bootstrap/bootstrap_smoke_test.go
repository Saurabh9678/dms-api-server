package bootstrap_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/internal/bootstrap"
	"infiour.local/dms-api-server/pkg/config"
)

func newMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return gormDB, mock
}

func TestBootstrap_BuildDependencies(t *testing.T) {
	gormDB, _ := newMockGormDB(t)
	cfg := &config.Config{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	deps := bootstrap.BuildDependencies(cfg, gormDB, log)

	assert.NotNil(t, deps)
	assert.NotNil(t, deps.ShowroomRolesMiddleware)
}

func TestBootstrap_NewRouter_DevelopmentMode(t *testing.T) {
	gormDB, _ := newMockGormDB(t)
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.Config{Env: "development"}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := bootstrap.BuildDependencies(cfg, gormDB, log)

	engine := bootstrap.NewRouter(cfg, log, deps, sqlDB)
	assert.NotNil(t, engine)

	routeMap := map[string]bool{}
	for _, r := range engine.Routes() {
		routeMap[r.Method+":"+r.Path] = true
	}
	assert.True(t, routeMap["GET:/health"])
	assert.True(t, routeMap["POST:/api/v1/vehicle"])
	assert.True(t, routeMap["GET:/api/v1/vehicle/listing"])
	assert.True(t, routeMap["GET:/api/v1/vehicle/:id"])
}

func TestBootstrap_NewRouter_ProductionMode(t *testing.T) {
	gormDB, _ := newMockGormDB(t)
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.Config{Env: "production"}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := bootstrap.BuildDependencies(cfg, gormDB, log)

	engine := bootstrap.NewRouter(cfg, log, deps, sqlDB)
	assert.NotNil(t, engine)
}
