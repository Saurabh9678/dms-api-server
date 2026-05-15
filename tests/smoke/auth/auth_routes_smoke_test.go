package smoke_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/modules/auth"
	"infiour.local/dms-api-server/pkg/middleware"
)

type smokeAuthService struct{}

func (s *smokeAuthService) Register(_ context.Context, _ auth.RegisterRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "ok"}, nil
}

func (s *smokeAuthService) Login(_ context.Context, _ auth.LoginRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "ok"}, nil
}

func (s *smokeAuthService) VerifyOTP(_ context.Context, _ auth.VerifyOTPRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{AccessToken: "a", RefreshToken: "r", ExpiresIn: 900, TokenType: "Bearer"}, nil
}

func (s *smokeAuthService) RefreshToken(_ context.Context, _ auth.RefreshTokenRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{AccessToken: "a2", RefreshToken: "r2", ExpiresIn: 900, TokenType: "Bearer"}, nil
}

func (s *smokeAuthService) Logout(_ context.Context, _ auth.LogoutRequest) error {
	return nil
}

func TestAuthLoginRouteShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api/v1")
	api.Use(middleware.RequireDeviceContext())
	auth.RegisterRoutes(api, auth.NewHandler(&smokeAuthService{}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestAuthLoginMissingDeviceContextHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api/v1")
	api.Use(middleware.RequireDeviceContext())
	auth.RegisterRoutes(api, auth.NewHandler(&smokeAuthService{}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_DEVICE_CONTEXT"`) {
		t.Fatalf("expected INVALID_DEVICE_CONTEXT code, got %s", resp.Body.String())
	}
}

func TestAuthLoginInvalidPlatformHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api/v1")
	api.Use(middleware.RequireDeviceContext())
	auth.RegisterRoutes(api, auth.NewHandler(&smokeAuthService{}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "watch")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_DEVICE_CONTEXT"`) {
		t.Fatalf("expected INVALID_DEVICE_CONTEXT code, got %s", resp.Body.String())
	}
}
