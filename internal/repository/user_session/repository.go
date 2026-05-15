package usersession

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

func (r *Repository) Create(ctx context.Context, entity *user.UserSessionEntity) (*user.UserSessionEntity, error) {
	model := UserSession{
		UserID:           entity.UserID,
		Platform:         SessionPlatformType(entity.Platform),
		DeviceID:         entity.DeviceID,
		IPAddress:        entity.IPAddress,
		RefreshTokenHash: entity.RefreshTokenHash,
		Revoked:          entity.Revoked,
		Compromised:      entity.Compromised,
		RevokedReason:    entity.RevokedReason,
		CreatedAt:        entity.CreatedAt,
		LastUsedAt:       entity.LastUsedAt,
		ExpiresAt:        entity.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *Repository) FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*user.UserSessionEntity, error) {
	var model UserSession
	err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", refreshTokenHash).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, user.ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *Repository) RotateRefreshToken(ctx context.Context, sessionID uint64, refreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("id = ? AND revoked = false", sessionID).
		Updates(map[string]any{
			"refresh_token_hash": refreshTokenHash,
			"expires_at":         expiresAt,
			"last_used_at":       lastUsedAt,
		}).Error
}

func (r *Repository) Revoke(ctx context.Context, sessionID uint64, reason string, compromised bool, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&UserSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]any{
			"revoked":        true,
			"compromised":    compromised,
			"revoked_reason": reason,
			"last_used_at":   revokedAt,
		}).Error
}

func toDomain(model *UserSession) *user.UserSessionEntity {
	return &user.UserSessionEntity{
		ID:               model.ID,
		UserID:           model.UserID,
		Platform:         user.OTPPlatform(model.Platform),
		DeviceID:         model.DeviceID,
		IPAddress:        model.IPAddress,
		RefreshTokenHash: model.RefreshTokenHash,
		Revoked:          model.Revoked,
		Compromised:      model.Compromised,
		RevokedReason:    model.RevokedReason,
		CreatedAt:        model.CreatedAt,
		LastUsedAt:       model.LastUsedAt,
		ExpiresAt:        model.ExpiresAt,
	}
}
