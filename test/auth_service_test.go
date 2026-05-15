package diff_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"infiour.local/dms-api-server/internal/application/auth"
	"infiour.local/dms-api-server/internal/domain/user"
)

type fakeUserRepo struct {
	records map[string]*user.UserEntity
	nextID  uint64
}

func (f *fakeUserRepo) FindByPhone(_ context.Context, countryCode string, phoneNumber string) (*user.UserEntity, error) {
	key := countryCode + "|" + phoneNumber
	entity, ok := f.records[key]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return entity, nil
}

func (f *fakeUserRepo) Create(_ context.Context, entity *user.UserEntity) (*user.UserEntity, error) {
	f.nextID++
	copy := *entity
	copy.ID = f.nextID
	key := copy.CountryCode + "|" + copy.PhoneNumber
	f.records[key] = &copy
	return &copy, nil
}

type fakeOTPRepo struct {
	lastCreated     *user.UserOTPEntity
	activeByUserID  map[uint64]*user.UserOTPEntity
	updatedAsUsedID uint64
	incrementedID   uint64
}

func (f *fakeOTPRepo) Create(_ context.Context, entity *user.UserOTPEntity) (*user.UserOTPEntity, error) {
	copy := *entity
	copy.ID = 100
	f.lastCreated = &copy
	f.activeByUserID[copy.UserID] = &copy
	return &copy, nil
}

func (f *fakeOTPRepo) FindLatestActiveByUserAndPlatform(_ context.Context, userID uint64, _ user.OTPPlatform, _ user.OTPFor) (*user.UserOTPEntity, error) {
	otp, ok := f.activeByUserID[userID]
	if !ok {
		return nil, user.ErrInvalidOTP
	}
	return otp, nil
}

func (f *fakeOTPRepo) IncrementAttempt(_ context.Context, otpID uint64) error {
	f.incrementedID = otpID
	return nil
}

func (f *fakeOTPRepo) MarkUsed(_ context.Context, otpID uint64, verifiedAt time.Time) error {
	f.updatedAsUsedID = otpID
	item := f.activeByUserID[1]
	if item != nil {
		item.IsUsed = true
		item.VerifiedAt = &verifiedAt
	}
	return nil
}

type fakeSessionRepo struct {
	lastCreated   *user.UserSessionEntity
	sessionByHash map[string]*user.UserSessionEntity
	rotatedID     uint64
	revokedID     uint64
}

func (f *fakeSessionRepo) Create(_ context.Context, entity *user.UserSessionEntity) (*user.UserSessionEntity, error) {
	copy := *entity
	copy.ID = 501
	f.lastCreated = &copy
	f.sessionByHash[copy.RefreshTokenHash] = &copy
	return &copy, nil
}

func (f *fakeSessionRepo) FindByRefreshTokenHash(_ context.Context, refreshTokenHash string) (*user.UserSessionEntity, error) {
	session, ok := f.sessionByHash[refreshTokenHash]
	if !ok {
		return nil, user.ErrInvalidRefreshToken
	}
	return session, nil
}

func (f *fakeSessionRepo) RotateRefreshToken(_ context.Context, sessionID uint64, refreshTokenHash string, expiresAt time.Time, lastUsedAt time.Time) error {
	f.rotatedID = sessionID
	for _, session := range f.sessionByHash {
		if session.ID == sessionID {
			delete(f.sessionByHash, session.RefreshTokenHash)
			session.RefreshTokenHash = refreshTokenHash
			session.ExpiresAt = &expiresAt
			session.LastUsedAt = lastUsedAt
			f.sessionByHash[refreshTokenHash] = session
			return nil
		}
	}
	return nil
}

func (f *fakeSessionRepo) Revoke(_ context.Context, sessionID uint64, _ string, _ bool, revokedAt time.Time) error {
	f.revokedID = sessionID
	for _, session := range f.sessionByHash {
		if session.ID == sessionID {
			session.Revoked = true
			session.LastUsedAt = revokedAt
			return nil
		}
	}
	return nil
}

type fakeOTPSender struct {
	lastDestination string
	lastCode        string
}

func (f *fakeOTPSender) Send(_ context.Context, destination string, code string) error {
	f.lastDestination = destination
	f.lastCode = code
	return nil
}

type fakeTokenService struct{}

