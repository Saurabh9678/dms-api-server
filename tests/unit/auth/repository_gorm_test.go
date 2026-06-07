package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"infiour.local/dms-api-server/internal/modules/auth"
)

func newAuthMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return gormDB, mock
}

// ---------------------------------------------------------------------------
// OTPRepository
// ---------------------------------------------------------------------------

func TestNewOTPRepository(t *testing.T) {
	gormDB, _ := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)
	assert.NotNil(t, repo)
}

func TestOTPRepositoryWithTx(t *testing.T) {
	gormDB, _ := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)
	txRepo := repo.WithTx(gormDB)
	assert.NotNil(t, txRepo)
}

func TestOTPRepositoryCreate_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "user_otps"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	mock.ExpectCommit()

	now := time.Now()
	entity := &auth.UserOTP{
		CountryCode: "+91",
		PhoneNumber: "9999999999",
		RequestID:   "Ab12Cd34",
		OTPCode:     "123456",
		Platform:    auth.OTPPlatformWeb,
		OTPFor:      auth.OTPForMobile,
		DeviceID:    "device-1",
		ExpiresAt:   now.Add(5 * time.Minute),
		CreatedAt:   now,
	}

	result, err := repo.Create(context.Background(), entity)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), result.ID)
}

func TestOTPRepositoryCreate_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "user_otps"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), &auth.UserOTP{})
	assert.Error(t, err)
}

func TestOTPRepositoryFindLatestActive_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "country_code", "phone_number", "request_id", "otp_code", "platform", "otp_for",
		"device_id", "attempt_count", "resend_count", "is_used", "expires_at", "created_at", "verified_at",
	}).AddRow(
		uint64(10), "+91", "9999999999", "Ab12Cd34", "123456", "web", "mobile",
		"device-1", 0, 0, false, now.Add(5*time.Minute), now, nil,
	)

	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	result, err := repo.FindLatestActiveByRequestIDAndPlatform(context.Background(), "Ab12Cd34", auth.OTPPlatformWeb, auth.OTPForMobile)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(10), result.ID)
}

func TestOTPRepositoryFindLatestActive_NotFound(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindLatestActiveByRequestIDAndPlatform(context.Background(), "notfound", auth.OTPPlatformWeb, auth.OTPForMobile)
	assert.ErrorIs(t, err, auth.ErrInvalidOTP)
}

func TestOTPRepositoryFindLatestActive_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("connection error"))

	_, err := repo.FindLatestActiveByRequestIDAndPlatform(context.Background(), "Ab12Cd34", auth.OTPPlatformWeb, auth.OTPForMobile)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, auth.ErrInvalidOTP)
}

func TestOTPRepositoryIncrementAttempt_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_otps"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.IncrementAttempt(context.Background(), 10)
	assert.NoError(t, err)
}

func TestOTPRepositoryIncrementAttempt_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_otps"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.IncrementAttempt(context.Background(), 10)
	assert.Error(t, err)
}

func TestOTPRepositoryMarkUsed_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_otps"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.MarkUsed(context.Background(), 10, time.Now())
	assert.NoError(t, err)
}

func TestOTPRepositoryMarkUsed_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_otps"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.MarkUsed(context.Background(), 10, time.Now())
	assert.Error(t, err)
}

func TestOTPRepositoryFindLatestByPhone_Found(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "country_code", "phone_number", "request_id", "otp_code", "platform", "otp_for",
		"device_id", "attempt_count", "resend_count", "is_used", "expires_at", "created_at", "verified_at",
	}).AddRow(
		uint64(20), "+91", "9999999999", "XyZ12345", "654321", "web", "mobile",
		"device-2", 0, 0, false, now.Add(5*time.Minute), now, nil,
	)

	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	result, err := repo.FindLatestByPhone(context.Background(), "+91", "9999999999")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(20), result.ID)
}

