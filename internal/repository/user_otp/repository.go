package userotp

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/internal/domain/user"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, entity *user.UserOTPEntity) (*user.UserOTPEntity, error) {
	model := UserOTP{
		UserID:       entity.UserID,
		OTPCode:      entity.OTPCode,
		Platform:     PlatformType(entity.Platform),
		OTPFor:       OTPForType(entity.OTPFor),
		DeviceID:     entity.DeviceID,
		AttemptCount: entity.AttemptCount,
		ResendCount:  entity.ResendCount,
		IsUsed:       entity.IsUsed,
		ExpiresAt:    entity.ExpiresAt,
		CreatedAt:    entity.CreatedAt,
		VerifiedAt:   entity.VerifiedAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *Repository) FindLatestActiveByUserAndPlatform(ctx context.Context, userID uint64, platform user.OTPPlatform, otpFor user.OTPFor) (*user.UserOTPEntity, error) {
	var model UserOTP
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND platform = ? AND otp_for = ?", userID, string(platform), string(otpFor)).
		Order("created_at DESC").
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrInvalidOTP
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *Repository) IncrementAttempt(ctx context.Context, otpID uint64) error {
	return r.db.WithContext(ctx).Model(&UserOTP{}).
		Where("id = ?", otpID).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error
}

func (r *Repository) MarkUsed(ctx context.Context, otpID uint64, verifiedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&UserOTP{}).
		Where("id = ? AND is_used = false", otpID).
		Updates(map[string]any{
			"is_used":     true,
			"verified_at": verifiedAt,
		}).Error
}

func toDomain(model *UserOTP) *user.UserOTPEntity {
	return &user.UserOTPEntity{
		ID:           model.ID,
		UserID:       model.UserID,
		OTPCode:      model.OTPCode,
		Platform:     user.OTPPlatform(model.Platform),
		OTPFor:       user.OTPFor(model.OTPFor),
		DeviceID:     model.DeviceID,
		AttemptCount: model.AttemptCount,
		ResendCount:  model.ResendCount,
		IsUsed:       model.IsUsed,
		ExpiresAt:    model.ExpiresAt,
		CreatedAt:    model.CreatedAt,
		VerifiedAt:   model.VerifiedAt,
	}
}
