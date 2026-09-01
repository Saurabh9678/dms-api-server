package vehicle_test

import (
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

func documentFileHeader(filename string, size int64, contentType string) *multipart.FileHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return &multipart.FileHeader{Filename: filename, Header: h, Size: size}
}

func TestAddVehicleDocument_Success(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	data := []byte("pdf-bytes")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(data)), vehicle.WithSignedURLTTL(30*time.Minute))

	file := documentFileHeader("rc.pdf", int64(len(data)), "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(18)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "7/vehicle/18/docs/") && strings.HasSuffix(key, ".pdf")
	}), data, "application/pdf").Return("7/vehicle/18/docs/20260101120000.pdf", nil)
	repo.On("CreateDocument", mock.Anything, mock.Anything).Return(&vehicle.VehicleDocument{
		ID:           11,
		VehicleID:    18,
		DocumentType: vehicle.VehicleDocumentTypeRegistrationCertificate,
		DocumentURL:  "7/vehicle/18/docs/20260101120000.pdf",
		UploadedAt:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UploadedBy:   7,
	}, nil)
	storage.On("SignedURL", mock.Anything, "7/vehicle/18/docs/20260101120000.pdf", 30*time.Minute).
		Return("https://signed/rc", nil)

	resp, err := svc.AddVehicleDocument(context.Background(), 7, 18, "registration_certificate", file)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(11), resp.ID)
	assert.Equal(t, "registration_certificate", resp.DocumentType)
	assert.Equal(t, "https://signed/rc", resp.URL)
}

func TestAddVehicleDocument_AllDocumentTypes(t *testing.T) {
	types := []string{"registration_certificate", "insurance", "pollution"}
	for _, docType := range types {
		t.Run(docType, func(t *testing.T) {
			repo := new(mockVehicleRepo)
			storage := new(mockImageStorage)
			data := []byte("doc")
			svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(data)))
			file := documentFileHeader("doc.pdf", int64(len(data)), "application/pdf")
			repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
			storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("1/vehicle/1/docs/x.pdf", nil)
			repo.On("CreateDocument", mock.Anything, mock.Anything).Return(&vehicle.VehicleDocument{
				ID: 1, VehicleID: 1, DocumentType: vehicle.VehicleDocumentType(docType), DocumentURL: "1/vehicle/1/docs/x.pdf", UploadedAt: time.Now(),
			}, nil)
			storage.On("SignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed", nil)
			_, err := svc.AddVehicleDocument(context.Background(), 1, 1, docType, file)
			require.NoError(t, err)
		})
	}
}

