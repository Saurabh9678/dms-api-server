package user

import (
	"context"
	"time"
)

type UserRepository interface {
	FindByPhone(ctx context.Context, countryCode string, phoneNumber string) (*UserEntity, error)
	Create(ctx context.Context, entity *UserEntity) (*UserEntity, error)
}

type OTPRepository interface {
	Create(ctx context.Context, entity *UserOTPEntity) (*UserOTPEntity, error)
	FindLatestActiveByUserAndPlatform(ctx context.Context, userID uint64, platform OTPPlatform, otpFor OTPFor) (*UserOTPEntity, error)
	IncrementAttempt(ctx context.Context, otpID uint64) error
	MarkUsed(ctx context.Context, otpID uint64, verifiedAt time.Time) error
}

type SessionRepository interface {
	Create(ctx context.Context, entity *UserSessionEntity) (*UserSessionEntity, error)
	FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*UserSessionEntity, error)
	RotateRefreshToken(ctx context.Context, sessionID uint64, refreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error
	Revoke(ctx context.Context, sessionID uint64, reason string, compromised bool, revokedAt time.Time) error
}
