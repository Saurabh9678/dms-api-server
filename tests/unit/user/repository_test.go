package user_test

import (
	"context"
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