func TestAddVehicleDocument_InvalidType(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	file := documentFileHeader("rc.pdf", 10, "application/pdf")
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "license", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_EmptyType(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	file := documentFileHeader("rc.pdf", 10, "application/pdf")
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "  ", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_NilFile(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", nil)
	assert.Error(t, err)
}

func TestAddVehicleDocument_TooLarge(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	file := documentFileHeader("rc.pdf", 15*1024*1024+1, "application/pdf")
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_InvalidExtension(t *testing.T) {
	svc := newTestService(new(mockVehicleRepo))
	file := documentFileHeader("rc.gif", 100, "image/gif")
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_StatusError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	file := documentFileHeader("rc.pdf", 10, "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusType(""), errors.New("db"))
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_Sold(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := newTestService(repo)
	file := documentFileHeader("rc.pdf", 10, "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeSold, nil)
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.ErrorIs(t, err, vehicle.ErrVehicleSold)
}

func TestAddVehicleDocument_OpenError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := vehicle.NewService(repo, passthroughStorage{}, vehicle.WithFileOpener(func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return nil, errors.New("open fail")
	}))
	file := documentFileHeader("rc.pdf", 10, "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_ReadError(t *testing.T) {
	repo := new(mockVehicleRepo)
	svc := vehicle.NewService(repo, passthroughStorage{}, vehicle.WithFileOpener(func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return errReadCloser{}, nil
	}))
	file := documentFileHeader("rc.pdf", 10, "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_EmptyContentType_UsesOctetStream(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	data := []byte("doc")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(data)))
	file := documentFileHeader("rc.pdf", int64(len(data)), "")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "application/octet-stream").Return("1/vehicle/1/docs/a.pdf", nil)
	repo.On("CreateDocument", mock.Anything, mock.Anything).Return(&vehicle.VehicleDocument{
		ID: 1, VehicleID: 1, DocumentType: vehicle.VehicleDocumentTypeInsurance, DocumentURL: "1/vehicle/1/docs/a.pdf", UploadedAt: time.Now(),
	}, nil)
	storage.On("SignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed", nil)
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	require.NoError(t, err)
	storage.AssertExpectations(t)
}

func TestAddVehicleDocument_UploadError(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	data := []byte("doc")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(data)))
	file := documentFileHeader("rc.pdf", int64(len(data)), "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("gcs down"))
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_CreateError(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	data := []byte("doc")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(data)))
	file := documentFileHeader("rc.pdf", int64(len(data)), "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("key.pdf", nil)
	repo.On("CreateDocument", mock.Anything, mock.Anything).Return(nil, errors.New("insert fail"))
	_, err := svc.AddVehicleDocument(context.Background(), 1, 1, "insurance", file)
	assert.Error(t, err)
}

func TestAddVehicleDocument_SignedURLFailure_ReturnsEmptyURL(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	data := []byte("doc")
	svc := vehicle.NewService(repo, storage, vehicle.WithFileOpener(inMemoryImageOpener(data)))
	file := documentFileHeader("rc.pdf", int64(len(data)), "application/pdf")
	repo.On("GetCurrentStatus", mock.Anything, uint64(1)).Return(vehicle.VehicleStatusTypeGarage, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("key.pdf", nil)
	repo.On("CreateDocument", mock.Anything, mock.Anything).Return(&vehicle.VehicleDocument{
		ID: 1, VehicleID: 1, DocumentType: vehicle.VehicleDocumentTypePollution, DocumentURL: "key.pdf", UploadedAt: time.Now(),
	}, nil)
	storage.On("SignedURL", mock.Anything, "key.pdf", mock.Anything).Return("", errors.New("sign fail"))
	resp, err := svc.AddVehicleDocument(context.Background(), 1, 1, "pollution", file)
	require.NoError(t, err)
	assert.Equal(t, "", resp.URL)
}

func TestGetVehicleByID_SignsDocumentsAndOmitsFailures(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage)
	details := &vehicle.VehicleFullDetails{
		Vehicle: vehicle.Vehicle{ID: 1},
		Documents: []vehicle.VehicleDocument{
			{ID: 1, DocumentType: vehicle.VehicleDocumentTypeInsurance, DocumentURL: "ok.pdf"},
			{ID: 2, DocumentType: vehicle.VehicleDocumentTypeInsurance, DocumentURL: ""},
			{ID: 3, DocumentType: vehicle.VehicleDocumentTypePollution, DocumentURL: "bad.pdf"},
		},
	}
	repo.On("GetByIDWithFullDetails", mock.Anything, uint64(1)).Return(details, nil)
	storage.On("SignedURL", mock.Anything, "ok.pdf", time.Hour).Return("https://signed/ok", nil)
	storage.On("SignedURL", mock.Anything, "bad.pdf", time.Hour).Return("", errors.New("sign fail"))

	result, err := svc.GetVehicleByID(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, result.Documents, 1)
	assert.Equal(t, uint64(1), result.Documents[0].ID)
	assert.Equal(t, "https://signed/ok", result.Documents[0].DocumentURL)
}

func TestGetVehicleByID_NoDocuments(t *testing.T) {
	repo := new(mockVehicleRepo)
	storage := new(mockImageStorage)
	svc := vehicle.NewService(repo, storage)
	details := &vehicle.VehicleFullDetails{Vehicle: vehicle.Vehicle{ID: 1}}
	repo.On("GetByIDWithFullDetails", mock.Anything, uint64(1)).Return(details, nil)
	result, err := svc.GetVehicleByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, result.Documents)
	storage.AssertNotCalled(t, "SignedURL")
}
