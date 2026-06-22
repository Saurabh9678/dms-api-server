package showroom_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/internal/modules/showroom"
)

func newShowroomMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func TestShowroomTableName(t *testing.T) {
	assert.Equal(t, "showrooms", showroom.Showroom{}.TableName())
}

func TestCreateWithOwner_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "showrooms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(5), "owner"))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	s := &showroom.Showroom{Name: "Test Showroom"}
	result, err := repo.CreateWithOwner(context.Background(), uint64(10), s)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), result.ID)
}

func TestCreateWithOwner_CreateError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "showrooms"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreateWithOwner(context.Background(), uint64(10), &showroom.Showroom{Name: "X"})
	assert.Error(t, err)
}

func TestCreateWithOwner_OwnerRoleNotFound(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "showrooms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}))
	mock.ExpectRollback()

	_, err := repo.CreateWithOwner(context.Background(), uint64(10), &showroom.Showroom{Name: "X"})
	assert.ErrorIs(t, err, showroom.ErrOwnerRoleNotFound)
}

func TestCreateWithOwner_RoleLookupDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "showrooms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreateWithOwner(context.Background(), uint64(10), &showroom.Showroom{Name: "X"})
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrOwnerRoleNotFound)
}

func TestCreateWithOwner_RelationInsertError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "showrooms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(5), "owner"))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.CreateWithOwner(context.Background(), uint64(10), &showroom.Showroom{Name: "X"})
	assert.Error(t, err)
}

func TestUpdateFilePaths_BothNil(t *testing.T) {
	gormDB, _ := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	err := repo.UpdateFilePaths(context.Background(), uint64(1), nil, nil)
	assert.NoError(t, err)
}

func TestUpdateFilePaths_LogoOnly(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	logo := "path/to/logo.jpg"
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "showrooms"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateFilePaths(context.Background(), uint64(1), &logo, nil)
	assert.NoError(t, err)
}

func TestUpdateFilePaths_BannerOnly(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	banner := "path/to/banner.jpg"
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "showrooms"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateFilePaths(context.Background(), uint64(1), nil, &banner)
	assert.NoError(t, err)
}

func TestUpdateFilePaths_Both(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	logo := "path/to/logo.jpg"
	banner := "path/to/banner.jpg"
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "showrooms"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateFilePaths(context.Background(), uint64(1), &logo, &banner)
	assert.NoError(t, err)
}

func TestUpdateFilePaths_DBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	logo := "path/to/logo.jpg"
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "showrooms"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.UpdateFilePaths(context.Background(), uint64(1), &logo, nil)
	assert.Error(t, err)
}
