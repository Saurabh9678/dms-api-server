package showroom

import (
	"context"
	"encoding/json"
	"errors"
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
	AddMember(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, req *AddMemberRequest) (*AddMemberResponse, error)
	ListMembers(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, page, limit int) (*ListMembersResponse, error)
	RemoveMember(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64) error
	UpdateMemberRole(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64, req *UpdateMemberRoleRequest) (*AddMemberResponse, error)
}

type showroomRepo interface {
	CreateWithOwner(ctx context.Context, userID uint64, s *Showroom) (*Showroom, error)
	UpdateFilePaths(ctx context.Context, showroomID uint64, logoPath, bannerPath *string) error
	AddMember(ctx context.Context, showroomID, targetUserID uint64, roleType string) error
	ListMembers(ctx context.Context, showroomID uint64, page, limit int) ([]MemberRecord, int64, error)
	GetMemberRole(ctx context.Context, showroomID, targetUserID uint64) (string, error)
	RemoveMember(ctx context.Context, showroomID, targetUserID uint64) error
	UpdateMemberRole(ctx context.Context, showroomID, targetUserID uint64, newRoleType string) error
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

func (s *service) AddMember(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, req *AddMemberRequest) (*AddMemberResponse, error) {
	callerRole := callerRoles[showroomID]
	if callerRole != "owner" && callerRole != "manager" {
		return nil, apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
	}

	if req.Role != "manager" && req.Role != "employee" {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	// managers can only assign the employee role
	if callerRole == "manager" && req.Role != "employee" {
		return nil, apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
	}

	if err := s.repo.AddMember(ctx, showroomID, req.UserID, req.Role); err != nil {
		return nil, mapMemberRepoError(err)
	}

	return &AddMemberResponse{ShowroomID: showroomID, UserID: req.UserID, Role: req.Role}, nil
}

func (s *service) ListMembers(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, page, limit int) (*ListMembersResponse, error) {
	callerRole := callerRoles[showroomID]
	if callerRole != "owner" && callerRole != "manager" {
		return nil, apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
	}

	records, total, err := s.repo.ListMembers(ctx, showroomID, page, limit)
	if err != nil {
		return nil, err
	}

	members := make([]MemberItem, 0, len(records))
	for _, r := range records {
		item := MemberItem{UserID: r.UserID, Role: r.Role}
		if r.Name != "" {
			name := r.Name
			item.Name = &name
		}
		if r.CountryCode != "" || r.PhoneNumber != "" {
			combined := r.CountryCode + r.PhoneNumber
			item.PhoneNumber = &combined
		}
		members = append(members, item)
	}

	return &ListMembersResponse{Members: members, Total: total, Page: page, Limit: limit}, nil
}

func (s *service) RemoveMember(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64) error {
	callerRole := callerRoles[showroomID]

	// Self-removal is allowed for any member of the showroom.
	if callerUserID == targetUserID {
		if callerRole == "" {
			return apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
		}
		return mapMemberRepoError(s.repo.RemoveMember(ctx, showroomID, targetUserID))
	}

	if callerRole != "owner" && callerRole != "manager" {
		return apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
	}

	// Managers may only remove employees, not other managers or owners.
	if callerRole == "manager" {
		targetRole, err := s.repo.GetMemberRole(ctx, showroomID, targetUserID)
		if err != nil {
			return mapMemberRepoError(err)
		}
		if targetRole != "employee" {
			return apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
		}
	}

	return mapMemberRepoError(s.repo.RemoveMember(ctx, showroomID, targetUserID))
}

func (s *service) UpdateMemberRole(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64, req *UpdateMemberRoleRequest) (*AddMemberResponse, error) {
	callerRole := callerRoles[showroomID]
	if callerRole != "owner" {
		return nil, apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
	}

	if callerUserID == targetUserID {
		return nil, apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
	}

	if req.Role != "manager" && req.Role != "employee" {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if err := s.repo.UpdateMemberRole(ctx, showroomID, targetUserID, req.Role); err != nil {
		return nil, mapMemberRepoError(err)
	}

	return &AddMemberResponse{ShowroomID: showroomID, UserID: targetUserID, Role: req.Role}, nil
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

func mapMemberRepoError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrTargetUserNotFound):
		return apperrors.NewAppError(apperrors.CodeTargetUserNotFound, "invalid request", http.StatusUnprocessableEntity, nil)
	case errors.Is(err, ErrDuplicateMember):
		return apperrors.NewAppError(apperrors.CodeAlreadyAMember, "invalid request", http.StatusConflict, nil)
	case errors.Is(err, ErrMemberNotFound):
		return apperrors.NewAppError(apperrors.CodeMemberNotFound, "invalid request", http.StatusNotFound, nil)
	default:
		return err
	}
}
