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
}

type profileRepo interface {
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
