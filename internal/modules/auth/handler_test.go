package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockService is a test double for the Service interface.
type mockService struct {
	sendOTPFn      func(ctx context.Context, req SendOTPRequest) (*TriggerOTPResponse, error)
	registerFn     func(ctx context.Context, req RegisterRequest) (*TriggerOTPResponse, error)
	loginFn        func(ctx context.Context, req LoginRequest) (*TriggerOTPResponse, error)
	verifyOTPFn    func(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error)
	refreshTokenFn func(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error)
	logoutFn       func(ctx context.Context, req LogoutRequest) error
}

func (m *mockService) SendOTP(ctx context.Context, req SendOTPRequest) (*TriggerOTPResponse, error) {
	return m.sendOTPFn(ctx, req)
}
func (m *mockService) Register(ctx context.Context, req RegisterRequest) (*TriggerOTPResponse, error) {
	return m.registerFn(ctx, req)
}
func (m *mockService) Login(ctx context.Context, req LoginRequest) (*TriggerOTPResponse, error) {
	return m.loginFn(ctx, req)
}
func (m *mockService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error) {
	return m.verifyOTPFn(ctx, req)
}
func (m *mockService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error) {
	return m.refreshTokenFn(ctx, req)
}
func (m *mockService) Logout(ctx context.Context, req LogoutRequest) error {
	return m.logoutFn(ctx, req)
}

func newTestEngine(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	RegisterRoutes(e.Group("/api/v1"), h)
	return e
}

func deviceHeaders(req *http.Request) {
	req.Header.Set("X-Platform", "web")
	req.Header.Set("X-Device-Id", "device-123")
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ---- SendOTP ----------------------------------------------------------------

func TestSendOTP_ValidNumericPhone(t *testing.T) {
	svc := &mockService{
		sendOTPFn: func(_ context.Context, _ SendOTPRequest) (*TriggerOTPResponse, error) {
			return &TriggerOTPResponse{Message: "OTP sent successfully", RequestID: "ABCD1234"}, nil
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSendOTP_AlphanumericPhoneRejected(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "99adshbbfhk"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendOTP_AlphanumericCountryCodeRejected(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "+91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendOTP_MissingPhoneNumber(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "91"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendOTP_MissingCountryCode(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendOTP_MissingDeviceHeaders(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendOTP_ServiceError(t *testing.T) {
	svc := &mockService{
		sendOTPFn: func(_ context.Context, _ SendOTPRequest) (*TriggerOTPResponse, error) {
			return nil, errors.New("internal error")
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- Register ---------------------------------------------------------------

func TestRegister_ValidNumericPhone(t *testing.T) {
	svc := &mockService{
		registerFn: func(_ context.Context, _ RegisterRequest) (*TriggerOTPResponse, error) {
			return &TriggerOTPResponse{Message: "OTP sent successfully", RequestID: "ABCD1234"}, nil
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegister_AlphanumericPhoneRejected(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "99adshbbfhk"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_MissingDeviceHeaders(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_ServiceError(t *testing.T) {
	svc := &mockService{
		registerFn: func(_ context.Context, _ RegisterRequest) (*TriggerOTPResponse, error) {
			return nil, errors.New("internal error")
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- Login ------------------------------------------------------------------

func TestLogin_ValidNumericPhone(t *testing.T) {
	svc := &mockService{
		loginFn: func(_ context.Context, _ LoginRequest) (*TriggerOTPResponse, error) {
			return &TriggerOTPResponse{Message: "OTP sent successfully", RequestID: "ABCD1234"}, nil
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogin_AlphanumericPhoneRejected(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "99adshbbfhk"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_MissingDeviceHeaders(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_ServiceError(t *testing.T) {
	svc := &mockService{
		loginFn: func(_ context.Context, _ LoginRequest) (*TriggerOTPResponse, error) {
			return nil, errors.New("internal error")
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"countryCode": "91", "phoneNumber": "9876543210"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- VerifyOTP --------------------------------------------------------------

func TestVerifyOTP_Valid(t *testing.T) {
	svc := &mockService{
		verifyOTPFn: func(_ context.Context, _ VerifyOTPRequest) (*VerifyOTPResponse, error) {
			return &VerifyOTPResponse{AccessToken: "at", RefreshToken: "rt", TokenType: "Bearer"}, nil
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"requestId": "ABCD1234", "otpCode": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVerifyOTP_InvalidBinding(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	// otpCode must be exactly 6 numeric digits
	body := jsonBody(t, map[string]any{"requestId": "ABCD1234", "otpCode": "abc"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyOTP_MissingDeviceHeaders(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{"requestId": "ABCD1234", "otpCode": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-otp", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyOTP_ServiceError(t *testing.T) {
	svc := &mockService{
		verifyOTPFn: func(_ context.Context, _ VerifyOTPRequest) (*VerifyOTPResponse, error) {
			return nil, ErrInvalidOTP
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"requestId": "ABCD1234", "otpCode": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-otp", body)
	req.Header.Set("Content-Type", "application/json")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---- RefreshToken -----------------------------------------------------------

func TestRefreshToken_Valid(t *testing.T) {
	svc := &mockService{
		refreshTokenFn: func(_ context.Context, _ RefreshTokenRequest) (*TokenResponse, error) {
			return &TokenResponse{AccessToken: "at", RefreshToken: "rt", TokenType: "Bearer"}, nil
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"refreshToken": "some-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRefreshToken_InvalidBinding(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	body := jsonBody(t, map[string]any{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshToken_ServiceError(t *testing.T) {
	svc := &mockService{
		refreshTokenFn: func(_ context.Context, _ RefreshTokenRequest) (*TokenResponse, error) {
			return nil, ErrInvalidRefreshToken
		},
	}
	e := newTestEngine(NewHandler(svc))

	body := jsonBody(t, map[string]any{"refreshToken": "bad-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---- Logout -----------------------------------------------------------------

func TestLogout_Valid(t *testing.T) {
	svc := &mockService{
		logoutFn: func(_ context.Context, _ LogoutRequest) error { return nil },
	}
	e := newTestEngine(NewHandler(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer some-access-token")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogout_MissingAuthorizationHeader(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_MissingBearerPrefix(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "some-access-token")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_MissingDeviceHeaders(t *testing.T) {
	e := newTestEngine(NewHandler(&mockService{}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer some-access-token")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_ServiceError(t *testing.T) {
	svc := &mockService{
		logoutFn: func(_ context.Context, _ LogoutRequest) error { return ErrInvalidAccessToken },
	}
	e := newTestEngine(NewHandler(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer some-access-token")
	deviceHeaders(req)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
