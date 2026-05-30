package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/pkg/middleware"
)

type mockShowroomRolesProvider struct {
	roles map[uint64]string
	err   error
}

func (m *mockShowroomRolesProvider) LoadUserShowroomRoles(_ context.Context, _ uint64) (map[uint64]string, error) {
	return m.roles, m.err
}

func TestRequireShowroomRoles_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockShowroomRolesProvider{}
	engine := gin.New()
	engine.Use(middleware.RequireShowroomRoles(provider))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireShowroomRoles_WrongUserIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockShowroomRolesProvider{}
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "not-a-uint64")
		c.Next()
	})
	engine.Use(middleware.RequireShowroomRoles(provider))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireShowroomRoles_ProviderError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockShowroomRolesProvider{err: errors.New("db error")}
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		c.Next()
	})
	engine.Use(middleware.RequireShowroomRoles(provider))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequireShowroomRoles_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockShowroomRolesProvider{
		roles: map[uint64]string{1: "owner", 2: "manager"},
	}
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		c.Next()
	})
	engine.Use(middleware.RequireShowroomRoles(provider))
	engine.GET("/test", func(c *gin.Context) {
		val, exists := c.Get(middleware.ContextKeyShowroomRoles)
		assert.True(t, exists)
		roles := val.(map[uint64]string)
		assert.Equal(t, "owner", roles[1])
		assert.Equal(t, "manager", roles[2])
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireShowroomRoles_EmptyRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &mockShowroomRolesProvider{roles: map[uint64]string{}}
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(99))
		c.Next()
	})
	engine.Use(middleware.RequireShowroomRoles(provider))
	engine.GET("/test", func(c *gin.Context) {
		val, exists := c.Get(middleware.ContextKeyShowroomRoles)
		assert.True(t, exists)
		roles := val.(map[uint64]string)
		assert.Empty(t, roles)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
