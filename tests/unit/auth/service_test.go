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
	byID    map[uint64]*user.User
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

func (f *fakeUserRepo) FindByID(_ context.Context, userID uint64) (*user.User, error) {
	entity, ok := f.byID[userID]
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
	f.byID[copy.ID] = &copy
	return &copy, nil
}

type fakeOTPRepo struct {
	lastCreated       *auth.UserOTP
	activeByRequestID map[string]*auth.UserOTP
	incrementedID     uint64
	// error injection
	createErr        error
	createErrCount   int  // if > 0, return createErr for this many calls, then succeed
	createCallCount  int
	markUsedErr      error
}

func (f *fakeOTPRepo) Create(_ context.Context, entity *auth.UserOTP) (*auth.UserOTP, error) {
	f.createCallCount++
	if f.createErr != nil {
		if f.createErrCount <= 0 || f.createCallCount <= f.createErrCount {
			return nil, f.createErr
		}
	}
	copy := *entity
	copy.ID = 100
	f.lastCreated = &copy
	f.activeByRequestID[copy.RequestID] = &copy
	return &copy, nil
}

func (f *fakeOTPRepo) FindLatestActiveByRequestIDAndPlatform(_ context.Context, requestID string, _ auth.OTPPlatform, _ auth.OTPFor) (*auth.UserOTP, error) {
	otp, ok := f.activeByRequestID[requestID]
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
	if f.markUsedErr != nil {
		return f.markUsedErr
	}
	for _, item := range f.activeByRequestID {
		if item.ID == otpID {
			item.IsUsed = true
			item.VerifiedAt = &verifiedAt
			return nil
		}
	}
	return nil
}

type fakeSessionRepo struct {
	sessionByHash map[string]*auth.UserSession
	rotatedID     uint64
	revokedID     uint64
	revokedUserID uint64
	revokedOnPlat auth.OTPPlatform
	// error injection
	createErr           error
	rotateErr           error
}

