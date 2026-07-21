package user_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/modules/user"
	"infiour.local/dms-api-server/pkg/middleware"
)

type fakeRoutesService struct{}

func (f *fakeRoutesService) UpdateProfile(_ context.Context, _ uint64, _ user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	return &user.UpdateProfileResponse{Name: "John Doe"}, nil
}

func (f *fakeRoutesService) GetProfile(_ context.Context, _ uint64) (*user.GetProfileResponse, error) {
	name := "Alice"
	countryCode := "+91"
	phone := "9999999999"
	return &user.GetProfileResponse{
		Name:          &name,
		CountryCode:   &countryCode,
		PhoneNumber:   &phone,
		ShowroomRoles: []user.ShowroomRole{},
		RequiredName:  false,
		HasShowrooms:  true,
		HasVehicles:   false,
	}, nil
}

func TestRegisterRoutesSuccessfulPatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api/v1")
	api.Use(middleware.RequireDeviceContext())
	protected := api.Group("")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		c.Next()
	})
	h := user.NewHandler(&fakeRoutesService{})
	user.RegisterRoutes(protected, h)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/user/me", bytes.NewBufferString(`{"name":"John Doe"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestRegisterRoutesUndefinedPath404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api/v1")
	api.Use(middleware.RequireDeviceContext())
	protected := api.Group("")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		c.Next()
	})
	h := user.NewHandler(&fakeRoutesService{})
	user.RegisterRoutes(protected, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/undefined", nil)
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestRegisterRoutesSuccessfulGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api/v1")
	api.Use(middleware.RequireDeviceContext())
	protected := api.Group("")
	protected.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		c.Next()
	})
	h := user.NewHandler(&fakeRoutesService{})
	user.RegisterRoutes(protected, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"has_showrooms":true`) {
		t.Fatalf("expected has_showrooms in response, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"country_code":"+91"`) {
		t.Fatalf("expected separate country_code in response, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"phone_number":"9999999999"`) {
		t.Fatalf("expected local phone_number in response, got %s", resp.Body.String())
	}
}
