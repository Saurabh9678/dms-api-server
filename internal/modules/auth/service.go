package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/internal/modules/user"
	otpprovider "infiour.local/dms-api-server/internal/providers/otp"
	tokenprovider "infiour.local/dms-api-server/internal/providers/token"
	"infiour.local/dms-api-server/pkg/config"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*TriggerOTPResponse, error)
	Login(ctx context.Context, req LoginRequest) (*TriggerOTPResponse, error)
	SendOTP(ctx context.Context, req SendOTPRequest) (*TriggerOTPResponse, error)
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error)
	Logout(ctx context.Context, req LogoutRequest) error
}

type userRepo interface {
	FindByPhone(ctx context.Context, countryCode string, phoneNumber string) (*user.User, error)
	Create(ctx context.Context, record *user.User) (*user.User, error)
}

type otpRepo interface {
	Create(ctx context.Context, entity *UserOTP) (*UserOTP, error)
	FindLatestActiveByRequestIDAndPlatform(ctx context.Context, requestID string, platform OTPPlatform, otpFor OTPFor) (*UserOTP, error)
	IncrementAttempt(ctx context.Context, otpID uint64) error
	MarkUsed(ctx context.Context, otpID uint64, verifiedAt time.Time) error
	FindLatestByPhone(ctx context.Context, countryCode string, phoneNumber string) (*UserOTP, error)
	CountRecentByPhone(ctx context.Context, countryCode string, phoneNumber string, since time.Time) (int64, error)
}

type sessionRepo interface {
	Create(ctx context.Context, entity *UserSession) (*UserSession, error)
	FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*UserSession, error)
	RotateRefreshToken(ctx context.Context, sessionID uint64, refreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error
	Revoke(ctx context.Context, sessionID uint64, reason string, compromised bool, revokedAt time.Time) error
	RevokeAllByUserIDAndPlatform(ctx context.Context, userID uint64, platform OTPPlatform, reason string, compromised bool, revokedAt time.Time) error
}

type service struct {
	users         userRepo
	otps          otpRepo
	sessions      sessionRepo
	otpProvider   otpprovider.Provider
	tokenProvider tokenprovider.Provider
	config        config.AuthConfig
	db            *gorm.DB
	nowFn         func() time.Time
	env           string
}

const (
	requestIDLength          = 8
	requestIDGenerateRetries = 5
)

