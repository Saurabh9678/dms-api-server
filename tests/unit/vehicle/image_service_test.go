package vehicle_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"infiour.local/dms-api-server/internal/modules/vehicle"
)

type mockImageStorage struct {
	mock.Mock
}

func (m *mockImageStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, key, data, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockImageStorage) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	args := m.Called(ctx, key, ttl)
	return args.String(0), args.Error(1)
}

func imageFileHeader(filename string, size int64, contentType string) *multipart.FileHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="photo"; filename="`+filename+`"`)
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return &multipart.FileHeader{Filename: filename, Header: h, Size: size}
}

func inMemoryImageOpener(content []byte) func(*multipart.FileHeader) (io.ReadCloser, error) {
	return func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(content)), nil
	}
}

type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) { return 0, errors.New("read fail") }
func (errReadCloser) Close() error             { return nil }

func realPhotoHeader(t *testing.T, filename string, data []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("photo", filename)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(1 << 20)
	require.NoError(t, err)
	t.Cleanup(func() { _ = form.RemoveAll() })
	files := form.File["photo"]
	require.NotEmpty(t, files)
	return files[0]
}

func TestAddVehicleImage_DefaultFileOpener(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	img := []byte("img-bytes")
	svc := vehicle.NewService(repo, storage)
	photo := realPhotoHeader(t, "front.jpg", img)

	repo.On("GetCurrentStatus", mock.Anything, uint64(18)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, img, mock.Anything).Return("7/vehicle/18/key.jpg", nil)
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(&vehicle.VehicleImage{
		ID: 9, VehicleID: 18, ImageURL: "7/vehicle/18/key.jpg", Label: vehicle.VehicleImageLabelFront, UploadedAt: time.Now(),
	}, nil)
	storage.On("SignedURL", mock.Anything, "7/vehicle/18/key.jpg", time.Hour).Return("https://signed", nil)

	resp, err := svc.AddVehicleImage(context.Background(), 7, 18, "front", photo)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(9), resp.ID)
}

func TestAddVehicleImage_Success(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	img := []byte("img-bytes")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(img)), vehicle.WithSignedURLTTL(30*time.Minute))

	photo := imageFileHeader("front.jpg", int64(len(img)), "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(18)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "7/vehicle/18/") && strings.HasSuffix(key, ".jpg")
	}), img, "image/jpeg").Return("7/vehicle/18/20260101120000.jpg", nil)
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(&vehicle.VehicleImage{
		ID:         9,
		VehicleID:  18,
		ImageURL:   "7/vehicle/18/20260101120000.jpg",
		Label:      vehicle.VehicleImageLabelFront,
		UploadedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UploadedBy: 7,
	}, nil)
	storage.On("SignedURL", mock.Anything, "7/vehicle/18/20260101120000.jpg", 30*time.Minute).
		Return("https://signed/front", nil)

	resp, err := svc.AddVehicleImage(context.Background(), 7, 18, " front ", photo)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(9), resp.ID)
	assert.Equal(t, uint64(18), resp.VehicleID)
	assert.Equal(t, "front", resp.Label)
	assert.Equal(t, "https://signed/front", resp.URL)
	storage.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestAddVehicleImage_AllLabels(t *testing.T) {
	labels := []string{"front", "interior", "exterior", "back", "wheel"}
	for _, label := range labels {
		t.Run(label, func(t *testing.T) {
			repo := new(mockVehicleRepo)
			storage := new(mockImageStorage)
			img := []byte("x")
			svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(img)))
			photo := imageFileHeader(label+".png", 1, "image/png")
			repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeReadyForSale, nil)
			storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/png").Return("1/vehicle/1/x.png", nil)
			repo.On("CreateImage", mock.Anything, mock.Anything).Return(&vehicle.VehicleImage{
				ID: 1, VehicleID: 1, ImageURL: "1/vehicle/1/x.png", Label: vehicle.VehicleImageLabel(label), UploadedAt: time.Now(),
			}, nil)
			storage.On("SignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed", nil)
			resp, err := svc.AddVehicleImage(context.Background(), 1, 1, label, photo)
			require.NoError(t, err)
			assert.Equal(t, label, resp.Label)
		})
	}
}

func TestAddVehicleImage_InvalidLabel(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "side", imageFileHeader("a.jpg", 1, "image/jpeg"))
	assert.Error(t, err)
}

func TestAddVehicleImage_EmptyLabel(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "  ", imageFileHeader("a.jpg", 1, "image/jpeg"))
	assert.Error(t, err)
}

func TestAddVehicleImage_NilPhoto(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", nil)
	assert.Error(t, err)
}

func TestAddVehicleImage_TooLarge(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	photo := imageFileHeader("a.jpg", 15*1024*1024+1, "image/jpeg")
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.Error(t, err)
}

func TestAddVehicleImage_InvalidExtension(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	photo := imageFileHeader("a.gif", 100, "image/gif")
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.Error(t, err)
}

func TestAddVehicleImage_StatusError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	photo := imageFileHeader("a.jpg", 10, "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusType(""), errors.New("db"))
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.Error(t, err)
}

func TestAddVehicleImage_Sold(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	photo := imageFileHeader("a.jpg", 10, "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeSold, nil)
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.ErrorIs(t, err, vehicle.ErrVehicleSold)
}

func TestAddVehicleImage_OpenError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := vehicle.NewService(repo, passthroughStorage{}, vehicle.WithFileOpener(func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return nil, errors.New("open fail")
	}))
	photo := imageFileHeader("a.jpg", 10, "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.Error(t, err)
}

func TestAddVehicleImage_ReadError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := vehicle.NewService(repo, passthroughStorage{}, vehicle.WithFileOpener(func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return errReadCloser{}, nil
	}))
	photo := imageFileHeader("a.jpg", 10, "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.Error(t, err)
}

func TestAddVehicleImage_EmptyContentType_UsesOctetStream(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	img := []byte("img")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(img)))
	photo := imageFileHeader("a.jpg", int64(len(img)), "")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "application/octet-stream").Return("1/vehicle/1/a.jpg", nil)
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(&vehicle.VehicleImage{
		ID: 1, VehicleID: 1, ImageURL: "1/vehicle/1/a.jpg", Label: vehicle.VehicleImageLabelFront, UploadedAt: time.Now(),
	}, nil)
	storage.On("SignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed", nil)
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	require.NoError(t, err)
	storage.AssertExpectations(t)
}

func TestAddVehicleImage_UploadError(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	img := []byte("img")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(img)))
	photo := imageFileHeader("a.jpg", int64(len(img)), "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("gcs down"))
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.Error(t, err)
}

func TestAddVehicleImage_CreateError(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	img := []byte("img")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(img)))
	photo := imageFileHeader("a.jpg", int64(len(img)), "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("key.jpg", nil)
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(nil, errors.New("insert fail"))
	_, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	assert.Error(t, err)
}

func TestAddVehicleImage_SignedURLFailure_ReturnsEmptyURL(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	img := []byte("img")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(img)))
	photo := imageFileHeader("a.jpg", int64(len(img)), "image/jpeg")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("key.jpg", nil)
	repo.On("CreateImage", mock.Anything, mock.Anything).Return(&vehicle.VehicleImage{
		ID: 1, VehicleID: 1, ImageURL: "key.jpg", Label: vehicle.VehicleImageLabelFront, UploadedAt: time.Now(),
	}, nil)
	storage.On("SignedURL", mock.Anything, "key.jpg", mock.Anything).Return("", errors.New("sign fail"))
	resp, err := svc.AddVehicleImage(context.Background(), 1, 1, "front", photo)
	require.NoError(t, err)
	assert.Equal(t, "", resp.URL)
}

func TestAddVehicleImage_ZeroTTLOptionIgnored(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage, vehicle.WithSignedURLTTL(0))
	details := &vehicle.VehicleFullDetails{
		Vehicle: vehicle.Vehicle{ID: 1},
		Images:  []vehicle.VehicleImage{{ID: 1, ImageURL: "k.jpg"}},
	}
	repo.On("GetByIDWithFullDetails", mock.Anything, uint64(1)).Return(details, nil)
	storage.On("SignedURL", mock.Anything, "k.jpg", time.Hour).Return("https://signed", nil)
	result, err := svc.GetVehicleByID(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, result.Images, 1)
	assert.Equal(t, "https://signed", result.Images[0].ImageURL)
}

func TestGetVehicleByID_SignsImagesAndOmitsFailures(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage)
	details := &vehicle.VehicleFullDetails{
		Vehicle: vehicle.Vehicle{ID: 1},
		Images: []vehicle.VehicleImage{
			{ID: 1, ImageURL: "ok.jpg"},
			{ID: 2, ImageURL: ""},
			{ID: 3, ImageURL: "bad.jpg"},
		},
	}
	repo.On("GetByIDWithFullDetails", mock.Anything, uint64(1)).Return(details, nil)
	storage.On("SignedURL", mock.Anything, "ok.jpg", time.Hour).Return("https://signed/ok", nil)
	storage.On("SignedURL", mock.Anything, "bad.jpg", time.Hour).Return("", errors.New("sign fail"))

	result, err := svc.GetVehicleByID(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, result.Images, 1)
	assert.Equal(t, uint64(1), result.Images[0].ID)
	assert.Equal(t, "https://signed/ok", result.Images[0].ImageURL)
}

func TestGetVehicleByID_NoImages(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage)
	details := &vehicle.VehicleFullDetails{Vehicle: vehicle.Vehicle{ID: 1}}
	repo.On("GetByIDWithFullDetails", mock.Anything, uint64(1)).Return(details, nil)
	result, err := svc.GetVehicleByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, result.Images)
	storage.AssertNotCalled(t, "SignedURL")
}

func TestGetVehicleByID_NilDetails(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage)
	repo.On("GetByIDWithFullDetails", mock.Anything, uint64(1)).Return(nil, nil)
	result, err := svc.GetVehicleByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, result)
	storage.AssertNotCalled(t, "SignedURL")
}

func TestDeleteVehicleImage_Success(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	repo.On("SoftDeleteImage", mock.Anything, uint64(1), uint64(9)).Return(nil)
	err := svc.DeleteVehicleImage(context.Background(), 1, 9)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteVehicleImage_StatusError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusType(""), errors.New("db"))
	err := svc.DeleteVehicleImage(context.Background(), 1, 9)
	assert.Error(t, err)
}

func TestDeleteVehicleImage_Sold(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeSold, nil)
	err := svc.DeleteVehicleImage(context.Background(), 1, 9)
	assert.ErrorIs(t, err, vehicle.ErrVehicleSold)
}

func TestDeleteVehicleImage_NotFound(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	repo.On("SoftDeleteImage", mock.Anything, uint64(1), uint64(9)).Return(vehicle.ErrVehicleImageNotFound)
	err := svc.DeleteVehicleImage(context.Background(), 1, 9)
	assert.ErrorIs(t, err, vehicle.ErrVehicleImageNotFound)
}

func TestListVehicles_WithSignedImages(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage)

	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}
	vehicles := []vehicle.VehicleWithDetails{{ID: 10, VehicleType: vehicle.VehicleTypeCar}}
	repo.On("CountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]vehicle.CategoryMetrics{vehicle.VehicleTypeCar: {Total: 1}}, nil)
	repo.On("List", mock.Anything, mock.Anything).Return(vehicles, nil)
	repo.On("ListImagesByVehicleIDs", mock.Anything, []uint64{10}).Return(map[uint64][]vehicle.VehicleImage{
		10: {
			{ID: 1, VehicleID: 10, ImageURL: "front.jpg", Label: vehicle.VehicleImageLabelFront},
			{ID: 2, VehicleID: 10, ImageURL: "int.jpg", Label: vehicle.VehicleImageLabelInterior},
			{ID: 3, VehicleID: 10, ImageURL: "", Label: vehicle.VehicleImageLabelBack},
			{ID: 4, VehicleID: 10, ImageURL: "bad.jpg", Label: vehicle.VehicleImageLabelWheel},
			{ID: 5, VehicleID: 10, ImageURL: "orphan.jpg", Label: ""},
		},
	}, nil)
	storage.On("SignedURL", mock.Anything, "front.jpg", time.Hour).Return("https://signed/front", nil)
	storage.On("SignedURL", mock.Anything, "int.jpg", time.Hour).Return("https://signed/int", nil)
	storage.On("SignedURL", mock.Anything, "bad.jpg", time.Hour).Return("", errors.New("sign fail"))

	resp, err := svc.ListVehicles(context.Background(), query)
	require.NoError(t, err)
	require.NotNil(t, resp.Cars)
	require.Len(t, resp.Cars.Vehicles, 1)
	images := resp.Cars.Vehicles[0].Images
	require.NotNil(t, images)
	assert.Equal(t, []vehicle.VehicleImageItem{{ID: 1, URL: "https://signed/front"}}, images["front"])
	assert.Equal(t, []vehicle.VehicleImageItem{{ID: 2, URL: "https://signed/int"}}, images["interior"])
	_, hasWheel := images["wheel"]
	assert.False(t, hasWheel)
	_, hasBack := images["back"]
	assert.False(t, hasBack)
}

func TestListVehicles_ListImagesError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	query := &vehicle.ListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20}
	repo.On("CountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]vehicle.CategoryMetrics{vehicle.VehicleTypeCar: {Total: 1}}, nil)
	repo.On("List", mock.Anything, mock.Anything).Return([]vehicle.VehicleWithDetails{{ID: 1, VehicleType: vehicle.VehicleTypeCar}}, nil)
	repo.On("ListImagesByVehicleIDs", mock.Anything, []uint64{1}).Return(nil, errors.New("db fail"))

	resp, err := svc.ListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPublicListVehicles_WithSignedImages(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage)

	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}
	vehicles := []vehicle.VehicleWithDetails{{ID: 7, VehicleType: vehicle.VehicleTypeBike}}
	repo.On("PublicCountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]int64{vehicle.VehicleTypeBike: 1}, nil)
	repo.On("PublicList", mock.Anything, mock.Anything).Return(vehicles, nil)
	repo.On("ListImagesByVehicleIDs", mock.Anything, []uint64{7}).Return(map[uint64][]vehicle.VehicleImage{
		7: {{ID: 9, VehicleID: 7, ImageURL: "bike.jpg", Label: vehicle.VehicleImageLabelExterior}},
	}, nil)
	storage.On("SignedURL", mock.Anything, "bike.jpg", time.Hour).Return("https://signed/bike", nil)

	resp, err := svc.PublicListVehicles(context.Background(), query)
	require.NoError(t, err)
	require.NotNil(t, resp.Bikes)
	require.Len(t, resp.Bikes.Vehicles, 1)
	images := resp.Bikes.Vehicles[0].Images
	require.Equal(t, []vehicle.VehicleImageItem{{ID: 9, URL: "https://signed/bike"}}, images["exterior"])
}

func TestPublicListVehicles_ListImagesError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	query := &vehicle.PublicListVehiclesQuery{ShowroomID: 1, Page: 1, Limit: 20, SortBy: "price_asc"}
	repo.On("PublicCountByType", mock.Anything, mock.Anything).Return(map[vehicle.VehicleType]int64{}, nil)
	repo.On("PublicList", mock.Anything, mock.Anything).Return([]vehicle.VehicleWithDetails{{ID: 2, VehicleType: vehicle.VehicleTypeCar}}, nil)
	repo.On("ListImagesByVehicleIDs", mock.Anything, []uint64{2}).Return(nil, errors.New("images fail"))

	resp, err := svc.PublicListVehicles(context.Background(), query)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
