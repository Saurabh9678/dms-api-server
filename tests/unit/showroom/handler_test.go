package showroom_test

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"infiour.local/dms-api-server/internal/modules/showroom"
	"infiour.local/dms-api-server/pkg/middleware"
)

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

func TestHandler_CreateShowroom_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	engine.POST("/showroom", handler.CreateShowroom)

	req := httptest.NewRequest(http.MethodPost, "/showroom", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertNotCalled(t, "CreateShowroom")
}

func TestHandler_CreateShowroom_BadUserIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
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
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
	engine.POST("/showroom", func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		handler.CreateShowroom(c)
	})

	// A plain GET body with wrong content-type triggers multipart parse failure
	req := httptest.NewRequest(http.MethodPost, "/showroom", bytes.NewBufferString("not-multipart"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateShowroom_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
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
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
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
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockShowroomService)
	handler := showroom.NewHandler(mockSvc)

	engine := gin.New()
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