func TestOTPRepositoryFindLatestByPhone_NotFound(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.FindLatestByPhone(context.Background(), "+91", "0000000000")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestOTPRepositoryFindLatestByPhone_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("connection error"))

	result, err := repo.FindLatestByPhone(context.Background(), "+91", "9999999999")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestOTPRepositoryCountRecentByPhone_NonZero(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(3))
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	count, err := repo.CountRecentByPhone(context.Background(), "+91", "9999999999", time.Now().Add(-24*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestOTPRepositoryCountRecentByPhone_Zero(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(0))
	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	count, err := repo.CountRecentByPhone(context.Background(), "+91", "0000000000", time.Now().Add(-24*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestOTPRepositoryCountRecentByPhone_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewOTPRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("db error"))

	count, err := repo.CountRecentByPhone(context.Background(), "+91", "9999999999", time.Now().Add(-24*time.Hour))
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
}

// ---------------------------------------------------------------------------
// SessionRepository
// ---------------------------------------------------------------------------

func TestNewSessionRepository(t *testing.T) {
	gormDB, _ := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)
	assert.NotNil(t, repo)
}

func TestSessionRepositoryWithTx(t *testing.T) {
	gormDB, _ := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)
	txRepo := repo.WithTx(gormDB)
	assert.NotNil(t, txRepo)
}

func TestSessionRepositoryCreate_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "user_sessions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(501)))
	mock.ExpectCommit()

	now := time.Now()
	expiry := now.Add(7 * 24 * time.Hour)
	entity := &auth.UserSession{
		UserID:           1,
		Platform:         auth.OTPPlatformWeb,
		DeviceID:         "device-1",
		RefreshTokenHash: "hash-value",
		LastUsedAt:       now,
		CreatedAt:        now,
		ExpiresAt:        &expiry,
	}

	result, err := repo.Create(context.Background(), entity)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(501), result.ID)
}

func TestSessionRepositoryCreate_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "user_sessions"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), &auth.UserSession{})
	assert.Error(t, err)
}

func TestSessionRepositoryFindByRefreshTokenHash_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	now := time.Now()
	expiry := now.Add(7 * 24 * time.Hour)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "platform", "device_id", "ip_address",
		"refresh_token_hash", "revoked", "compromised", "revoked_reason",
		"created_at", "last_used_at", "expires_at",
	}).AddRow(
		uint64(501), uint64(1), "web", "device-1", "127.0.0.1",
		"hash-value", false, false, "",
		now, now, expiry,
	)

	mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	result, err := repo.FindByRefreshTokenHash(context.Background(), "hash-value")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(501), result.ID)
}

func TestSessionRepositoryFindByRefreshTokenHash_NotFound(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByRefreshTokenHash(context.Background(), "notfound")
	assert.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
}

func TestSessionRepositoryFindByRefreshTokenHash_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("connection error"))

	_, err := repo.FindByRefreshTokenHash(context.Background(), "hash-value")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, auth.ErrInvalidRefreshToken)
}

func TestSessionRepositoryRotateRefreshToken_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_sessions"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.RotateRefreshToken(context.Background(), 501, "new-hash", time.Now().Add(7*24*time.Hour), time.Now())
	assert.NoError(t, err)
}

func TestSessionRepositoryRotateRefreshToken_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_sessions"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.RotateRefreshToken(context.Background(), 501, "new-hash", time.Now().Add(7*24*time.Hour), time.Now())
	assert.Error(t, err)
}

func TestSessionRepositoryRevoke_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_sessions"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Revoke(context.Background(), 501, "expired", false, time.Now())
	assert.NoError(t, err)
}

func TestSessionRepositoryRevoke_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_sessions"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Revoke(context.Background(), 501, "expired", false, time.Now())
	assert.Error(t, err)
}

func TestSessionRepositoryRevokeAllByUserIDAndPlatform_Success(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_sessions"`).
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.RevokeAllByUserIDAndPlatform(context.Background(), 1, auth.OTPPlatformWeb, "logout", false, time.Now())
	assert.NoError(t, err)
}

func TestSessionRepositoryRevokeAllByUserIDAndPlatform_DBError(t *testing.T) {
	gormDB, mock := newAuthMockDB(t)
	repo := auth.NewSessionRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_sessions"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.RevokeAllByUserIDAndPlatform(context.Background(), 1, auth.OTPPlatformWeb, "logout", false, time.Now())
	assert.Error(t, err)
}
