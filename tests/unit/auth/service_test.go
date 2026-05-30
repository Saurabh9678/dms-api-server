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
	records       map[string]*user.User
	byID          map[uint64]*user.User
	nextID        uint64
	findPhoneErr  error // non-nil: return this error (overrides ErrUserNotFound)
	createUserErr error // non-nil: return this error from Create
}

func (f *fakeUserRepo) FindByPhone(_ context.Context, countryCode string, phoneNumber string) (*user.User, error) {
	if f.findPhoneErr != nil {
		return nil, f.findPhoneErr
	}
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
	if f.createUserErr != nil {
		return nil, f.createUserErr
	}
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
	createErr       error
	createErrCount  int // if > 0, return createErr for this many calls, then succeed
	createCallCount int
	markUsedErr     error
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
	createErr    error
	rotateErr    error
	revokeAllErr error
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
	if f.revokeAllErr != nil {
		return f.revokeAllErr
	}
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
	sendErr         error
}

var _ otpprovider.Provider = (*fakeOTPProvider)(nil)

func (f *fakeOTPProvider) Send(_ context.Context, destination string, code string) error {
	if f.sendErr != nil {
		return f.sendErr
	}
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

// ---------------------------------------------------------------------------
// fakeTokenProviderWithError — for testing Rotate/Issue error paths
// ---------------------------------------------------------------------------

type fakeTokenProviderWithError struct {
	issueErr  error
	rotateErr error
}

var _ tokenprovider.Provider = (*fakeTokenProviderWithError)(nil)

func (f *fakeTokenProviderWithError) Issue(_ uint64) (*tokenprovider.TokenPair, error) {
	if f.issueErr != nil {
		return nil, f.issueErr
	}
	return &tokenprovider.TokenPair{
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		AccessTokenTTL:   900,
		RefreshTokenTTL:  604800,
		RefreshTokenHash: "refresh-token-hash",
	}, nil
}

func (f *fakeTokenProviderWithError) Rotate(_ uint64) (*tokenprovider.TokenPair, error) {
	if f.rotateErr != nil {
		return nil, f.rotateErr
	}
	return &tokenprovider.TokenPair{
		AccessToken:      "new-access-token",
		RefreshToken:     "new-refresh-token",
		AccessTokenTTL:   900,
		RefreshTokenTTL:  604800,
		RefreshTokenHash: "new-refresh-token-hash",
	}, nil
}

func (f *fakeTokenProviderWithError) HashRefreshToken(token string) string {
	if token == "refresh-token" {
		return "refresh-token-hash"
	}
	return "unknown"
}

func (f *fakeTokenProviderWithError) ParseAccessToken(token string) (uint64, error) {
	if token == "access-token" {
		return 1, nil
	}
	return 0, auth.ErrInvalidAccessToken
}

// ---------------------------------------------------------------------------
// Login tests
// ---------------------------------------------------------------------------

func TestLoginTriggersOTP(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	sender := &fakeOTPProvider{}
	tokens := &fakeTokenProvider{}

	svc := auth.NewService(userRepo, otpRepo, sessionRepo, sender, tokens, config.AuthConfig{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 5,
	}, &gorm.DB{})

	resp, err := svc.Login(context.Background(), auth.LoginRequest{
		CountryCode: "+91",
		PhoneNumber: "8888888888",
		Platform:    "ios_mobile",
		DeviceID:    "device-login",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || resp.Message == "" {
		t.Fatalf("expected trigger response")
	}
	if sender.lastDestination != "+918888888888" {
		t.Fatalf("expected destination +918888888888, got %q", sender.lastDestination)
	}
	if len(resp.RequestID) != 8 {
		t.Fatalf("expected request id with length 8, got %q", resp.RequestID)
	}
}

// ---------------------------------------------------------------------------
// triggerOTP retry / exhaustion tests
// ---------------------------------------------------------------------------

func TestTriggerOTP_DuplicateRequestIDRetry(t *testing.T) {
	// Create returns ErrDuplicatedKey on first call, succeeds on second
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		createErr:         gorm.ErrDuplicatedKey,
		createErrCount:    1, // fail first call, succeed from second onward
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	sender := &fakeOTPProvider{}

	svc := auth.NewService(userRepo, otpRepo, sessionRepo, sender, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL: 5 * time.Minute,
	}, &gorm.DB{})

	resp, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91",
		PhoneNumber: "1234567890",
		Platform:    "web",
		DeviceID:    "device-x",
	})
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if otpRepo.createCallCount < 2 {
		t.Fatalf("expected at least 2 Create calls, got %d", otpRepo.createCallCount)
	}
}

func TestTriggerOTP_AllRetriesExhausted(t *testing.T) {
	// Create always returns ErrDuplicatedKey — all 5 retries fail
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		createErr:         gorm.ErrDuplicatedKey,
		createErrCount:    0, // 0 means "always error"
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}

	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL: 5 * time.Minute,
	}, &gorm.DB{})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91",
		PhoneNumber: "1111111111",
		Platform:    "web",
		DeviceID:    "device-y",
	})
	if err == nil {
		t.Fatal("expected error when all retries exhausted")
	}
	if otpRepo.createCallCount != 5 {
		t.Fatalf("expected exactly 5 Create calls (requestIDGenerateRetries), got %d", otpRepo.createCallCount)
	}
}

