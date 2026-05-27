package user

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

type Service interface {
	UpdateProfile(ctx context.Context, userID uint64, req UpdateProfileRequest) (*UpdateProfileResponse, error)
	GetProfile(ctx context.Context, userID uint64) (*GetProfileResponse, error)
}

type profileRepo interface {
	FindByID(ctx context.Context, userID uint64) (*User, error)
	FindShowroomRolesByUserID(ctx context.Context, userID uint64) ([]ShowroomRole, error)
	UpdateName(ctx context.Context, userID uint64, name string) error
}

type service struct {
	repo profileRepo
}

func NewService(repo profileRepo) Service {
	return &service{
		repo: repo,
	}
}

var namePattern = regexp.MustCompile(`^[\p{L}\s''-]+$`)

func (s *service) UpdateProfile(ctx context.Context, userID uint64, req UpdateProfileRequest) (*UpdateProfileResponse, error) {
	trimmedName := strings.TrimSpace(req.Name)
	if trimmedName == "" {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if !namePattern.MatchString(trimmedName) {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if err := s.repo.UpdateName(ctx, userID, trimmedName); err != nil {
		return nil, err
	}

	return &UpdateProfileResponse{Name: trimmedName}, nil
}

func (s *service) GetProfile(ctx context.Context, userID uint64) (*GetProfileResponse, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	showroomRoles, err := s.repo.FindShowroomRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var name *string
	if u.Name != "" {
		n := u.Name
		name = &n
	}

	var phoneNumber *string
	if u.CountryCode != "" || u.PhoneNumber != "" {
		combined := u.CountryCode + u.PhoneNumber
		phoneNumber = &combined
	}

	return &GetProfileResponse{
		Name:          name,
		PhoneNumber:   phoneNumber,
		ShowroomRoles: showroomRoles,
	}, nil
}
