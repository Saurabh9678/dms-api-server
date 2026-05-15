package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"infiour.local/dms-api-server/internal/domain/user"
)

type Config struct {
	OTPTTL         time.Duration
	OTPMaxAttempts int
	OTPFor         user.OTPFor
}

type Service struct {
	users    user.UserRepository
	otps     user.OTPRepository
	sessions user.SessionRepository
	sender   user.OTPSender
	tokens   user.TokenService
	config   Config
	nowFn    func() time.Time
}

func NewService(
	users user.UserRepository,
	otps user.OTPRepository,
	sessions user.SessionRepository,
	sender user.OTPSender,
	tokens user.TokenService,
	config Config,
) *Service {
	if config.OTPTTL <= 0 {
		config.OTPTTL = 5 * time.Minute
	}
	if config.OTPMaxAttempts <= 0 {
		config.OTPMaxAttempts = 5
	}
	if config.OTPFor == "" {
		config.OTPFor = user.OTPForMobile
	}

	return &Service{
		users:    users,
		otps:     otps,
		sessions: sessions,
		sender:   sender,
		tokens:   tokens,
		config:   config,
		nowFn:    time.Now,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*TriggerOTPResponse, error) {
	return s.triggerOTP(ctx, req.CountryCode, req.PhoneNumber, req.Platform, req.DeviceID, true)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TriggerOTPResponse, error) {
	return s.triggerOTP(ctx, req.CountryCode, req.PhoneNumber, req.Platform, req.DeviceID, false)
}

func (s *Service) triggerOTP(ctx context.Context, countryCode string, phoneNumber string, platformValue string, deviceID string, _ bool) (*TriggerOTPResponse, error) {
	normalizedCountryCode := strings.TrimSpace(countryCode)
	normalizedPhoneNumber := strings.TrimSpace(phoneNumber)
	platform := user.OTPPlatform(platformValue)

	existingUser, err := s.users.FindByPhone(ctx, normalizedCountryCode, normalizedPhoneNumber)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}
	if existingUser == nil {
		existingUser, err = s.users.Create(ctx, &user.UserEntity{
			CountryCode: normalizedCountryCode,
			PhoneNumber: normalizedPhoneNumber,
		})
		if err != nil {
			return nil, err
		}
	}

	code := generateOTPCode()
	now := s.nowFn()
	_, err = s.otps.Create(ctx, &user.UserOTPEntity{
		UserID:    existingUser.ID,
		OTPCode:   code,
		Platform:  platform,
		OTPFor:    s.config.OTPFor,
		DeviceID:  strings.TrimSpace(deviceID),
		ExpiresAt: now.Add(s.config.OTPTTL),
		CreatedAt: now,
	})
	if err != nil {
		return nil, err
	}

	destination := fmt.Sprintf("%s%s", normalizedCountryCode, normalizedPhoneNumber)
	if err := s.sender.Send(ctx, destination, code); err != nil {
		return nil, err
	}

	return &TriggerOTPResponse{
		Message: "If the account is valid, an OTP has been sent",
	}, nil
}

func (s *Service) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*TokenResponse, error) {
	foundUser, err := s.users.FindByPhone(ctx, strings.TrimSpace(req.CountryCode), strings.TrimSpace(req.PhoneNumber))
	if err != nil {
		return nil, err
	}

	platform := user.OTPPlatform(req.Platform)
	otpRecord, err := s.otps.FindLatestActiveByUserAndPlatform(ctx, foundUser.ID, platform, s.config.OTPFor)
	if err != nil {
		return nil, err
	}

	now := s.nowFn()
	if otpRecord.IsUsed {
		return nil, user.ErrOTPAlreadyUsed
	}
	if now.After(otpRecord.ExpiresAt) {
		return nil, user.ErrOTPExpired
	}
	if otpRecord.AttemptCount >= s.config.OTPMaxAttempts {
		return nil, user.ErrOTPAttemptsExceeded
	}
	if otpRecord.OTPCode != strings.TrimSpace(req.OTPCode) {
		_ = s.otps.IncrementAttempt(ctx, otpRecord.ID)
		return nil, user.ErrInvalidOTP
	}
	if err := s.otps.MarkUsed(ctx, otpRecord.ID, now); err != nil {
		return nil, err
	}

	pair, err := s.tokens.Issue(foundUser.ID)
	if err != nil {
		return nil, err
	}
	refreshExpiry := now.Add(time.Duration(pair.RefreshTokenTTL) * time.Second)
	_, err = s.sessions.Create(ctx, &user.UserSessionEntity{
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

	return &TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.AccessTokenTTL,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error) {
	plain := strings.TrimSpace(req.RefreshToken)
	hashed := s.tokens.HashRefreshToken(plain)

	session, err := s.sessions.FindByRefreshTokenHash(ctx, hashed)
	if err != nil {
		return nil, err
	}
	if session.Revoked {
		return nil, user.ErrSessionRevoked
	}
	now := s.nowFn()
	if session.ExpiresAt != nil && now.After(*session.ExpiresAt) {
		_ = s.sessions.Revoke(ctx, session.ID, "refresh token expired", false, now)
		return nil, user.ErrInvalidRefreshToken
	}

	pair, err := s.tokens.Rotate(session.UserID)
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

func (s *Service) Logout(ctx context.Context, req LogoutRequest) error {
	plain := strings.TrimSpace(req.RefreshToken)
	hashed := s.tokens.HashRefreshToken(plain)

	session, err := s.sessions.FindByRefreshTokenHash(ctx, hashed)
	if err != nil {
		if errors.Is(err, user.ErrInvalidRefreshToken) {
			return nil
		}
		return err
	}
	if session.Revoked {
		return nil
	}
	return s.sessions.Revoke(ctx, session.ID, "user logout", false, s.nowFn())
}

func generateOTPCode() string {
	max := big.NewInt(1000000)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", value.Int64())
}
