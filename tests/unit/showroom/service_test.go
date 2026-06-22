package showroom_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"infiour.local/dms-api-server/internal/modules/showroom"
)

// mockShowroomRepo satisfies showroom's internal showroomRepo interface via its exported methods.
type mockShowroomRepo struct {
	mock.Mock
}

func (m *mockShowroomRepo) CreateWithOwner(ctx context.Context, userID uint64, s *showroom.Showroom) (*showroom.Showroom, error) {
	args := m.Called(ctx, userID, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.Showroom), args.Error(1)
}

func (m *mockShowroomRepo) UpdateFilePaths(ctx context.Context, showroomID uint64, logoPath, bannerPath *string) error {
	args := m.Called(ctx, showroomID, logoPath, bannerPath)
	return args.Error(0)
}

// mockStorageProvider satisfies storage.Provider.
type mockStorageProvider struct {
	mock.Mock
}

func (m *mockStorageProvider) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, key, data, contentType)
	return args.String(0), args.Error(1)
}

// makeFileHeader builds a minimal multipart.FileHeader for tests.
func makeFileHeader(filename string, size int64) *multipart.FileHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "image/jpeg")
	return &multipart.FileHeader{
		Filename: filename,
		Header:   h,
		Size:     size,
	}
}

// inMemoryOpener returns a WithFileOpener that serves fixed content.
func inMemoryOpener(content []byte) func(*multipart.FileHeader) (io.ReadCloser, error) {
	return func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(content)), nil
	}
}

// errorOpener always returns an error on Open.
func errorOpener(err error) func(*multipart.FileHeader) (io.ReadCloser, error) {
	return func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return nil, err
	}
}

func TestCreateShowroom_EmptyName(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "  "}, nil, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_InvalidGeolocationJSON(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{
		Name:        "Showroom",
		Geolocation: "not-json",
	}, nil, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_LogoTooLarge(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	logo := makeFileHeader("logo.jpg", 11*1024*1024)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, logo, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_BannerInvalidExtension(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	banner := makeFileHeader("banner.gif", 100)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, nil, banner)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_CreateWithOwnerError(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).
		Return(nil, errors.New("db error"))

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, nil, nil)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCreateShowroom_NoFiles_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	created := &showroom.Showroom{ID: 7, Name: "MyShowroom"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "MyShowroom"}, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(7), resp.ID)
	assert.Equal(t, "MyShowroom", resp.Name)
	assert.Nil(t, resp.ShowroomLogo)
	assert.Nil(t, resp.ShowroomBanner)
	repo.AssertExpectations(t)
	storage.AssertNotCalled(t, "Upload")
}

func TestCreateShowroom_WithValidGeolocation_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	created := &showroom.Showroom{ID: 8, Name: "Geo"}
	repo.On("CreateWithOwner", mock.Anything, uint64(2), mock.Anything).Return(created, nil)

	geo := `{"address":"123 Main St","city":"Bengaluru"}`
	resp, err := svc.CreateShowroom(context.Background(), 2, &showroom.CreateShowroomRequest{
		Name:        "Geo",
		Geolocation: geo,
	}, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	repo.AssertExpectations(t)
}

func TestCreateShowroom_LogoOpenError_SkipsUpload(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(errorOpener(errors.New("open failed"))))

	created := &showroom.Showroom{ID: 3, Name: "X"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	logo := makeFileHeader("logo.jpg", 100)
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, logo, nil)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomLogo)
	storage.AssertNotCalled(t, "Upload")
}

func TestCreateShowroom_LogoStorageError_SkipsUpload(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("fake-image-data")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 4, Name: "X"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New("upload failed"))

	logo := makeFileHeader("logo.jpg", int64(len(imgData)))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, logo, nil)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomLogo)
}

func TestCreateShowroom_BothFiles_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("image-bytes")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 5, Name: "Both"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	logoPath := "1/5/logo.jpg"
	bannerPath := "1/5/banner.png"
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(logoPath, nil).Once()
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(bannerPath, nil).Once()
	repo.On("UpdateFilePaths", mock.Anything, uint64(5), mock.Anything, mock.Anything).Return(nil)

	logo := makeFileHeader("logo.jpg", int64(len(imgData)))
	banner := makeFileHeader("banner.png", int64(len(imgData)))

	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "Both"}, logo, banner)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ShowroomLogo)
	assert.NotNil(t, resp.ShowroomBanner)
	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestCreateShowroom_BannerStorageError_SkipsUpload(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("image-bytes")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 6, Name: "X"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New("upload failed"))

	banner := makeFileHeader("banner.jpg", int64(len(imgData)))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, nil, banner)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomBanner)
}

func TestCreateShowroom_LogoOnlyUploadSuccess_UpdateFilePaths(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("img")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 9, Name: "Logo"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("1/9/logo.jpg", nil)
	repo.On("UpdateFilePaths", mock.Anything, uint64(9), mock.Anything, (*string)(nil)).Return(nil)

	logo := makeFileHeader("logo.jpg", int64(len(imgData)))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "Logo"}, logo, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp.ShowroomLogo)
	assert.Nil(t, resp.ShowroomBanner)
	repo.AssertExpectations(t)
}

// parsedFileHeader creates a real multipart.FileHeader by parsing a multipart form,
// exercising the default h.Open() code path in NewService.
func parsedFileHeader(t *testing.T, fieldName, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(fieldName, filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, "/", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(10<<20))
	return req.MultipartForm.File[fieldName][0]
}

func TestCreateShowroom_DefaultOpenFile_UploadSuccess(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	// No WithFileOpener — exercises the default h.Open() code path in NewService
	svc := showroom.NewService(repo, storage)

	created := &showroom.Showroom{ID: 20, Name: "Default"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("1/20/logo.jpg", nil)
	repo.On("UpdateFilePaths", mock.Anything, uint64(20), mock.Anything, (*string)(nil)).Return(nil)

	logo := parsedFileHeader(t, "showroom_logo", "logo.jpg", []byte("img-data"))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "Default"}, logo, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ShowroomLogo)
	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestCreateShowroom_EmptyContentType_UsesOctetStream(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("img")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 10, Name: "CT"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "application/octet-stream").
		Return("1/10/logo.jpg", nil)
	repo.On("UpdateFilePaths", mock.Anything, uint64(10), mock.Anything, (*string)(nil)).Return(nil)

	// header with no Content-Type
	fh := &multipart.FileHeader{
		Filename: "logo.jpg",
		Header:   make(textproto.MIMEHeader),
		Size:     int64(len(imgData)),
	}
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "CT"}, fh, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	storage.AssertExpectations(t)
}
