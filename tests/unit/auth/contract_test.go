package auth_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/modules/auth"
)

type contractService struct{}

func (f *contractService) Register(_ context.Context, _ auth.RegisterRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "If the account is valid, an OTP has been sent"}, nil
}

func (f *contractService) Login(_ context.Context, _ auth.LoginRequest) (*auth.TriggerOTPResponse, error) {
	return &auth.TriggerOTPResponse{Message: "If the account is valid, an OTP has been sent"}, nil
}

func (f *contractService) VerifyOTP(_ context.Context, _ auth.VerifyOTPRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{AccessToken: "a", RefreshToken: "r", ExpiresIn: 900, TokenType: "Bearer"}, nil
}

func (f *contractService) RefreshToken(_ context.Context, _ auth.RefreshTokenRequest) (*auth.TokenResponse, error) {
	return &auth.TokenResponse{AccessToken: "a2", RefreshToken: "r2", ExpiresIn: 900, TokenType: "Bearer"}, nil
}

func (f *contractService) Logout(_ context.Context, _ auth.LogoutRequest) error {
	return nil
}

func TestAuthRouteContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	api := engine.Group("/api/v1")
	auth.RegisterRoutes(api, auth.NewHandler(&contractService{}))

	cases := []struct {
		name       string
		path       string
		body       string
		statusCode int
	}{
		{name: "register", path: "/api/v1/auth/register", body: `{"countryCode":"+91","phoneNumber":"9999999999","platform":"web","deviceId":"d-1"}`, statusCode: http.StatusOK},
		{name: "login", path: "/api/v1/auth/login", body: `{"countryCode":"+91","phoneNumber":"9999999999","platform":"web","deviceId":"d-1"}`, statusCode: http.StatusOK},
		{name: "verify-otp", path: "/api/v1/auth/verify-otp", body: `{"countryCode":"+91","phoneNumber":"9999999999","otpCode":"123456","platform":"web","deviceId":"d-1"}`, statusCode: http.StatusOK},
		{name: "refresh-token", path: "/api/v1/auth/refresh-token", body: `{"refreshToken":"r"}`, statusCode: http.StatusOK},
		{name: "logout", path: "/api/v1/auth/logout", body: `{"refreshToken":"r2"}`, statusCode: http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)

			if resp.Code != tc.statusCode {
				t.Fatalf("expected %d, got %d", tc.statusCode, resp.Code)
			}
		})
	}
}
