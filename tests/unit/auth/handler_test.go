package auth_test

import (
	"bytes"
	"context"
	"errors"
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

func (f *fakeHandlerAuthService) VerifyOTP(_ context.Context, _ auth.VerifyOTPRequest) (*auth.VerifyOTPResponse, error) {
	return &auth.VerifyOTPResponse{
		AccessToken:  "a",
		RefreshToken: "r",
		ExpiresIn:    900,
		TokenType:    "Bearer",
		RequiredName: false,
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
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_DEVICE_CONTEXT"`) {
		t.Fatalf("expected INVALID_DEVICE_CONTEXT code, got %s", resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// fakeErrorAuthService — returns errors on every method
// ---------------------------------------------------------------------------

var errFakeService = errors.New("fake service error")

type fakeErrorAuthService struct{}

func (f *fakeErrorAuthService) Register(_ context.Context, _ auth.RegisterRequest) (*auth.TriggerOTPResponse, error) {
	return nil, errFakeService
}

func (f *fakeErrorAuthService) Login(_ context.Context, _ auth.LoginRequest) (*auth.TriggerOTPResponse, error) {
	return nil, errFakeService
}

func (f *fakeErrorAuthService) VerifyOTP(_ context.Context, _ auth.VerifyOTPRequest) (*auth.VerifyOTPResponse, error) {
	return nil, errFakeService
}

func (f *fakeErrorAuthService) RefreshToken(_ context.Context, _ auth.RefreshTokenRequest) (*auth.TokenResponse, error) {
	return nil, errFakeService
}

func (f *fakeErrorAuthService) Logout(_ context.Context, _ auth.LogoutRequest) error {
	return errFakeService
}

// ---------------------------------------------------------------------------
// Register handler tests
// ---------------------------------------------------------------------------

func TestRegister_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/register", h.Register)

	// Missing required phoneNumber field
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"countryCode":"+91"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST, got %s", resp.Body.String())
	}
}

func TestRegister_MissingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	// No X-Platform header
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_DEVICE_CONTEXT"`) {
		t.Fatalf("expected INVALID_DEVICE_CONTEXT, got %s", resp.Body.String())
	}
}

func TestRegister_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeErrorAuthService{})
	engine.POST("/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		t.Fatalf("expected non-200, got %d", resp.Code)
	}
}

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"success":true`) {
		t.Fatalf("expected success response, got %s", resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Login handler tests
// ---------------------------------------------------------------------------

func TestLogin_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeErrorAuthService{})
	engine.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		t.Fatalf("expected non-200, got %d", resp.Code)
	}
}

// ---------------------------------------------------------------------------
// VerifyOTP handler tests
// ---------------------------------------------------------------------------

func TestVerifyOTP_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/verify-otp", h.VerifyOTP)

	// otpCode missing
	req := httptest.NewRequest(http.MethodPost, "/verify-otp", bytes.NewBufferString(`{"requestId":"Ab12Cd34"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST, got %s", resp.Body.String())
	}
}

func TestVerifyOTP_MissingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/verify-otp", h.VerifyOTP)

	req := httptest.NewRequest(http.MethodPost, "/verify-otp", bytes.NewBufferString(`{"requestId":"Ab12Cd34","otpCode":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	// Missing X-Platform
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_DEVICE_CONTEXT"`) {
		t.Fatalf("expected INVALID_DEVICE_CONTEXT, got %s", resp.Body.String())
	}
}

func TestVerifyOTP_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeErrorAuthService{})
	engine.POST("/verify-otp", h.VerifyOTP)

	req := httptest.NewRequest(http.MethodPost, "/verify-otp", bytes.NewBufferString(`{"requestId":"Ab12Cd34","otpCode":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		t.Fatalf("expected non-200, got %d", resp.Code)
	}
}

// ---------------------------------------------------------------------------
// RefreshToken handler tests
// ---------------------------------------------------------------------------

func TestRefreshToken_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/refresh-token", h.RefreshToken)

	// Empty body — missing required refreshToken field
	req := httptest.NewRequest(http.MethodPost, "/refresh-token", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST, got %s", resp.Body.String())
	}
}

func TestRefreshToken_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeErrorAuthService{})
	engine.POST("/refresh-token", h.RefreshToken)

	req := httptest.NewRequest(http.MethodPost, "/refresh-token", bytes.NewBufferString(`{"refreshToken":"some-token"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		t.Fatalf("expected non-200, got %d", resp.Code)
	}
}

// ---------------------------------------------------------------------------
// Logout handler tests
// ---------------------------------------------------------------------------

func TestLogout_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeErrorAuthService{})
	engine.POST("/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "d-1")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code == http.StatusOK {
		t.Fatalf("expected non-200, got %d", resp.Code)
	}
}

// ---------------------------------------------------------------------------
// extractBearerToken edge cases
// ---------------------------------------------------------------------------

func TestExtractBearerToken_EmptyBearerValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/logout", h.Logout)

	// Authorization header has "Bearer " prefix but token part is whitespace only
	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer   ")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty bearer token, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST, got %s", resp.Body.String())
	}
}

func TestExtractBearerToken_NoBearerPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/logout", h.Logout)

	// Authorization header does not start with "Bearer "
	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token some-token")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-Bearer Authorization, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST, got %s", resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Login handler: missing device context headers
// ---------------------------------------------------------------------------

func TestLogin_MissingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := auth.NewHandler(&fakeHandlerAuthService{})
	engine.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"countryCode":"+91","phoneNumber":"9999999999"}`))
	req.Header.Set("Content-Type", "application/json")
	// Missing X-Platform and X-Device-Id
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_DEVICE_CONTEXT"`) {
		t.Fatalf("expected INVALID_DEVICE_CONTEXT, got %s", resp.Body.String())
	}
}
