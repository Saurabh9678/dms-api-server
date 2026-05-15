package auth_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/modules/auth"
)

type fakeHandlerAuthService struct{}

func (f *fakeHandlerAuthService) Register(_ context.Context, _ auth.RegisterRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "ok"}, nil
}

func (f *fakeHandlerAuthService) Login(_ context.Context, _ auth.LoginRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "ok"}, nil
}

func (f *fakeHandlerAuthService) VerifyOTP(_ context.Context, _ auth.VerifyOTPRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{
		AccessToken:  "a",
		RefreshToken: "r",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (f *fakeHandlerAuthService) RefreshToken(_ context.Context, _ auth.RefreshTokenRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{
		AccessToken:  "a2",
		RefreshToken: "r2",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (f *fakeHandlerAuthService) Logout(_ context.Context, _ auth.LogoutRequest) error {
	return nil
}

func TestLoginBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"countryCode":"+91"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"success":false`) {
		t.Fatalf("expected error response envelope, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST code, got %s", resp.Body.String())
	}
}

func TestVerifyOTPSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/verify-otp", h.VerifyOTP)

	req := httptest.NewRequest(http.MethodPost, "/verify-otp", bytes.NewBufferString(`{"requestId":"Ab12Cd34","otpCode":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"success":true`) {
		t.Fatalf("expected success response envelope, got %s", resp.Body.String())
	}
}

func TestLogoutRequiresAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST code, got %s", resp.Body.String())
	}
}

func TestLogoutSuccessWithAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("X-Platform", "web")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"success":true`) {
		t.Fatalf("expected success response envelope, got %s", resp.Body.String())
	}
}

func TestLogoutRequiresPlatformHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer access-token")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST code, got %s", resp.Body.String())
	}
}
