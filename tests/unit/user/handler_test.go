package user_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/modules/user"
	"infiour.local/dms-api-server/pkg/middleware"
)

type fakeHandlerUpdateService struct {
	result *user.UpdateProfileResponse
	err    error
}

func (f *fakeHandlerUpdateService) UpdateProfile(_ context.Context, _ uint64, _ user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	return f.result, f.err
}

func TestUpdateProfileSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	service := &fakeHandlerUpdateService{
		result: &user.UpdateProfileResponse{Name: "John Doe"},
	}
	h := user.NewHandler(service)
	engine.PATCH("/user/me", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		h.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/user/me", bytes.NewBufferString(`{"name":"John Doe"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"success":true`) {
		t.Fatalf("expected success response envelope, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"John Doe"`) {
		t.Fatalf("expected name in response, got %s", resp.Body.String())
	}
}

func TestUpdateProfileEmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	service := &fakeHandlerUpdateService{}
	h := user.NewHandler(service)
	engine.PATCH("/user/me", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		h.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/user/me", bytes.NewBufferString(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"success":false`) {
		t.Fatalf("expected error response, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("expected INVALID_REQUEST code, got %s", resp.Body.String())
	}
}

func TestUpdateProfileMissingUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	service := &fakeHandlerUpdateService{}
	h := user.NewHandler(service)
	engine.PATCH("/user/me", h.UpdateProfile)

	req := httptest.NewRequest(http.MethodPatch, "/user/me", bytes.NewBufferString(`{"name":"John Doe"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"code":"INVALID_ACCESS_TOKEN"`) {
		t.Fatalf("expected INVALID_ACCESS_TOKEN code, got %s", resp.Body.String())
	}
}

func TestUpdateProfileServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	service := &fakeHandlerUpdateService{
		err: errors.New("database error"),
	}
	h := user.NewHandler(service)
	engine.PATCH("/user/me", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		h.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/user/me", bytes.NewBufferString(`{"name":"John Doe"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"success":false`) {
		t.Fatalf("expected error response, got %s", resp.Body.String())
	}
}

func TestUpdateProfileMissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	service := &fakeHandlerUpdateService{}
	h := user.NewHandler(service)
	engine.PATCH("/user/me", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(42))
		h.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/user/me", bytes.NewBufferString(`{}`))
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
