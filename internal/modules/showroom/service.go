package showroom

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

const maxFileSize = 10 * 1024 * 1024 // 10 MB

const (
	showroomIDLength          = 8
	showroomIDGenerateRetries = 5
)

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

// WithSignedURLTTL overrides the signed URL lifetime used in API responses.
func WithSignedURLTTL(ttl time.Duration) ServiceOption {
	return func(s *service) {
		if ttl > 0 {
			s.signedURLTTL = ttl
		}
	}
}

type Service interface {
	CreateShowroom(ctx context.Context, userID uint64, req *CreateShowroomRequest, logo, banner *multipart.FileHeader) (*CreateShowroomResponse, error)
	ListShowrooms(ctx context.Context, userID uint64) (*ListShowroomsResponse, error)
	UpdateShowroom(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID uint64, req *UpdateShowroomRequest, logo, banner *multipart.FileHeader) (*CreateShowroomResponse, error)
	AddMember(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, req *AddMemberRequest) (*AddMemberResponse, error)
	ListMembers(ctx context.Context, callerRoles map[uint64]string, showroomID uint64, page, limit int) (*ListMembersResponse, error)
	RemoveMember(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64) error
	UpdateMemberRole(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID, targetUserID uint64, req *UpdateMemberRoleRequest) (*AddMemberResponse, error)
}

type showroomRepo interface {
	CreateWithOwner(ctx context.Context, userID uint64, s *Showroom) (*Showroom, error)
	UpdateFilePaths(ctx context.Context, showroomID uint64, logoPath, bannerPath *string) error
	ListByUserID(ctx context.Context, userID uint64) ([]ShowroomListRecord, error)
	GetByID(ctx context.Context, showroomID uint64) (*Showroom, error)
	UpdateShowroomFields(ctx context.Context, showroomID uint64, updates map[string]any) error
	AddMember(ctx context.Context, showroomID, targetUserID uint64, roleType string) error
	ListMembers(ctx context.Context, showroomID uint64, page, limit int) ([]MemberRecord, int64, error)
	GetMemberRole(ctx context.Context, showroomID, targetUserID uint64) (string, error)
	RemoveMember(ctx context.Context, showroomID, targetUserID uint64) error
	UpdateMemberRole(ctx context.Context, showroomID, targetUserID uint64, newRoleType string) error
}

type service struct {
	repo         showroomRepo
	storage      storageprovider.Provider
	signedURLTTL time.Duration
	openFile     func(*multipart.FileHeader) (io.ReadCloser, error)
}

func NewService(repo showroomRepo, storage storageprovider.Provider, opts ...ServiceOption) Service {
	s := &service{
		repo:         repo,
		storage:      storage,
		signedURLTTL: time.Hour,
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

	var created *Showroom
	var err error
	for attempt := 0; attempt < showroomIDGenerateRetries; attempt++ {
		created, err = s.repo.CreateWithOwner(ctx, userID, &Showroom{
			ShowroomID:          generateShowroomID(showroomIDLength),
			Name:                name,
			ShowroomGeolocation: geolocationRaw,
		})
		if err == nil {
			break
		}
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	logoPath := s.maybeUpload(ctx, userID, created.ShowroomID, logo)
	bannerPath := s.maybeUpload(ctx, userID, created.ShowroomID, banner)

	if logoPath != nil || bannerPath != nil {
		_ = s.repo.UpdateFilePaths(ctx, created.ID, logoPath, bannerPath)
		created.ShowroomLogo = logoPath
		created.ShowroomBanner = bannerPath
	}

	return &CreateShowroomResponse{
		ID:             created.ID,
		ShowroomID:     created.ShowroomID,
		Name:           created.Name,
		ShowroomLogo:   s.resolveAccessURL(ctx, created.ShowroomLogo),
		ShowroomBanner: s.resolveAccessURL(ctx, created.ShowroomBanner),
		Geolocation:    created.ShowroomGeolocation,
	}, nil
}

func (s *service) ListShowrooms(ctx context.Context, userID uint64) (*ListShowroomsResponse, error) {
	records, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]ShowroomListItem, 0, len(records))
	for _, r := range records {
		items = append(items, ShowroomListItem{
			ID:         r.ID,
			ShowroomID: r.ShowroomID,
			Name:       r.Name,
			Role:       r.Role,
		})
	}

	return &ListShowroomsResponse{Showrooms: items}, nil
}

