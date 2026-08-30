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

// ─── CreateWithOwner ────────────────────────────────────────────────────────

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

	s := &showroom.Showroom{ShowroomID: "SHOP0001", Name: "Test Showroom"}
	result, err := repo.CreateWithOwner(context.Background(), uint64(10), s)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "SHOP0001", result.ShowroomID)
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

// ─── UpdateFilePaths ─────────────────────────────────────────────────────────

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

// ─── AddMember ───────────────────────────────────────────────────────────────

func TestAddMember_CreatesUserWhenMissing_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	// phone lookup — not found
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}))
	// create user
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(99)))
	// role lookup
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	// duplicate check
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// insert relation
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	userID, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.NoError(t, err)
	assert.Equal(t, uint64(99), userID)
}

func TestAddMember_ExistingUser_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Existing", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	userID, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.NoError(t, err)
	assert.Equal(t, uint64(99), userID)
}

func TestAddMember_ExistingUserEmptyName_UpdatesName(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "", "91", "9876543210", nil, nil, nil))
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	userID, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.NoError(t, err)
	assert.Equal(t, uint64(99), userID)
}

func TestAddMember_UserLookupDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
}

func TestAddMember_CreateUserError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
}

func TestAddMember_CreateUserRace_RefetchSuccess(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Existing", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	userID, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.NoError(t, err)
	assert.Equal(t, uint64(99), userID)
}

func TestAddMember_CreateUserRace_RefetchEmptyNameUpdates(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "", "91", "9876543210", nil, nil, nil))
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	userID, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.NoError(t, err)
	assert.Equal(t, uint64(99), userID)
}

func TestAddMember_CreateUserRace_RefetchError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
}

func TestAddMember_UpdateNameError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "", "91", "9876543210", nil, nil, nil))
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
}

func TestAddMember_RoleNotFound(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}))
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.ErrorIs(t, err, showroom.ErrMemberRoleNotFound)
}

func TestAddMember_RoleLookupDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrMemberRoleNotFound)
}

func TestAddMember_DuplicateMember(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.ErrorIs(t, err, showroom.ErrDuplicateMember)
}

func TestAddMember_CountCheckDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrDuplicateMember)
}

func TestAddMember_InsertError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
}

func TestAddMember_UniqueConflict_RestoresSoftDeleted(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	userID, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.NoError(t, err)
	assert.Equal(t, uint64(99), userID)
}

func TestAddMember_UniqueConflict_RestoreNoRow_DuplicateMember(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.ErrorIs(t, err, showroom.ErrDuplicateMember)
}

func TestAddMember_UniqueConflict_RestoreDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "Jane", "91", "9876543210", nil, nil, nil))
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(3), "employee"))
	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO "user_showroom_relations"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrDuplicateMember)
}

func TestAddMember_RaceRefetchUpdateNameError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "country_code", "phone_number", "created_at", "updated_at", "deleted_at"}).
			AddRow(uint64(99), "", "91", "9876543210", nil, nil, nil))
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	_, err := repo.AddMember(context.Background(), uint64(1), "Jane Doe", "91", "9876543210", "employee")
	assert.Error(t, err)
}

// ─── ListMembers ─────────────────────────────────────────────────────────────

func TestListMembers_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT usr.user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "name", "country_code", "phone_number", "role"}).
			AddRow(uint64(10), "Alice", "+91", "9999999999", "manager").
			AddRow(uint64(11), "", "", "", "employee"))

	records, total, err := repo.ListMembers(context.Background(), uint64(1), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, records, 2)
	assert.Equal(t, "Alice", records[0].Name)
	assert.Equal(t, "manager", records[0].Role)
}

func TestListMembers_Empty(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT usr.user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "name", "country_code", "phone_number", "role"}))

	records, total, err := repo.ListMembers(context.Background(), uint64(1), 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, records)
}

func TestListMembers_CountDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)

	_, _, err := repo.ListMembers(context.Background(), uint64(1), 1, 20)
	assert.Error(t, err)
}

func TestListMembers_ScanDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "user_showroom_relations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT usr.user_id`).
		WillReturnError(gorm.ErrInvalidData)

	_, _, err := repo.ListMembers(context.Background(), uint64(1), 1, 20)
	assert.Error(t, err)
}

// ─── GetMemberRole ────────────────────────────────────────────────────────────

func TestGetMemberRole_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT ur.type AS role`).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("employee"))

	role, err := repo.GetMemberRole(context.Background(), uint64(1), uint64(99))
	assert.NoError(t, err)
	assert.Equal(t, "employee", role)
}

func TestGetMemberRole_NotFound(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT ur.type AS role`).
		WillReturnRows(sqlmock.NewRows([]string{"role"}))

	_, err := repo.GetMemberRole(context.Background(), uint64(1), uint64(99))
	assert.ErrorIs(t, err, showroom.ErrMemberNotFound)
}

