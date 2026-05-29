package bootstrap

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/pkg/config"
)

func newTestGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return gormDB, mock
}

func saveHooks() (exitFn func(int), connectFn func(string) (*gorm.DB, error), sqlDBFn func(*gorm.DB) (*sql.DB, error)) {
	return osExit, connectDB, getSQLDB
}

func restoreHooks(exitFn func(int), connectFn func(string) (*gorm.DB, error), sqlDBFn func(*gorm.DB) (*sql.DB, error)) {
	osExit = exitFn
	connectDB = connectFn
	getSQLDB = sqlDBFn
}

// Compile-time check that exported names exist.
func TestExportsExist(t *testing.T) {
	assert.NotNil(t, NewRouter)
	assert.NotNil(t, BuildDependencies)
}

func TestNewApp_ExitsOnDBConnectError(t *testing.T) {
	savedExit, savedConnect, savedGetSQL := saveHooks()
	defer restoreHooks(savedExit, savedConnect, savedGetSQL)

	exitCode := -1
	osExit = func(code int) { exitCode = code }
	connectDB = func(_ string) (*gorm.DB, error) { return nil, errors.New("connect error") }

	result := NewApp()
	assert.Equal(t, 1, exitCode)
	assert.Nil(t, result)
}

func TestNewApp_ExitsOnSQLDBError(t *testing.T) {
	gormDB, _ := newTestGormDB(t)
	savedExit, savedConnect, savedGetSQL := saveHooks()
	defer restoreHooks(savedExit, savedConnect, savedGetSQL)

	exitCode := -1
	osExit = func(code int) { exitCode = code }
	connectDB = func(_ string) (*gorm.DB, error) { return gormDB, nil }
	getSQLDB = func(_ *gorm.DB) (*sql.DB, error) { return nil, errors.New("sql error") }

	result := NewApp()
	assert.Equal(t, 1, exitCode)
	assert.Nil(t, result)
}

func TestNewApp_Success(t *testing.T) {
	gormDB, _ := newTestGormDB(t)
	rawSQL, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = rawSQL.Close() })

	savedExit, savedConnect, savedGetSQL := saveHooks()
	defer restoreHooks(savedExit, savedConnect, savedGetSQL)

	osExit = func(code int) { t.Fatalf("unexpected os.Exit(%d)", code) }
	connectDB = func(_ string) (*gorm.DB, error) { return gormDB, nil }
	getSQLDB = func(_ *gorm.DB) (*sql.DB, error) { return rawSQL, nil }

	result := NewApp()
	assert.NotNil(t, result)
}

func TestRun_ExitsOnServerError(t *testing.T) {
	savedExit, savedConnect, savedGetSQL := saveHooks()
	defer restoreHooks(savedExit, savedConnect, savedGetSQL)

	exitCode := -1
	osExit = func(code int) { exitCode = code }

	gin.SetMode(gin.TestMode)
	app := &App{
		cfg:    &config.Config{Server: config.ServerConfig{Port: "abc"}},
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		engine: gin.New(),
	}
	app.Run()
	assert.Equal(t, 1, exitCode)
}

func TestNewRouter_HealthEndpoint_PingError(t *testing.T) {
	gormDB, _ := newTestGormDB(t)
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	mock.ExpectPing().WillReturnError(errors.New("db unavailable"))

	cfg := &config.Config{Env: "development"}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := buildDependencies(cfg, gormDB, log)
	engine := newRouter(cfg, log, deps, sqlDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewRouter_HealthEndpoint_PingSuccess(t *testing.T) {
	gormDB, _ := newTestGormDB(t)
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	mock.ExpectPing()

	cfg := &config.Config{Env: "development"}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps := buildDependencies(cfg, gormDB, log)
	engine := newRouter(cfg, log, deps, sqlDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
