package middleware_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/pkg/middleware"
)

type mockTokenParser struct {
	userID uint64
	err    error
}

func (m *mockTokenParser) ParseAccessToken(_ string) (uint64, error) {
	return m.userID, m.err
}

func TestRequireAuth_NoAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.GET("/", middleware.RequireAuth(&mockTokenParser{userID: 1}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_InvalidBearerFormat(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.GET("/", middleware.RequireAuth(&mockTokenParser{userID: 1}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc123")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_ParseError(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.GET("/", middleware.RequireAuth(&mockTokenParser{err: errors.New("invalid")}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_Success(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	var gotUserID uint64
	engine.GET("/", middleware.RequireAuth(&mockTokenParser{userID: 42}), func(c *gin.Context) {
		val, _ := c.Get(middleware.ContextKeyUserID)
		gotUserID, _ = val.(uint64)
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint64(42), gotUserID)
}

func TestRecovery_NoPanic(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.Recovery(log))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecovery_PanicReturns500(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.Recovery(log))
	engine.GET("/", func(c *gin.Context) { panic("boom") })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequestID_WithExistingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "my-fixed-id")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "my-fixed-id", w.Header().Get("X-Request-ID"))
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequireDeviceContext_MissingHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.RequireDeviceContext())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequireDeviceContext_InvalidPlatform(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.RequireDeviceContext())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Platform", "watch")
	req.Header.Set("X-Device-Id", "device-1")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequireDeviceContext_Success(t *testing.T) {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.RequireDeviceContext())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "device-1")
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestLog(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(middleware.RequestLog(log))
	engine.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
