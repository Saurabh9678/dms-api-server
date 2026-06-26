package showroom_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"infiour.local/dms-api-server/internal/modules/showroom"
	"infiour.local/dms-api-server/pkg/middleware"
)

// ─── Mock service ─────────────────────────────────────────────────────────────

type mockShowroomService struct {
	mock.Mock
}

func (m *mockShowroomService) CreateShowroom(ctx context.Context, userID uint64, req *showroom.CreateShowroomRequest, logo, banner *multipart.FileHeader) (*showroom.CreateShowroomResponse, error) {
	args := m.Called(ctx, userID, req, logo, banner)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.CreateShowroomResponse), args.Error(1)
}

func (m *mockShowroomService) AddMember(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, req *showroom.AddMemberRequest) (*showroom.AddMemberResponse, error) {
	args := m.Called(ctx, callerRoles, showroomID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.AddMemberResponse), args.Error(1)
}

func (m *mockShowroomService) ListMembers(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, page, limit int) (*showroom.ListMembersResponse, error) {
	args := m.Called(ctx, callerRoles, showroomID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.ListMembersResponse), args.Error(1)
}

func (m *mockShowroomService) RemoveMember(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64) error {
	args := m.Called(ctx, callerUserID, callerRoles, showroomID, targetUserID)
	return args.Error(0)
}

func (m *mockShowroomService) UpdateMemberRole(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64, req *showroom.UpdateMemberRoleRequest) (*showroom.AddMemberResponse, error) {
	args := m.Called(ctx, callerUserID, callerRoles, showroomID, targetUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.AddMemberResponse), args.Error(1)
}

func (m *mockShowroomService) UpdateShowroom(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID uint64, req *showroom.UpdateShowroomRequest, logo, banner *multipart.FileHeader) (*showroom.CreateShowroomResponse, error) {
	args := m.Called(ctx, callerUserID, callerRoles, showroomID, req, logo, banner)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.CreateShowroomResponse), args.Error(1)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func setupShowroomEngine(h *showroom.Handler, userID uint64, roles map[uint64]string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		if userID > 0 {
			c.Set(middleware.ContextKeyUserID, userID)
		}
		if roles != nil {
			c.Set(middleware.ContextKeyShowroomRoles, roles)
		}
	})
	engine.POST("/showroom", h.CreateShowroom)
	engine.PATCH("/showroom/:id", h.UpdateShowroom)
	engine.POST("/showroom/:id/member", h.AddMember)
	engine.GET("/showroom/:id/member", h.ListMembers)
	engine.DELETE("/showroom/:id/member/:user_id", h.RemoveMember)
	engine.PATCH("/showroom/:id/member/:user_id", h.UpdateMemberRole)
	return engine
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

// ─── CreateShowroom ───────────────────────────────────────────────────────────