func (s *service) UpdateShowroom(ctx context.Context, callerUserID uint64, callerRoles map[uint64]string, showroomID uint64, req *UpdateShowroomRequest, logo, banner *multipart.FileHeader) (*CreateShowroomResponse, error) {
	callerRole := callerRoles[showroomID]
	if callerRole != "owner" && callerRole != "manager" {
		return nil, apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)
	}

	existing, err := s.repo.GetByID(ctx, showroomID)
	if err != nil {
		if errors.Is(err, ErrShowroomNotFound) {
			return nil, apperrors.NewAppError(apperrors.CodeShowroomNotFound, "invalid request", http.StatusNotFound, nil)
		}
		return nil, err
	}

	updates := map[string]any{}

	if name := strings.TrimSpace(req.Name); name != "" {
		updates["name"] = name
		existing.Name = name
	}

	if req.Geolocation != "" {
		if !json.Valid([]byte(req.Geolocation)) {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		geoRaw := json.RawMessage(req.Geolocation)
		updates["showroom_geolocation"] = geoRaw
		existing.ShowroomGeolocation = geoRaw
	}

	if req.RemoveLogo == "true" {
		updates["showroom_logo"] = nil
		existing.ShowroomLogo = nil
	}
	if logo != nil {
		if err := validateFile(logo); err != nil {
			return nil, err
		}
		if path := s.maybeUpload(ctx, callerUserID, existing.ShowroomID, logo); path != nil {
			updates["showroom_logo"] = *path
			existing.ShowroomLogo = path
		}
	}

	if req.RemoveBanner == "true" {
		updates["showroom_banner"] = nil
		existing.ShowroomBanner = nil
	}
	if banner != nil {
		if err := validateFile(banner); err != nil {
			return nil, err
		}
		if path := s.maybeUpload(ctx, callerUserID, existing.ShowroomID, banner); path != nil {
			updates["showroom_banner"] = *path
			existing.ShowroomBanner = path
		}
	}

	if len(updates) > 0 {
		if err := s.repo.UpdateShowroomFields(ctx, showroomID, updates); err != nil {
			return nil, err
		}
	}

	return &CreateShowroomResponse{
		ID:             existing.ID,
		ShowroomID:     existing.ShowroomID,
		Name:           existing.Name,
		ShowroomLogo:   s.resolveAccessURL(ctx, existing.ShowroomLogo),
		ShowroomBanner: s.resolveAccessURL(ctx, existing.ShowroomBanner),
		Geolocation:    existing.ShowroomGeolocation,
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
		if r.CountryCode != "" {
			countryCode := r.CountryCode
			item.CountryCode = &countryCode
		}
		if r.PhoneNumber != "" {
			phoneNumber := r.PhoneNumber
			item.PhoneNumber = &phoneNumber
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

func (s *service) maybeUpload(ctx context.Context, userID uint64, externalShowroomID string, header *multipart.FileHeader) *string {
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
	key := fmt.Sprintf("%d/showroom/%s/%s%s", userID, externalShowroomID, time.Now().Format("20060102150405"), ext)

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

// resolveAccessURL turns a stored object key into a client-facing access URL (signed when using GCS).
// On signing failure the field is omitted (nil) so clients never receive a bare private key path.
func (s *service) resolveAccessURL(ctx context.Context, key *string) *string {
	if key == nil || *key == "" {
		return nil
	}
	url, err := s.storage.SignedURL(ctx, *key, s.signedURLTTL)
	if err != nil {
		return nil
	}
	return &url
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

func generateShowroomID(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	max := big.NewInt(int64(len(chars)))
	for i := 0; i < length; i++ {
		// crypto/rand panics internally on entropy failure (Go 1.20+); error is unreachable.
		value, _ := rand.Int(rand.Reader, max)
		result[i] = chars[value.Int64()]
	}
	return string(result)
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