func NewService(
	users userRepo,
	otps otpRepo,
	sessions sessionRepo,
	otpProvider otpprovider.Provider,
	tokenProvider tokenprovider.Provider,
	cfg config.AuthConfig,
	db *gorm.DB,
	env string,
) Service {
	if cfg.OTPTTL <= 0 {
		cfg.OTPTTL = 5 * time.Minute
	}
	if cfg.OTPMaxAttempts <= 0 {
		cfg.OTPMaxAttempts = 5
	}
	if cfg.OTPCooldownSeconds <= 0 {
		cfg.OTPCooldownSeconds = 60
	}
	if cfg.OTPMaxDailySends <= 0 {
		cfg.OTPMaxDailySends = 10
	}
	return &service{
		users:         users,
		otps:          otps,
		sessions:      sessions,
		otpProvider:   otpProvider,
		tokenProvider: tokenProvider,
		config:        cfg,
		db:            db,
		nowFn:         time.Now,
		env:           env,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*TriggerOTPResponse, error) {
	return s.triggerOTP(ctx, req.CountryCode, req.PhoneNumber, req.Platform, req.DeviceID)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*TriggerOTPResponse, error) {
	return s.triggerOTP(ctx, req.CountryCode, req.PhoneNumber, req.Platform, req.DeviceID)
}

func (s *service) SendOTP(ctx context.Context, req SendOTPRequest) (*TriggerOTPResponse, error) {
	return s.triggerOTP(ctx, req.CountryCode, req.PhoneNumber, req.Platform, req.DeviceID)
}

func (s *service) triggerOTP(ctx context.Context, countryCode string, phoneNumber string, platformValue string, deviceID string) (*TriggerOTPResponse, error) {
	normalizedCountryCode := strings.TrimSpace(countryCode)
	normalizedPhoneNumber := strings.TrimSpace(phoneNumber)
	platform := OTPPlatform(platformValue)
	now := s.nowFn()

	// Rate limiting by phone — no user lookup required
	latest, err := s.otps.FindLatestByPhone(ctx, normalizedCountryCode, normalizedPhoneNumber)
	if err != nil {
		return nil, err
	}
	if latest != nil && now.Sub(latest.CreatedAt) < time.Duration(s.config.OTPCooldownSeconds)*time.Second {
		return nil, ErrOTPCooldown
	}

	count, err := s.otps.CountRecentByPhone(ctx, normalizedCountryCode, normalizedPhoneNumber, now.Add(-24*time.Hour))
	if err != nil {
		return nil, err
	}
	if count >= int64(s.config.OTPMaxDailySends) {
		return nil, ErrOTPRateLimitExceeded
	}

	code := generateOTPCode()
	requestID := ""
	for attempt := 0; attempt < requestIDGenerateRetries; attempt++ {
		requestID = generateRequestID(requestIDLength)
		_, err = s.otps.Create(ctx, &UserOTP{
			CountryCode: normalizedCountryCode,
			PhoneNumber: normalizedPhoneNumber,
			RequestID:   requestID,
			OTPCode:     code,
			Platform:    platform,
			OTPFor:      OTPForMobile,
			DeviceID:    strings.TrimSpace(deviceID),
			ExpiresAt:   now.Add(s.config.OTPTTL),
			CreatedAt:   now,
		})
		if err == nil {
			break
		}
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	destination := fmt.Sprintf("%s%s", normalizedCountryCode, normalizedPhoneNumber)
	if err := s.otpProvider.Send(ctx, destination, code); err != nil {
		return nil, err
	}

	resp := &TriggerOTPResponse{
		Message:   "OTP sent successfully",
		RequestID: requestID,
	}
	if s.env == "development" || s.env == "staging" {
		resp.OTPCode = &code
	}
	return resp, nil
}

func (s *service) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error) {
	platform := OTPPlatform(req.Platform)
	otpRecord, err := s.otps.FindLatestActiveByRequestIDAndPlatform(ctx, strings.TrimSpace(req.RequestID), platform, OTPForMobile)
	if err != nil {
		return nil, err
	}

	now := s.nowFn()
	if otpRecord.IsUsed {
		return nil, ErrOTPAlreadyUsed
	}
	if now.After(otpRecord.ExpiresAt) {
		return nil, ErrOTPExpired
	}
	if otpRecord.AttemptCount >= s.config.OTPMaxAttempts {
		return nil, ErrOTPAttemptsExceeded
	}
	if otpRecord.OTPCode != strings.TrimSpace(req.OTPCode) {
		_ = s.otps.IncrementAttempt(ctx, otpRecord.ID)
		return nil, ErrInvalidOTP
	}
	if err := s.otps.MarkUsed(ctx, otpRecord.ID, now); err != nil {
		return nil, err
	}

	// Find or create user — users table is the canonical identity source.
	// The OTP record provides the phone snapshot; the resulting user is authoritative.
	foundUser, err := s.users.FindByPhone(ctx, otpRecord.CountryCode, otpRecord.PhoneNumber)
	if errors.Is(err, user.ErrUserNotFound) {
		foundUser, err = s.users.Create(ctx, &user.User{
			CountryCode: otpRecord.CountryCode,
			PhoneNumber: otpRecord.PhoneNumber,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				// Concurrent request won the race — re-fetch the surviving record
				foundUser, err = s.users.FindByPhone(ctx, otpRecord.CountryCode, otpRecord.PhoneNumber)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	} else if err != nil {
		return nil, err
	}

	pair, err := s.tokenProvider.Issue(foundUser.ID)
	if err != nil {
		return nil, err
	}
	refreshExpiry := now.Add(time.Duration(pair.RefreshTokenTTL) * time.Second)
	if err := s.sessions.RevokeAllByUserIDAndPlatform(ctx, foundUser.ID, platform, "new session issued for platform", false, now); err != nil {
		return nil, err
	}
	_, err = s.sessions.Create(ctx, &UserSession{
		UserID:           foundUser.ID,
		Platform:         platform,
		DeviceID:         strings.TrimSpace(req.DeviceID),
		RefreshTokenHash: pair.RefreshTokenHash,
		LastUsedAt:       now,
		CreatedAt:        now,
		ExpiresAt:        &refreshExpiry,
	})
	if err != nil {
		return nil, err
	}

	return &VerifyOTPResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.AccessTokenTTL,
		TokenType:    "Bearer",
		RequiredName: foundUser.Name == "",
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error) {
	plain := strings.TrimSpace(req.RefreshToken)
	hashed := s.tokenProvider.HashRefreshToken(plain)

	session, err := s.sessions.FindByRefreshTokenHash(ctx, hashed)
	if err != nil {
		return nil, err
	}
	if session.Revoked {
		return nil, ErrSessionRevoked
	}
	now := s.nowFn()
	if session.ExpiresAt != nil && now.After(*session.ExpiresAt) {
		_ = s.sessions.Revoke(ctx, session.ID, "refresh token expired", false, now)
		return nil, ErrInvalidRefreshToken
	}

	pair, err := s.tokenProvider.Rotate(session.UserID)
	if err != nil {
		return nil, err
	}
	refreshExpiry := now.Add(time.Duration(pair.RefreshTokenTTL) * time.Second)
	if err := s.sessions.RotateRefreshToken(ctx, session.ID, pair.RefreshTokenHash, refreshExpiry, now); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.AccessTokenTTL,
		TokenType:    "Bearer",
	}, nil
}

func (s *service) Logout(ctx context.Context, req LogoutRequest) error {
	userID, err := s.tokenProvider.ParseAccessToken(strings.TrimSpace(req.AccessToken))
	if err != nil {
		return ErrInvalidAccessToken
	}
	return s.sessions.RevokeAllByUserIDAndPlatform(
		ctx,
		userID,
		OTPPlatform(strings.TrimSpace(req.Platform)),
		"user logout",
		false,
		s.nowFn(),
	)
}

func generateOTPCode() string {
	max := big.NewInt(1000000)
	// crypto/rand panics internally on entropy failure (Go 1.20+); error is unreachable.
	value, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", value.Int64())
}

func generateRequestID(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	max := big.NewInt(int64(len(chars)))
	for i := 0; i < length; i++ {
		// crypto/rand panics internally on entropy failure (Go 1.20+); error is unreachable.
		value, _ := rand.Int(rand.Reader, max)
		result[i] = chars[value.Int64()]
	}
	return string(result)
}
