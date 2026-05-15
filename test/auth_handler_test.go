package diff_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/api/http/handler"
	"infiour.local/dms-api-server/internal/application/auth"
)

type fakeAuthService struct{}

func (f *fakeAuthService) Register(_ context.Context, _ auth.RegisterRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "ok"}, nil
}

func (f *fakeAuthService) Login(_ context.Context, _ auth.LoginRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "ok"}, nil
}

func (f *fakeAuthService) VerifyOTP(_ context.Context, _ auth.VerifyOTPRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{
		AccessToken:  "a",
		RefreshToken: "r",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (f *fakeAuthService) RefreshToken(_ context.Context, _ auth.RefreshTokenRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{
		AccessToken:  "a2",
		RefreshToken: "r2",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (f *fakeAuthService) Logout(_ context.Context, _ auth.LogoutRequest) error {
	return nil
}

func TestLoginBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	authHandler := handler.NewAuthHandler(&fakeAuthService{})
	engine.POST("/login", authHandler.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"countryCode":"+91"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestVerifyOTPSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	authHandler := handler.NewAuthHandler(&fakeAuthService{})
	engine.POST("/verify-otp", authHandler.VerifyOTP)

	req := httptest.NewRequest(http.MethodPost, "/verify-otp", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999","otpCode":"123456","platform":"web","deviceId":"d-1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}
