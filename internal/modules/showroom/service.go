package showroom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// ServiceOption configures the service. Used in tests to inject behaviour.
type ServiceOption func(*service)

// WithFileOpener overrides how multipart file headers are opened. Used in tests.
func WithFileOpener(fn func(*multipart.FileHeader) (io.ReadCloser, error)) ServiceOption {
	return func(s *service) { s.openFile = fn }
}

type Service interface {
	CreateShowroom(ctx context.Context, userID uint64, req *CreateShowroomRequest, logo, banner *multipart.FileHeader) (*CreateShowroomResponse, error)
}

type showroomRepo interface {
	CreateWithOwner(ctx context.Context, userID uint64, s *Showroom) (*Showroom, error)
	UpdateFilePaths(ctx context.Context, showroomID uint64, logoPath, bannerPath *string) error
}

type service struct {
	repo     showroomRepo
	storage  storageprovider.Provider
	openFile func(*multipart.FileHeader) (io.ReadCloser, error)
}

func NewService(repo showroomRepo, storage storageprovider.Provider, opts ...ServiceOption) Service {
	s := &service{
		repo:    repo,
		storage: storage,
		openFile: func(h *multipart.FileHeader) (io.ReadCloser, error) {
			return h.Open()
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *service) CreateShowroom(ctx context.Context, userID uint64, req *CreateShowroomRequest, logo, banner *multipart.FileHeader) (*CreateShowroomResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	var geolocationRaw json.RawMessage
	if req.Geolocation != "" {
		if !json.Valid([]byte(req.Geolocation)) {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		geolocationRaw = json.RawMessage(req.Geolocation)
	}

	if err := validateFile(logo); err != nil {
		return nil, err
	}
	if err := validateFile(banner); err != nil {
		return nil, err
	}

	created, err := s.repo.CreateWithOwner(ctx, userID, &Showroom{
		Name:                name,
		ShowroomGeolocation: geolocationRaw,
	})
	if err != nil {
		return nil, err
	}

	logoPath := s.maybeUpload(ctx, userID, created.ID, logo)
	bannerPath := s.maybeUpload(ctx, userID, created.ID, banner)

	if logoPath != nil || bannerPath != nil {
		_ = s.repo.UpdateFilePaths(ctx, created.ID, logoPath, bannerPath)
		created.ShowroomLogo = logoPath
		created.ShowroomBanner = bannerPath
	}

	return &CreateShowroomResponse{
		ID:             created.ID,
		Name:           created.Name,
		ShowroomLogo:   created.ShowroomLogo,
		ShowroomBanner: created.ShowroomBanner,
		Geolocation:    created.ShowroomGeolocation,
	}, nil
}

func (s *service) maybeUpload(ctx context.Context, userID, showroomID uint64, header *multipart.FileHeader) *string {
	if header == nil {
		return nil
	}
	f, err := s.openFile(header)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	data, _ := io.ReadAll(f)

	ext := strings.ToLower(filepath.Ext(header.Filename))
	key := fmt.Sprintf("%d/%d/%s%s", userID, showroomID, time.Now().Format("20060102150405"), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	path, err := s.storage.Upload(ctx, key, data, contentType)
	if err != nil {
		return nil
	}
	return &path
}

func validateFile(header *multipart.FileHeader) error {
	if header == nil {
		return nil
	}
	if header.Size > maxFileSize {
		return apperrors.NewAppError(apperrors.CodeFileTooLarge, "invalid request", http.StatusBadRequest, nil)
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		return apperrors.NewAppError(apperrors.CodeInvalidFileType, "invalid request", http.StatusBadRequest, nil)
	}
	return nil
}