func TestTriggerOTP_NonDuplicateCreateError(t *testing.T) {
	// Create returns a non-duplicate error — should bail immediately
	nonDupErr := errors.New("some db constraint error")
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		createErr:         nonDupErr,
		createErrCount:    0,
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}

	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL: 5 * time.Minute,
	}, &gorm.DB{})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91",
		PhoneNumber: "2222222222",
		Platform:    "web",
		DeviceID:    "device-z",
	})
	if !errors.Is(err, nonDupErr) {
		t.Fatalf("expected nonDupErr, got %v", err)
	}
	if otpRepo.createCallCount != 1 {
		t.Fatalf("expected exactly 1 Create call (bail on non-duplicate), got %d", otpRepo.createCallCount)
	}
}

// ---------------------------------------------------------------------------
// VerifyOTP error paths
// ---------------------------------------------------------------------------

func TestVerifyOTP_MarkUsedError(t *testing.T) {
	markErr := errors.New("mark used failed")
	u := &user.User{ID: 5, CountryCode: "+91", PhoneNumber: "5555555555"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|5555555555": u},
		byID:    map[uint64]*user.User{5: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"MarkUsed1": {ID: 50, UserID: 5, RequestID: "MarkUsed1", OTPCode: "555555", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
		markUsedErr: markErr,
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "MarkUsed1",
		OTPCode:   "555555",
		Platform:  "web",
		DeviceID:  "device-5",
	})
	if !errors.Is(err, markErr) {
		t.Fatalf("expected markErr, got %v", err)
	}
}

func TestVerifyOTP_SessionCreateError(t *testing.T) {
	createErr := errors.New("session create failed")
	u := &user.User{ID: 6, CountryCode: "+91", PhoneNumber: "6666666666"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|6666666666": u},
		byID:    map[uint64]*user.User{6: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"SesCreate": {ID: 60, UserID: 6, RequestID: "SesCreate", OTPCode: "666666", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{},
		createErr:     createErr,
	}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "SesCreate",
		OTPCode:   "666666",
		Platform:  "web",
		DeviceID:  "device-6",
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected createErr, got %v", err)
	}
}

func TestVerifyOTP_RevokeAllError(t *testing.T) {
	revokeErr := errors.New("revoke all failed")
	u := &user.User{ID: 7, CountryCode: "+91", PhoneNumber: "7654321098"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|7654321098": u},
		byID:    map[uint64]*user.User{7: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"RevAll01": {ID: 70, UserID: 7, RequestID: "RevAll01", OTPCode: "777777", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{},
		revokeAllErr:  revokeErr,
	}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "RevAll01",
		OTPCode:   "777777",
		Platform:  "web",
		DeviceID:  "device-7",
	})
	if !errors.Is(err, revokeErr) {
		t.Fatalf("expected revokeErr, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// RefreshToken error paths
// ---------------------------------------------------------------------------

func TestRefreshToken_RevokedSession(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	expiry := time.Now().Add(7 * 24 * time.Hour)
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 1, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: true, ExpiresAt: &expiry},
		},
	}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, auth.ErrSessionRevoked) {
		t.Fatalf("expected ErrSessionRevoked, got %v", err)
	}
}

func TestRefreshToken_ExpiredSession(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	expiry := time.Now().Add(-1 * time.Hour) // past expiry
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 2, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: &expiry},
		},
	}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken for expired session, got %v", err)
	}
}

func TestRefreshToken_TokenProviderRotateError(t *testing.T) {
	rotErr := errors.New("rotate failed")
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	expiry := time.Now().Add(7 * 24 * time.Hour)
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 3, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: &expiry},
		},
	}
	tokenProv := &fakeTokenProviderWithError{rotateErr: rotErr}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, tokenProv, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, rotErr) {
		t.Fatalf("expected rotateErr, got %v", err)
	}
}

func TestRefreshToken_RotateRefreshTokenError(t *testing.T) {
	rotateErr := errors.New("db rotate failed")
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	expiry := time.Now().Add(7 * 24 * time.Hour)
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 4, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: &expiry},
		},
		rotateErr: rotateErr,
	}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, rotateErr) {
		t.Fatalf("expected rotateErr, got %v", err)
	}
}

func TestRefreshToken_NilExpiresAt(t *testing.T) {
	// ExpiresAt is nil — the expiry check is skipped, token should rotate successfully
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 5, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: nil},
		},
	}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	resp, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("expected success for nil ExpiresAt session, got %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatal("expected non-empty token response")
	}
}

// ---------------------------------------------------------------------------
// triggerOTP: users.FindByPhone non-ErrUserNotFound error
// ---------------------------------------------------------------------------

