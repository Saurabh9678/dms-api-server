package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/internal/modules/auth"
	"infiour.local/dms-api-server/internal/modules/user"
	otpprovider "infiour.local/dms-api-server/internal/providers/otp"
	tokenprovider "infiour.local/dms-api-server/internal/providers/token"
	"infiour.local/dms-api-server/pkg/config"
)

type fakeUserRepo struct {
	records map[string]*user.User
	nextID  uint64
}

func (f *fakeUserRepo) FindByPhone(_ context.Context, countryCode string, phoneNumber string) (*user.User, error) {
	key := countryCode + "|" + phoneNumber
	entity, ok := f.records[key]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return entity, nil
}

func (f *fakeUserRepo) Create(_ context.Context, entity *user.User) (*user.User, error) {
	f.nextID++
	copy := *entity
	copy.ID = f.nextID
	key := copy.CountryCode + "|" + copy.PhoneNumber
	f.records[key] = &copy
	return &copy, nil
}

type fakeOTPRepo struct {
	lastCreated    *auth.UserOTP
	activeByUserID map[uint64]*auth.UserOTP
	incrementedID  uint64
}

func (f *fakeOTPRepo) Create(_ context.Context, entity *auth.UserOTP) (*auth.UserOTP, error) {
	copy := *entity
	copy.ID = 100
	f.lastCreated = &copy
	f.activeByUserID[copy.UserID] = &copy
	return &copy, nil
}

func (f *fakeOTPRepo) FindLatestActiveByUserAndPlatform(_ context.Context, userID uint64, _ auth.OTPPlatform, _ auth.OTPFor) (*auth.UserOTP, error) {
	otp, ok := f.activeByUserID[userID]
	if !ok {
		return nil, auth.ErrInvalidOTP
	}
	return otp, nil
}

func (f *fakeOTPRepo) IncrementAttempt(_ context.Context, otpID uint64) error {
	f.incrementedID = otpID
	return nil
}

func (f *fakeOTPRepo) MarkUsed(_ context.Context, otpID uint64, verifiedAt time.Time) error {
	item := f.activeByUserID[1]
	if item != nil {
		item.IsUsed = true
		item.VerifiedAt = &verifiedAt
	}
	return nil
}

type fakeSessionRepo struct {
	sessionByHash map[string]*auth.UserSession
	rotatedID     uint64
	revokedID     uint64
}

func (f *fakeSessionRepo) Create(_ context.Context, entity *auth.UserSession) (*auth.UserSession, error) {
	copy := *entity
	copy.ID = 501
	f.sessionByHash[copy.RefreshTokenHash] = &copy
	return &copy, nil
}

func (f *fakeSessionRepo) FindByRefreshTokenHash(_ context.Context, refreshTokenHash string) (*auth.UserSession, error) {
	session, ok := f.sessionByHash[refreshTokenHash]
	if !ok {
		return nil, auth.ErrInvalidRefreshToken
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

type fakeOTPProvider struct {
	lastDestination string
	lastCode        string
}

var _ otpprovider.Provider = (*fakeOTPProvider)(nil)

func (f *fakeOTPProvider) Send(_ context.Context, destination string, code string) error {
	f.lastDestination = destination
	f.lastCode = code
	return nil
}

type fakeTokenProvider struct{}

var _ tokenprovider.Provider = (*fakeTokenProvider)(nil)

func (f *fakeTokenProvider) Issue(_ uint64) (*tokenprovider.TokenPair, error) {
	return &tokenprovider.TokenPair{
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		AccessTokenTTL:   900,
		RefreshTokenTTL:  604800,
		RefreshTokenHash: "refresh-token-hash",
	}, nil
}

func (f *fakeTokenProvider) Rotate(_ uint64) (*tokenprovider.TokenPair, error) {
	return &tokenprovider.TokenPair{
		AccessToken:      "new-access-token",
		RefreshToken:     "new-refresh-token",
		AccessTokenTTL:   900,
		RefreshTokenTTL:  604800,
		RefreshTokenHash: "new-refresh-token-hash",
	}, nil
}

func (f *fakeTokenProvider) HashRefreshToken(token string) string {
	if token == "refresh-token" {
		return "refresh-token-hash"
	}
	if token == "new-refresh-token" {
		return "new-refresh-token-hash"
	}
	return "unknown"
}

func TestRegisterTriggersOTP(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByUserID: map[uint64]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	sender := &fakeOTPProvider{}
	tokens := &fakeTokenProvider{}

	service := auth.NewService(userRepo, otpRepo, sessionRepo, sender, tokens, config.AuthConfig{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 5,
	}, &gorm.DB{})

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
		records: map[string]*user.User{
			"+91|9999999999": {ID: 1, CountryCode: "+91", PhoneNumber: "9999999999"},
		},
	}
	otpRepo := &fakeOTPRepo{
		activeByUserID: map[uint64]*auth.UserOTP{
			1: {ID: 7, UserID: 1, OTPCode: "123456", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}

	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})
	_, err := service.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		CountryCode: "+91",
		PhoneNumber: "9999999999",
		OTPCode:     "000000",
		Platform:    "web",
	})
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
	if otpRepo.incrementedID != 7 {
		t.Fatalf("expected attempt to increment for otp 7, got %d", otpRepo.incrementedID)
	}
}

func TestRefreshAndLogout(t *testing.T) {
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{
			"+91|9999999999": {ID: 1, CountryCode: "+91", PhoneNumber: "9999999999"},
		},
	}
	otpRepo := &fakeOTPRepo{
		activeByUserID: map[uint64]*auth.UserOTP{
			1: {ID: 10, UserID: 1, OTPCode: "123456", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

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