func (f *fakeSessionRepo) Create(_ context.Context, entity *auth.UserSession) (*auth.UserSession, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
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
	if f.rotateErr != nil {
		return f.rotateErr
	}
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

func (f *fakeSessionRepo) RevokeAllByUserIDAndPlatform(_ context.Context, userID uint64, platform auth.OTPPlatform, _ string, _ bool, _ time.Time) error {
	f.revokedUserID = userID
	f.revokedOnPlat = platform
	for _, session := range f.sessionByHash {
		if session.UserID == userID && session.Platform == platform {
			session.Revoked = true
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

func (f *fakeTokenProvider) ParseAccessToken(token string) (uint64, error) {
	if token == "access-token" || token == "new-access-token" {
		return 1, nil
	}
	return 0, auth.ErrInvalidAccessToken
}

func TestRegisterTriggersOTP(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
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
	if len(resp.RequestID) != 8 {
		t.Fatalf("expected request id with length 8, got %q", resp.RequestID)
	}
}

func TestVerifyOTPRejectsInvalidCode(t *testing.T) {
	u := &user.User{ID: 1, CountryCode: "+91", PhoneNumber: "9999999999"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|9999999999": u},
		byID:    map[uint64]*user.User{1: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Ab12Cd34": {ID: 7, UserID: 1, RequestID: "Ab12Cd34", OTPCode: "123456", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}

	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})
	_, err := service.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Ab12Cd34",
		OTPCode:   "000000",
		Platform:  "web",
	})
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
	if otpRepo.incrementedID != 7 {
		t.Fatalf("expected attempt to increment for otp 7, got %d", otpRepo.incrementedID)
	}
}

func TestRefreshAndLogout(t *testing.T) {
	u := &user.User{ID: 1, CountryCode: "+91", PhoneNumber: "9999999999", Name: "Alice"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|9999999999": u},
		byID:    map[uint64]*user.User{1: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Zx90Qw12": {ID: 10, UserID: 1, RequestID: "Zx90Qw12", OTPCode: "123456", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	verifyResp, err := service.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Zx90Qw12",
		OTPCode:   "123456",
		Platform:  "web",
		DeviceID:  "device-1",
	})
	if err != nil {
		t.Fatalf("verify otp should succeed, got %v", err)
	}
	if verifyResp.RefreshToken == "" {
		t.Fatalf("expected refresh token")
	}
	if verifyResp.RequiredName {
		t.Fatalf("expected required_name false for user with name set")
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

	if err := service.Logout(context.Background(), auth.LogoutRequest{AccessToken: "new-access-token", Platform: "web"}); err != nil {
		t.Fatalf("logout should succeed, got %v", err)
	}
	if sessionRepo.revokedUserID == 0 {
		t.Fatalf("expected revoke by user id to be called")
	}
	if sessionRepo.revokedOnPlat != auth.OTPPlatformWeb {
		t.Fatalf("expected revoke on web platform, got %s", sessionRepo.revokedOnPlat)
	}
}

func TestLogoutRejectsInvalidAccessToken(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	err := service.Logout(context.Background(), auth.LogoutRequest{AccessToken: "bad-token", Platform: "web"})
	if !errors.Is(err, auth.ErrInvalidAccessToken) {
		t.Fatalf("expected ErrInvalidAccessToken, got %v", err)
	}
}

func TestVerifyOTPRevokesExistingSessionsForSamePlatform(t *testing.T) {
	u := &user.User{ID: 1, CountryCode: "+91", PhoneNumber: "9999999999"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|9999999999": u},
		byID:    map[uint64]*user.User{1: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Mn34Rt78": {ID: 55, UserID: 1, RequestID: "Mn34Rt78", OTPCode: "123456", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"old-web": {ID: 22, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "old-web"},
			"old-ios": {ID: 23, UserID: 1, Platform: auth.OTPPlatformIOSMobile, RefreshTokenHash: "old-ios"},
		},
	}
	service := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := service.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Mn34Rt78",
		OTPCode:   "123456",
		Platform:  "web",
		DeviceID:  "device-1",
	})
	if err != nil {
		t.Fatalf("verify otp should succeed, got %v", err)
	}

	if sessionRepo.revokedUserID != 1 {
		t.Fatalf("expected revoke call for user 1, got %d", sessionRepo.revokedUserID)
	}
	if sessionRepo.revokedOnPlat != auth.OTPPlatformWeb {
		t.Fatalf("expected revoke call for web platform, got %s", sessionRepo.revokedOnPlat)
	}
	if !sessionRepo.sessionByHash["old-web"].Revoked {
		t.Fatalf("expected old web session to be revoked")
	}
	if sessionRepo.sessionByHash["old-ios"].Revoked {
		t.Fatalf("expected ios session to remain active")
	}
}

func TestVerifyOTPRequiredNameTrueWhenNameEmpty(t *testing.T) {
	u := &user.User{ID: 1, CountryCode: "+91", PhoneNumber: "9999999999", Name: ""}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|9999999999": u},
		byID:    map[uint64]*user.User{1: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Aa11Bb22": {ID: 1, UserID: 1, RequestID: "Aa11Bb22", OTPCode: "111111", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	resp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Aa11Bb22",
		OTPCode:   "111111",
		Platform:  "web",
		DeviceID:  "device-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.RequiredName {
		t.Fatalf("expected required_name true for user with no name")
	}
}

func TestVerifyOTPRequiredNameFalseWhenNameSet(t *testing.T) {
	u := &user.User{ID: 2, CountryCode: "+91", PhoneNumber: "8888888888", Name: "Bob"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|8888888888": u},
		byID:    map[uint64]*user.User{2: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Cc33Dd44": {ID: 2, UserID: 2, RequestID: "Cc33Dd44", OTPCode: "222222", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	resp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Cc33Dd44",
		OTPCode:   "222222",
		Platform:  "web",
		DeviceID:  "device-2",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.RequiredName {
		t.Fatalf("expected required_name false for user with name set")
	}
}

func TestVerifyOTPFailsWhenFindByIDErrors(t *testing.T) {
	u := &user.User{ID: 3, CountryCode: "+91", PhoneNumber: "7777777777"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|7777777777": u},
		byID:    map[uint64]*user.User{},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Ee55Ff66": {ID: 3, UserID: 3, RequestID: "Ee55Ff66", OTPCode: "333333", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Ee55Ff66",
		OTPCode:   "333333",
		Platform:  "web",
		DeviceID:  "device-3",
	})
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound when FindByID fails, got %v", err)
	}
}
