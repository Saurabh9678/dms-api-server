package user_test

import (
	"context"
	"errors"
	"testing"

	"infiour.local/dms-api-server/internal/modules/user"
)

type fakeRepoStore struct {
	byID map[uint64]*user.User
}

func (f *fakeRepoStore) FindByID(_ context.Context, userID uint64) (*user.User, error) {
	u, ok := f.byID[userID]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func TestFindByIDReturnsUser(t *testing.T) {
	store := &fakeRepoStore{
		byID: map[uint64]*user.User{
			1: {ID: 1, Name: "Alice", CountryCode: "+91", PhoneNumber: "9999999999"},
		},
	}

	result, err := store.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != 1 || result.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", result)
	}
}

func TestFindByIDReturnsNotFound(t *testing.T) {
	store := &fakeRepoStore{byID: map[uint64]*user.User{}}

	_, err := store.FindByID(context.Background(), 99)
	if err != user.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFindByIDUserWithEmptyName(t *testing.T) {
	store := &fakeRepoStore{
		byID: map[uint64]*user.User{
			2: {ID: 2, Name: "", CountryCode: "+91", PhoneNumber: "8888888888"},
		},
	}

	result, err := store.FindByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != "" {
		t.Fatalf("expected empty name, got %q", result.Name)
	}
}

type fakeShowroomRolesStore struct {
	roles []user.ShowroomRole
	err   error
}

func (f *fakeShowroomRolesStore) FindShowroomRolesByUserID(_ context.Context, _ uint64) ([]user.ShowroomRole, error) {
	return f.roles, f.err
}

func TestFindShowroomRolesByUserIDReturnsRoles(t *testing.T) {
	store := &fakeShowroomRolesStore{
		roles: []user.ShowroomRole{
			{ShowroomID: 1, ShowroomName: "Showroom A", Role: user.UserRoleTypeOwner},
			{ShowroomID: 2, ShowroomName: "Showroom B", Role: user.UserRoleTypeManager},
		},
	}

	results, err := store.FindShowroomRolesByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(results))
	}
	if results[0].Role != user.UserRoleTypeOwner {
		t.Fatalf("expected owner role, got %s", results[0].Role)
	}
}

func TestFindShowroomRolesByUserIDReturnsEmpty(t *testing.T) {
	store := &fakeShowroomRolesStore{roles: []user.ShowroomRole{}}

	results, err := store.FindShowroomRolesByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty slice, got %d", len(results))
	}
}

func TestFindShowroomRolesByUserIDError(t *testing.T) {
	store := &fakeShowroomRolesStore{err: errors.New("db error")}

	_, err := store.FindShowroomRolesByUserID(context.Background(), 42)
	if err == nil {
		t.Fatalf("expected error")
	}
}
