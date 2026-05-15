package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type OTPRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) WithTx(tx *gorm.DB) *OTPRepository {
	return &OTPRepository{db: tx}
}

func (r *OTPRepository) Create(ctx context.Context, entity *UserOTP) (*UserOTP, error) {
	model := *entity
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *OTPRepository) FindLatestActiveByUserAndPlatform(ctx context.Context, userID uint64, platform OTPPlatform, otpFor OTPFor) (*UserOTP, error) {
	var model UserOTP
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND platform = ? AND otp_for = ?", userID, string(platform), string(otpFor)).
		Order("created_at DESC").
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidOTP
	}
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *OTPRepository) IncrementAttempt(ctx context.Context, otpID uint64) error {
	return r.db.WithContext(ctx).Model(&UserOTP{}).
		Where("id = ?", otpID).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error
}

func (r *OTPRepository) MarkUsed(ctx context.Context, otpID uint64, verifiedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&UserOTP{}).
		Where("id = ? AND is_used = false", otpID).
		Updates(map[string]any{
			"is_used":     true,
			"verified_at": verifiedAt,
		}).Error
}

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) WithTx(tx *gorm.DB) *SessionRepository {
	return &SessionRepository{db: tx}
}

func (r *SessionRepository) Create(ctx context.Context, entity *UserSession) (*UserSession, error) {
	model := *entity
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *SessionRepository) FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*UserSession, error) {
	var model UserSession
	err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", refreshTokenHash).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *SessionRepository) RotateRefreshToken(ctx context.Context, sessionID uint64, refreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("id = ? AND revoked = false", sessionID).
		Updates(map[string]any{
			"refresh_token_hash": refreshTokenHash,
			"expires_at":         expiresAt,
			"last_used_at":       lastUsedAt,
		}).Error
}

func (r *SessionRepository) Revoke(ctx context.Context, sessionID uint64, reason string, compromised bool, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]any{
			"revoked":        true,
			"compromised":    compromised,
			"revoked_reason": reason,
			"last_used_at":   revokedAt,
		}).Error
}
