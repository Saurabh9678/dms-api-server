package user_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/internal/modules/user"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return gormDB, mock
}

func TestRepositoryNewRepository(t *testing.T) {
	gormDB, _ := newMockDB(t)
	repo := user.NewRepository(gormDB)
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}

func TestRepositoryWithTx(t *testing.T) {
	gormDB, _ := newMockDB(t)
	repo := user.NewRepository(gormDB)
	txRepo := repo.WithTx(gormDB)
	if txRepo == nil {
		t.Fatal("expected non-nil tx repository")
	}
}

func TestRepositoryFindByIDFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	rows := sqlmock.NewRows([]string{"id", "email", "phone_number", "country_code", "name", "created_at", "updated_at", "deleted_at"}).
		AddRow(uint64(1), "a@b.com", "9999999999", "+91", "Alice", time.Now(), time.Now(), nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnRows(rows)

	u, err := repo.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.ID != 1 {
		t.Fatalf("expected ID 1, got %d", u.ID)
	}
}

func TestRepositoryFindByIDNotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(context.Background(), 99)
	if err != user.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRepositoryFindByIDError(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.FindByID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryFindByPhoneFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	rows := sqlmock.NewRows([]string{"id", "email", "phone_number", "country_code", "name", "created_at", "updated_at", "deleted_at"}).
		AddRow(uint64(2), "", "9999999999", "+91", "Bob", time.Now(), time.Now(), nil)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnRows(rows)

	u, err := repo.FindByPhone(context.Background(), "+91", "9999999999")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.PhoneNumber != "9999999999" {
		t.Fatalf("unexpected phone: %s", u.PhoneNumber)
	}
}

func TestRepositoryFindByPhoneNotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByPhone(context.Background(), "+91", "0000000000")
	if err != user.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRepositoryFindByPhoneError(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.FindByPhone(context.Background(), "+91", "9999999999")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryCreateSuccess(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(10)))
	mock.ExpectCommit()

	record := &user.User{PhoneNumber: "9999999999", CountryCode: "+91"}
	created, err := repo.Create(context.Background(), record)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil {
		t.Fatal("expected created user")
	}
}

func TestRepositoryCreateError(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), &user.User{PhoneNumber: "9999999999", CountryCode: "+91"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryUpdateNameSuccess(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateName(context.Background(), 1, "Alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRepositoryUpdateNameNotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users"`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.UpdateName(context.Background(), 99, "Alice")
	if err != user.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRepositoryUpdateNameError(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users"`)).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.UpdateName(context.Background(), 1, "Alice")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryFindShowroomRolesByUserIDSuccess(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	rows := sqlmock.NewRows([]string{"showroom_id", "showroom_name", "role"}).
		AddRow(uint64(1), "Showroom A", "owner").
		AddRow(uint64(2), "Showroom B", "manager")
	mock.ExpectQuery(`user_showroom_relations`).
		WillReturnRows(rows)

	results, err := repo.FindShowroomRolesByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestRepositoryFindShowroomRolesByUserIDEmpty(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(`user_showroom_relations`).
		WillReturnRows(sqlmock.NewRows([]string{"showroom_id", "showroom_name", "role"}))

	results, err := repo.FindShowroomRolesByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty, got %d", len(results))
	}
}

func TestRepositoryFindShowroomRolesByUserIDError(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(`user_showroom_relations`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.FindShowroomRolesByUserID(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepositoryLoadUserShowroomRolesSuccess(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	rows := sqlmock.NewRows([]string{"showroom_id", "showroom_name", "role"}).
		AddRow(uint64(1), "Showroom A", "owner").
		AddRow(uint64(2), "Showroom B", "manager")
	mock.ExpectQuery(`user_showroom_relations`).
		WillReturnRows(rows)

	result, err := repo.LoadUserShowroomRoles(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[1] != "owner" {
		t.Fatalf("expected owner for showroom 1, got %s", result[1])
	}
	if result[2] != "manager" {
		t.Fatalf("expected manager for showroom 2, got %s", result[2])
	}
}

func TestRepositoryLoadUserShowroomRolesEmpty(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(`user_showroom_relations`).
		WillReturnRows(sqlmock.NewRows([]string{"showroom_id", "showroom_name", "role"}))

	result, err := repo.LoadUserShowroomRoles(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(result))
	}
}

func TestRepositoryLoadUserShowroomRolesError(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := user.NewRepository(gormDB)

	mock.ExpectQuery(`user_showroom_relations`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.LoadUserShowroomRoles(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error")
	}
}
