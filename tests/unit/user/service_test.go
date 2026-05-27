package user_test

import (
	"context"
	"errors"
	"testing"

	"infiour.local/dms-api-server/internal/modules/user"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

type fakeServiceRepo struct {
	updatedUserID uint64
	updatedName   string
	err           error
	callCount     int
	user          *user.User
	showroomRoles []user.ShowroomRole
	findErr       error
	rolesErr      error
}

func (f *fakeServiceRepo) FindByID(_ context.Context, _ uint64) (*user.User, error) {
	return f.user, f.findErr
}

func (f *fakeServiceRepo) FindShowroomRolesByUserID(_ context.Context, _ uint64) ([]user.ShowroomRole, error) {
	return f.showroomRoles, f.rolesErr
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

func TestUpdateProfileServiceEmptyName(t *testing.T) {
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

func TestGetProfileSuccess(t *testing.T) {
	repo := &fakeServiceRepo{
		user: &user.User{ID: 1, Name: "Alice", CountryCode: "+91", PhoneNumber: "9999999999"},
		showroomRoles: []user.ShowroomRole{
			{ShowroomID: 10, ShowroomName: "Showroom A", Role: user.UserRoleTypeOwner},
		},
	}
	svc := user.NewService(repo)

	resp, err := svc.GetProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name == nil || *resp.Name != "Alice" {
		t.Fatalf("expected name Alice, got %v", resp.Name)
	}
	if resp.PhoneNumber == nil || *resp.PhoneNumber != "+919999999999" {
		t.Fatalf("expected phone +919999999999, got %v", resp.PhoneNumber)
	}
	if len(resp.ShowroomRoles) != 1 || resp.ShowroomRoles[0].ShowroomID != 10 {
		t.Fatalf("unexpected showroom roles: %+v", resp.ShowroomRoles)
	}
}

func TestGetProfileNullName(t *testing.T) {
	repo := &fakeServiceRepo{
		user:          &user.User{ID: 2, Name: "", CountryCode: "+91", PhoneNumber: "8888888888"},
		showroomRoles: []user.ShowroomRole{},
	}
	svc := user.NewService(repo)

	resp, err := svc.GetProfile(context.Background(), 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name != nil {
		t.Fatalf("expected nil name, got %v", resp.Name)
	}
	if resp.ShowroomRoles == nil {
		t.Fatalf("expected empty slice, got nil")
	}
}

func TestGetProfileUserNotFound(t *testing.T) {
	repo := &fakeServiceRepo{findErr: user.ErrUserNotFound}
	svc := user.NewService(repo)

	_, err := svc.GetProfile(context.Background(), 99)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetProfileRolesRepoError(t *testing.T) {
	repo := &fakeServiceRepo{
		user:     &user.User{ID: 1, Name: "Alice", CountryCode: "+91", PhoneNumber: "9999999999"},
		rolesErr: errors.New("db error"),
	}
	svc := user.NewService(repo)

	_, err := svc.GetProfile(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error from roles repo")
	}
}

func TestGetProfileEmptyPhoneNumber(t *testing.T) {
	repo := &fakeServiceRepo{
		user:          &user.User{ID: 3, Name: "Bob", CountryCode: "", PhoneNumber: ""},
		showroomRoles: []user.ShowroomRole{},
	}
	svc := user.NewService(repo)

	resp, err := svc.GetProfile(context.Background(), 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.PhoneNumber != nil {
		t.Fatalf("expected nil phone number, got %v", resp.PhoneNumber)
	}
}

func TestGetProfileMultipleShowroomRoles(t *testing.T) {
	repo := &fakeServiceRepo{
		user: &user.User{ID: 4, Name: "Carol", CountryCode: "+91", PhoneNumber: "7777777777"},
		showroomRoles: []user.ShowroomRole{
			{ShowroomID: 1, ShowroomName: "Showroom A", Role: user.UserRoleTypeOwner},
			{ShowroomID: 2, ShowroomName: "Showroom B", Role: user.UserRoleTypeManager},
		},
	}
	svc := user.NewService(repo)

	resp, err := svc.GetProfile(context.Background(), 4)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.ShowroomRoles) != 2 {
		t.Fatalf("expected 2 showroom roles, got %d", len(resp.ShowroomRoles))
	}
}
