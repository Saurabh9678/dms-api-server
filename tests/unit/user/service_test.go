package user_test

import (
	"context"
	"errors"
	"testing"

	"infiour.local/dms-api-server/internal/modules/user"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

type fakeServiceRepo struct {
	updatedUserID    uint64
	updatedName      string
	err              error
	callCount        int
}

func (f *fakeServiceRepo) UpdateName(ctx context.Context, userID uint64, name string) error {
	f.callCount++
	f.updatedUserID = userID
	f.updatedName = name
	return f.err
}

func TestUpdateProfileValidName(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	resp, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "John Doe",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
	if resp.Name != "John Doe" {
		t.Fatalf("expected name to be John Doe, got %s", resp.Name)
	}
	if repo.callCount != 1 {
		t.Fatalf("expected repo.UpdateName to be called once, got %d calls", repo.callCount)
	}
	if repo.updatedUserID != 42 {
		t.Fatalf("expected userID 42, got %d", repo.updatedUserID)
	}
	if repo.updatedName != "John Doe" {
		t.Fatalf("expected name John Doe, got %s", repo.updatedName)
	}
}

func TestUpdateProfileEmptyName(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	_, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "",
	})

	if err == nil {
		t.Fatalf("expected error for empty name")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperrors.CodeInvalidRequest {
		t.Fatalf("expected INVALID_REQUEST code, got %s", appErr.Code)
	}
	if repo.callCount != 0 {
		t.Fatalf("expected repo.UpdateName not to be called, got %d calls", repo.callCount)
	}
}

func TestUpdateProfileWhitespaceOnlyName(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	_, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "   ",
	})

	if err == nil {
		t.Fatalf("expected error for whitespace-only name")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperrors.CodeInvalidRequest {
		t.Fatalf("expected INVALID_REQUEST code, got %s", appErr.Code)
	}
	if repo.callCount != 0 {
		t.Fatalf("expected repo.UpdateName not to be called, got %d calls", repo.callCount)
	}
}

func TestUpdateProfileInvalidNameCharacters(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	_, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "John@Doe",
	})

	if err == nil {
		t.Fatalf("expected error for invalid characters")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperrors.CodeInvalidRequest {
		t.Fatalf("expected INVALID_REQUEST code, got %s", appErr.Code)
	}
	if repo.callCount != 0 {
		t.Fatalf("expected repo.UpdateName not to be called, got %d calls", repo.callCount)
	}
}

func TestUpdateProfileValidNameWithApostrophe(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	resp, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "John O'Brien",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
	if resp.Name != "John O'Brien" {
		t.Fatalf("expected name to be John O'Brien, got %s", resp.Name)
	}
	if repo.callCount != 1 {
		t.Fatalf("expected repo.UpdateName to be called once, got %d calls", repo.callCount)
	}
}

func TestUpdateProfileValidNameWithHyphen(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	resp, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "Mary-Jane",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
	if resp.Name != "Mary-Jane" {
		t.Fatalf("expected name to be Mary-Jane, got %s", resp.Name)
	}
	if repo.callCount != 1 {
		t.Fatalf("expected repo.UpdateName to be called once, got %d calls", repo.callCount)
	}
}

func TestUpdateProfileRepoError(t *testing.T) {
	repo := &fakeServiceRepo{
		err: user.ErrUserNotFound,
	}
	service := user.NewService(repo)

	_, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "John Doe",
	})

	if err == nil {
		t.Fatalf("expected error from repo")
	}
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if repo.callCount != 1 {
		t.Fatalf("expected repo.UpdateName to be called once, got %d calls", repo.callCount)
	}
}

func TestUpdateProfileNameTrimming(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	resp, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "  John Doe  ",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name != "John Doe" {
		t.Fatalf("expected trimmed name John Doe, got %s", resp.Name)
	}
	if repo.updatedName != "John Doe" {
		t.Fatalf("expected repo to receive trimmed name John Doe, got %s", repo.updatedName)
	}
}

func TestUpdateProfileValidNameWithNumbers(t *testing.T) {
	repo := &fakeServiceRepo{}
	service := user.NewService(repo)

	_, err := service.UpdateProfile(context.Background(), uint64(42), user.UpdateProfileRequest{
		Name: "John123",
	})

	if err == nil {
		t.Fatalf("expected error for name with numbers")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != apperrors.CodeInvalidRequest {
		t.Fatalf("expected INVALID_REQUEST code, got %s", appErr.Code)
	}
}