func TestHandler_CreateShowroom_NoUserID(t *testing.T) {
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	gin.SetMode(gin.TestMode)
	engine.POST("/showroom", handler.CreateShowroom)

	req := httptest.NewRequest(http.MethodPost, "/showroom", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertNotCalled(t, "CreateShowroom")
}

func TestHandler_CreateShowroom_BadUserIDType(t *testing.T) {
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	gin.SetMode(gin.TestMode)
	engine.POST("/showroom", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "not-a-uint64")
		handler.CreateShowroom(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/showroom", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_CreateShowroom_ParseMultipartError(t *testing.T) {
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	gin.SetMode(gin.TestMode)
	engine.POST("/showroom", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		handler.CreateShowroom(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/showroom", bytes.NewBufferString("not-multipart"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateShowroom_ServiceError(t *testing.T) {
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	gin.SetMode(gin.TestMode)
	engine.POST("/showroom", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		handler.CreateShowroom(c)
	})

	mockSvc.On("CreateShowroom", mock.Anything, uint64(1), mock.Anything, (*multipart.FileHeader)(nil), (*multipart.FileHeader)(nil)).
		Return(nil, errors.New("service error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "Test")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/showroom", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_CreateShowroom_Success(t *testing.T) {
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	gin.SetMode(gin.TestMode)
	engine.POST("/showroom", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		handler.CreateShowroom(c)
	})

	logoPath := "path/to/logo.jpg"
	mockSvc.On("CreateShowroom", mock.Anything, uint64(1), mock.Anything, (*multipart.FileHeader)(nil), (*multipart.FileHeader)(nil)).
		Return(&showroom.CreateShowroomResponse{
			ID:           1,
			Name:         "Test",
			ShowroomLogo: &logoPath,
		}, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "Test")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/showroom", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_CreateShowroom_WithFiles(t *testing.T) {
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	gin.SetMode(gin.TestMode)
	engine.POST("/showroom", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		handler.CreateShowroom(c)
	})

	mockSvc.On("CreateShowroom", mock.Anything, uint64(1), mock.Anything,
		mock.AnythingOfType("*multipart.FileHeader"),
		mock.AnythingOfType("*multipart.FileHeader"),
	).Return(&showroom.CreateShowroomResponse{ID: 2, Name: "Files"}, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "Files")
	logoPart, _ := writer.CreateFormFile("showroom_logo", "logo.jpg")
	_, _ = logoPart.Write([]byte("logo-content"))
	bannerPart, _ := writer.CreateFormFile("showroom_banner", "banner.jpg")
	_, _ = bannerPart.Write([]byte("banner-content"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/showroom", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── AddMember ────────────────────────────────────────────────────────────────

func TestHandler_AddMember_InvalidShowroomID(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodPost, "/showroom/abc/member", jsonBody(map[string]any{"user_id": 99, "role": "employee"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_AddMember_MissingShowroomRoles(t *testing.T) {
	mockSvc := new(mockShowroomService)
	// roles = nil → context key not set
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, nil)

	req := httptest.NewRequest(http.MethodPost, "/showroom/1/member", jsonBody(map[string]any{"user_id": 99, "role": "employee"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_AddMember_InvalidBody(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodPost, "/showroom/1/member", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_AddMember_ServiceError(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("AddMember", mock.Anything, mock.Anything, uint64(1), mock.Anything).
		Return(nil, errors.New("service error"))

	req := httptest.NewRequest(http.MethodPost, "/showroom/1/member", jsonBody(map[string]any{"user_id": 99, "role": "employee"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_AddMember_Success(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("AddMember", mock.Anything, mock.Anything, uint64(1), mock.Anything).
		Return(&showroom.AddMemberResponse{ShowroomID: 1, UserID: 99, Role: "employee"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/showroom/1/member", jsonBody(map[string]any{"user_id": 99, "role": "employee"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── ListMembers ─────────────────────────────────────────────────────────────

func TestHandler_AddMember_BadShowroomRolesType(t *testing.T) {
	// ContextKeyShowroomRoles set to wrong type → 401
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockShowroomService)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		c.Set(middleware.ContextKeyShowroomRoles, "wrong-type") // not map[uint64]string
	})
	engine.POST("/showroom/:id/member", showroom.NewHandler(mockSvc).AddMember)

	req := httptest.NewRequest(http.MethodPost, "/showroom/1/member", jsonBody(map[string]any{"user_id": 99, "role": "employee"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_ListMembers_InvalidPageFallsToDefault(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	// page=abc is invalid → falls to default 1; limit=0 is below min → falls to default 20
	mockSvc.On("ListMembers", mock.Anything, mock.Anything, uint64(1), 1, 20).
		Return(&showroom.ListMembersResponse{Members: []showroom.MemberItem{}, Total: 0, Page: 1, Limit: 20}, nil)

	req := httptest.NewRequest(http.MethodGet, "/showroom/1/member?page=abc&limit=0", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_ListMembers_InvalidShowroomID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodGet, "/showroom/bad/member", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_ListMembers_MissingShowroomRoles(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/showroom/1/member", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_ListMembers_ServiceError(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("ListMembers", mock.Anything, mock.Anything, uint64(1), 1, 20).
		Return(nil, errors.New("service error"))

	req := httptest.NewRequest(http.MethodGet, "/showroom/1/member", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_ListMembers_DefaultPagination(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("ListMembers", mock.Anything, mock.Anything, uint64(1), 1, 20).
		Return(&showroom.ListMembersResponse{Members: []showroom.MemberItem{}, Total: 0, Page: 1, Limit: 20}, nil)

	req := httptest.NewRequest(http.MethodGet, "/showroom/1/member", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_ListMembers_CustomPagination(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("ListMembers", mock.Anything, mock.Anything, uint64(1), 2, 10).
		Return(&showroom.ListMembersResponse{Members: []showroom.MemberItem{}, Total: 0, Page: 2, Limit: 10}, nil)

	req := httptest.NewRequest(http.MethodGet, "/showroom/1/member?page=2&limit=10", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_ListMembers_LimitCappedAtMax(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	// Limit 999 is capped at 100
	mockSvc.On("ListMembers", mock.Anything, mock.Anything, uint64(1), 1, 100).
		Return(&showroom.ListMembersResponse{Members: []showroom.MemberItem{}, Total: 0, Page: 1, Limit: 100}, nil)

	req := httptest.NewRequest(http.MethodGet, "/showroom/1/member?limit=999", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── RemoveMember ─────────────────────────────────────────────────────────────

func TestHandler_RemoveMember_InvalidShowroomID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodDelete, "/showroom/bad/member/99", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_RemoveMember_InvalidTargetUserID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodDelete, "/showroom/1/member/bad", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_RemoveMember_NoUserID(t *testing.T) {
	// userID=0 means extractUserID will not set the context key
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 0, ownerRoles(1))

	req := httptest.NewRequest(http.MethodDelete, "/showroom/1/member/99", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_RemoveMember_MissingShowroomRoles(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, nil)

	req := httptest.NewRequest(http.MethodDelete, "/showroom/1/member/99", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_RemoveMember_ServiceError(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("RemoveMember", mock.Anything, uint64(1), mock.Anything, uint64(1), uint64(99)).
		Return(errors.New("service error"))

	req := httptest.NewRequest(http.MethodDelete, "/showroom/1/member/99", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_RemoveMember_Success(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("RemoveMember", mock.Anything, uint64(1), mock.Anything, uint64(1), uint64(99)).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/showroom/1/member/99", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── UpdateMemberRole ─────────────────────────────────────────────────────────

func TestHandler_UpdateMemberRole_InvalidShowroomID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodPatch, "/showroom/bad/member/99", jsonBody(map[string]any{"role": "manager"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_UpdateMemberRole_InvalidTargetUserID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1/member/bad", jsonBody(map[string]any{"role": "manager"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_UpdateMemberRole_NoUserID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 0, ownerRoles(1))

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1/member/99", jsonBody(map[string]any{"role": "manager"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_UpdateMemberRole_MissingShowroomRoles(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, nil)

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1/member/99", jsonBody(map[string]any{"role": "manager"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_UpdateMemberRole_InvalidBody(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1/member/99", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_UpdateMemberRole_ServiceError(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("UpdateMemberRole", mock.Anything, uint64(1), mock.Anything, uint64(1), uint64(99), mock.Anything).
		Return(nil, errors.New("service error"))

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1/member/99", jsonBody(map[string]any{"role": "manager"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_UpdateMemberRole_Success(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("UpdateMemberRole", mock.Anything, uint64(1), mock.Anything, uint64(1), uint64(99), mock.Anything).
		Return(&showroom.AddMemberResponse{ShowroomID: 1, UserID: 99, Role: "manager"}, nil)

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1/member/99", jsonBody(map[string]any{"role": "manager"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

// ─── UpdateShowroom ───────────────────────────────────────────────────────────

func multipartBody(fields map[string]string) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

func TestHandler_UpdateShowroom_NoUserID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 0, ownerRoles(1))

	body, ct := multipartBody(map[string]string{"name": "X"})
	req := httptest.NewRequest(http.MethodPatch, "/showroom/1", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_UpdateShowroom_BadShowroomID(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	body, ct := multipartBody(map[string]string{"name": "X"})
	req := httptest.NewRequest(http.MethodPatch, "/showroom/bad", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_UpdateShowroom_MissingShowroomRoles(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, nil)

	body, ct := multipartBody(map[string]string{"name": "X"})
	req := httptest.NewRequest(http.MethodPatch, "/showroom/1", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_UpdateShowroom_ParseMultipartError(t *testing.T) {
	engine := setupShowroomEngine(showroom.NewHandler(new(mockShowroomService)), 1, ownerRoles(1))

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1", bytes.NewBufferString("not-multipart"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_UpdateShowroom_ServiceError(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("UpdateShowroom", mock.Anything, uint64(1), mock.Anything, uint64(1), mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("service error"))

	body, ct := multipartBody(map[string]string{"name": "New"})
	req := httptest.NewRequest(http.MethodPatch, "/showroom/1", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_UpdateShowroom_Success(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("UpdateShowroom", mock.Anything, uint64(1), mock.Anything, uint64(1), mock.Anything, mock.Anything, mock.Anything).
		Return(&showroom.CreateShowroomResponse{ID: 1, Name: "New Name"}, nil)

	body, ct := multipartBody(map[string]string{"name": "New Name"})
	req := httptest.NewRequest(http.MethodPatch, "/showroom/1", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_UpdateShowroom_WithFiles_Success(t *testing.T) {
	mockSvc := new(mockShowroomService)
	engine := setupShowroomEngine(showroom.NewHandler(mockSvc), 1, ownerRoles(1))

	mockSvc.On("UpdateShowroom", mock.Anything, uint64(1), mock.Anything, uint64(1), mock.Anything, mock.Anything, mock.Anything).
		Return(&showroom.CreateShowroomResponse{ID: 1, Name: "X"}, nil)

	// build a multipart body with both file fields to exercise the file-branch in handler
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("name", "X")
	logoPart, _ := mw.CreateFormFile("showroom_logo", "logo.jpg")
	_, _ = logoPart.Write([]byte("img"))
	bannerPart, _ := mw.CreateFormFile("showroom_banner", "banner.jpg")
	_, _ = bannerPart.Write([]byte("img"))
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPatch, "/showroom/1", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}