func TestTriggerOTP_FindByPhoneError(t *testing.T) {
	dbErr := errors.New("connection refused")
	userRepo := &fakeUserRepo{
		records:      map[string]*user.User{},
		byID:         map[uint64]*user.User{},
		findPhoneErr: dbErr,
	}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL: 5 * time.Minute,
	}, &gorm.DB{})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91",
		PhoneNumber: "3333333333",
		Platform:    "web",
		DeviceID:    "device-err",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got %v", err)
	}
}

func TestTriggerOTP_CreateUserError(t *testing.T) {
	createErr := errors.New("user create failed")
	userRepo := &fakeUserRepo{
		records:       map[string]*user.User{},
		byID:          map[uint64]*user.User{},
		createUserErr: createErr,
	}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL: 5 * time.Minute,
	}, &gorm.DB{})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91",
		PhoneNumber: "4444444444",
		Platform:    "web",
		DeviceID:    "device-cu",
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected createErr, got %v", err)
	}
}

func TestTriggerOTP_SendError(t *testing.T) {
	sendErr := errors.New("sms gateway down")
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	sender := &fakeOTPProvider{sendErr: sendErr}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, sender, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL: 5 * time.Minute,
	}, &gorm.DB{})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91",
		PhoneNumber: "5544332211",
		Platform:    "web",
		DeviceID:    "device-send",
	})
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected sendErr, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// VerifyOTP: OTP already used, expired, attempts exceeded, Issue error
// ---------------------------------------------------------------------------

func TestVerifyOTP_OTPAlreadyUsed(t *testing.T) {
	u := &user.User{ID: 10, CountryCode: "+91", PhoneNumber: "1010101010"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|1010101010": u},
		byID:    map[uint64]*user.User{10: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"UsedOTP1": {ID: 100, UserID: 10, RequestID: "UsedOTP1", OTPCode: "101010", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, IsUsed: true, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "UsedOTP1",
		OTPCode:   "101010",
		Platform:  "web",
		DeviceID:  "d",
	})
	if !errors.Is(err, auth.ErrOTPAlreadyUsed) {
		t.Fatalf("expected ErrOTPAlreadyUsed, got %v", err)
	}
}

func TestVerifyOTP_OTPExpired(t *testing.T) {
	u := &user.User{ID: 11, CountryCode: "+91", PhoneNumber: "1111222233"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|1111222233": u},
		byID:    map[uint64]*user.User{11: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"ExpOTP01": {ID: 110, UserID: 11, RequestID: "ExpOTP01", OTPCode: "111222", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, IsUsed: false, ExpiresAt: time.Now().Add(-1 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "ExpOTP01",
		OTPCode:   "111222",
		Platform:  "web",
		DeviceID:  "d",
	})
	if !errors.Is(err, auth.ErrOTPExpired) {
		t.Fatalf("expected ErrOTPExpired, got %v", err)
	}
}

func TestVerifyOTP_AttemptsExceeded(t *testing.T) {
	u := &user.User{ID: 12, CountryCode: "+91", PhoneNumber: "1222333444"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|1222333444": u},
		byID:    map[uint64]*user.User{12: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"MaxAtt01": {ID: 120, UserID: 12, RequestID: "MaxAtt01", OTPCode: "121212", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, IsUsed: false, AttemptCount: 5, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{OTPMaxAttempts: 5}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "MaxAtt01",
		OTPCode:   "121212",
		Platform:  "web",
		DeviceID:  "d",
	})
	if !errors.Is(err, auth.ErrOTPAttemptsExceeded) {
		t.Fatalf("expected ErrOTPAttemptsExceeded, got %v", err)
	}
}

func TestVerifyOTP_IssueTokenError(t *testing.T) {
	issueErr := errors.New("token issue failed")
	u := &user.User{ID: 13, CountryCode: "+91", PhoneNumber: "1333444555"}
	userRepo := &fakeUserRepo{
		records: map[string]*user.User{"+91|1333444555": u},
		byID:    map[uint64]*user.User{13: u},
	}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"IssErr01": {ID: 130, UserID: 13, RequestID: "IssErr01", OTPCode: "133133", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, IsUsed: false, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	tokenProv := &fakeTokenProviderWithError{issueErr: issueErr}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, tokenProv, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "IssErr01",
		OTPCode:   "133133",
		Platform:  "web",
		DeviceID:  "d",
	})
	if !errors.Is(err, issueErr) {
		t.Fatalf("expected issueErr, got %v", err)
	}
}

func TestVerifyOTP_OTPNotFound(t *testing.T) {
	// FindLatestActiveByRequestIDAndPlatform returns error (ErrInvalidOTP) — requestID not in map
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{}, // empty — requestID won't be found
	}
	sessionRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "NotFound",
		OTPCode:   "000000",
		Platform:  "web",
		DeviceID:  "d",
	})
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP for missing requestID, got %v", err)
	}
}

func TestRefreshToken_SessionNotFound(t *testing.T) {
	// FindByRefreshTokenHash returns error — token hash not in map
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, byID: map[uint64]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sessionRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{}, // empty — hash won't be found
	}
	svc := auth.NewService(userRepo, otpRepo, sessionRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{}, &gorm.DB{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "nonexistent-token"})
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken for unknown token, got %v", err)
	}
}