func (f *fakeTokenService) Issue(_ uint64) (*user.TokenPair, error) {
	return &user.TokenPair{
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		AccessTokenTTL:   900,
		RefreshTokenTTL:  604800,
		RefreshTokenHash: "refresh-token-hash",
	}, nil
}

func (f *fakeTokenService) Rotate(_ uint64) (*user.TokenPair, error) {
	return &user.TokenPair{
		AccessToken:      "new-access-token",
		RefreshToken:     "new-refresh-token",
		AccessTokenTTL:   900,
		RefreshTokenTTL:  604800,
		RefreshTokenHash: "new-refresh-token-hash",
	}, nil
}

func (f *fakeTokenService) HashRefreshToken(token string) string {
	if token == "refresh-token" {
		return "refresh-token-hash"
	}
	if token == "new-refresh-token" {
		return "new-refresh-token-hash"
	}
	return "unknown"
}

func TestRegisterTriggersOTP(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.UserEntity{}}
	otpRepo := &fakeOTPRepo{activeByUserID: map[uint64]*user.UserOTPEntity{}}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*user.UserSessionEntity{}}
	sender := &fakeOTPSender{}
	tokens := &fakeTokenService{}

	service := auth.NewService(userRepo, otpRepo, sessionRepo, sender, tokens, auth.Config{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 5,
		OTPFor:         user.OTPForMobile,
	})

	resp, err := service.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91",
		PhoneNumber: "9999999999",
		Platform:    "web",
		DeviceID:    "device-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || resp.Message == "" {
		t.Fatalf("expected trigger response")
	}
	if sender.lastDestination != "+919999999999" {
		t.Fatalf("expected destination to be populated, got %q", sender.lastDestination)
	}
	if otpRepo.lastCreated == nil {
		t.Fatalf("expected otp to be created")
	}
}

func TestVerifyOTPRejectsInvalidCode(t *testing.T) {
	userRepo := &fakeUserRepo{
		records: map[string]*user.UserEntity{
			"+91|9999999999": {ID: 1, CountryCode: "+91", PhoneNumber: "9999999999"},
		},
	}
	otpRepo := &fakeOTPRepo{
		activeByUserID: map[uint64]*user.UserOTPEntity{
			1: {ID: 7, UserID: 1, OTPCode: "123456", Platform: user.OTPPlatformWeb, OTPFor: user.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*user.UserSessionEntity{}}

	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPSender{}, &fakeTokenService{}, auth.Config{OTPFor: user.OTPForMobile})
	_, err := service.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		CountryCode: "+91",
		PhoneNumber: "9999999999",
		OTPCode:     "000000",
		Platform:    "web",
	})
	if !errors.Is(err, user.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
	if otpRepo.incrementedID != 7 {
		t.Fatalf("expected attempt to increment for otp 7, got %d", otpRepo.incrementedID)
	}
}

func TestRefreshAndLogout(t *testing.T) {
	userRepo := &fakeUserRepo{
		records: map[string]*user.UserEntity{
			"+91|9999999999": {ID: 1, CountryCode: "+91", PhoneNumber: "9999999999"},
		},
	}
	otpRepo := &fakeOTPRepo{
		activeByUserID: map[uint64]*user.UserOTPEntity{
			1: {ID: 10, UserID: 1, OTPCode: "123456", Platform: user.OTPPlatformWeb, OTPFor: user.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*user.UserSessionEntity{}}
	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPSender{}, &fakeTokenService{}, auth.Config{OTPFor: user.OTPForMobile})

	verifyResp, err := service.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		CountryCode: "+91",
		PhoneNumber: "9999999999",
		OTPCode:     "123456",
		Platform:    "web",
		DeviceID:    "device-1",
	})
	if err != nil {
		t.Fatalf("verify otp should succeed, got %v", err)
	}
	if verifyResp.RefreshToken == "" {
		t.Fatalf("expected refresh token")
	}

	refreshResp, err := service.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("refresh should succeed, got %v", err)
	}
	if refreshResp.AccessToken != "new-access-token" {
		t.Fatalf("unexpected access token: %s", refreshResp.AccessToken)
	}
	if sessionRepo.rotatedID == 0 {
		t.Fatalf("expected rotate to be called")
	}

	if err := service.Logout(context.Background(), auth.LogoutRequest{RefreshToken: "new-refresh-token"}); err != nil {
		t.Fatalf("logout should succeed, got %v", err)
	}
	if sessionRepo.revokedID == 0 {
		t.Fatalf("expected revoke to be called")
	}
}