func TestGetMemberRole_DBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT ur.type AS role`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetMemberRole(context.Background(), uint64(1), uint64(99))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrMemberNotFound)
}

// ─── RemoveMember ────────────────────────────────────────────────────────────

func TestRemoveMember_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.RemoveMember(context.Background(), uint64(1), uint64(99))
	assert.NoError(t, err)
}

func TestRemoveMember_NotFound(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.RemoveMember(context.Background(), uint64(1), uint64(99))
	assert.ErrorIs(t, err, showroom.ErrMemberNotFound)
}

func TestRemoveMember_DBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.RemoveMember(context.Background(), uint64(1), uint64(99))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrMemberNotFound)
}

// ─── UpdateMemberRole ─────────────────────────────────────────────────────────

func TestUpdateMemberRole_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(2), "manager"))
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateMemberRole(context.Background(), uint64(1), uint64(99), "manager")
	assert.NoError(t, err)
}

func TestUpdateMemberRole_RoleNotFound(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}))
	mock.ExpectRollback()

	err := repo.UpdateMemberRole(context.Background(), uint64(1), uint64(99), "manager")
	assert.ErrorIs(t, err, showroom.ErrMemberRoleNotFound)
}

func TestUpdateMemberRole_RoleLookupDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.UpdateMemberRole(context.Background(), uint64(1), uint64(99), "manager")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrMemberRoleNotFound)
}

func TestUpdateMemberRole_MemberNotFound(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(2), "manager"))
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.UpdateMemberRole(context.Background(), uint64(1), uint64(99), "manager")
	assert.ErrorIs(t, err, showroom.ErrMemberNotFound)
}

func TestUpdateMemberRole_UpdateDBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "user_roles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(uint64(2), "manager"))
	mock.ExpectExec(`UPDATE "user_showroom_relations"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.UpdateMemberRole(context.Background(), uint64(1), uint64(99), "manager")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrMemberNotFound)
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func TestGetByID_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "showrooms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "showroom_id", "name"}).AddRow(uint64(1), "SHOP0001", "Test Showroom"))

	result, err := repo.GetByID(context.Background(), uint64(1))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "SHOP0001", result.ShowroomID)
	assert.Equal(t, "Test Showroom", result.Name)
}

func TestGetByID_NotFound(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "showrooms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	_, err := repo.GetByID(context.Background(), uint64(99))
	assert.ErrorIs(t, err, showroom.ErrShowroomNotFound)
}

func TestGetByID_DBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "showrooms"`).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.GetByID(context.Background(), uint64(1))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, showroom.ErrShowroomNotFound)
}

// ─── ListByUserID ─────────────────────────────────────────────────────────────

func TestListByUserID_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT s\.id, s\.showroom_id, s\.name, ur\.type AS role`).
		WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "showroom_id", "name", "role"}).
			AddRow(uint64(1), "SHOP0001", "Main Showroom", "owner").
			AddRow(uint64(2), "SHOP0002", "Second Showroom", "manager"))

	records, err := repo.ListByUserID(context.Background(), uint64(10))
	assert.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, uint64(1), records[0].ID)
	assert.Equal(t, "SHOP0001", records[0].ShowroomID)
	assert.Equal(t, "Main Showroom", records[0].Name)
	assert.Equal(t, "owner", records[0].Role)
}

func TestListByUserID_Empty(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT s\.id, s\.showroom_id, s\.name, ur\.type AS role`).
		WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "showroom_id", "name", "role"}))

	records, err := repo.ListByUserID(context.Background(), uint64(10))
	assert.NoError(t, err)
	assert.Empty(t, records)
}

func TestListByUserID_DBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectQuery(`SELECT s\.id, s\.showroom_id, s\.name, ur\.type AS role`).
		WithArgs(uint64(10)).
		WillReturnError(gorm.ErrInvalidData)

	_, err := repo.ListByUserID(context.Background(), uint64(10))
	assert.Error(t, err)
}

// ─── UpdateShowroomFields ─────────────────────────────────────────────────────

func TestUpdateShowroomFields_Success(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "showrooms"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateShowroomFields(context.Background(), uint64(1), map[string]any{"name": "New Name"})
	assert.NoError(t, err)
}

func TestUpdateShowroomFields_DBError(t *testing.T) {
	gormDB, mock := newShowroomMockDB(t)
	repo := showroom.NewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "showrooms"`).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.UpdateShowroomFields(context.Background(), uint64(1), map[string]any{"name": "X"})
	assert.Error(t, err)
}
