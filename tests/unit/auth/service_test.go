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

// ---------------------------------------------------------------------------
// fakeUserRepo — implements userRepo interface (FindByPhone + Create only)
// ---------------------------------------------------------------------------

type fakeUserRepo struct {
	records       map[string]*user.User
	nextID        uint64
	findPhoneErr  error // non-nil: always return this error on FindByPhone
	createUserErr error // non-nil: return this error from Create

	// findPhoneNotFoundFirstTime: if true, the first FindByPhone call returns
	// ErrUserNotFound regardless of records content. Used to simulate concurrent
	// user creation races (first call "misses", second call "finds" the winner).
	findPhoneNotFoundFirstTime bool
	findPhoneCallCount         int
	// findPhoneErrOnSecondCall: if non-nil, returned as the error on the 2nd FindByPhone call.
	// Used to cover the re-fetch-after-ErrDuplicatedKey error path.
	findPhoneErrOnSecondCall error
}

func (f *fakeUserRepo) FindByPhone(_ context.Context, countryCode string, phoneNumber string) (*user.User, error) {
	f.findPhoneCallCount++
	if f.findPhoneErr != nil {
		return nil, f.findPhoneErr
	}
	if f.findPhoneNotFoundFirstTime && f.findPhoneCallCount == 1 {
		return nil, user.ErrUserNotFound
	}
	if f.findPhoneErrOnSecondCall != nil && f.findPhoneCallCount == 2 {
		return nil, f.findPhoneErrOnSecondCall
	}
	key := countryCode + "|" + phoneNumber
	entity, ok := f.records[key]
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
	cp := *entity
	cp.ID = f.nextID
	key := cp.CountryCode + "|" + cp.PhoneNumber
	f.records[key] = &cp
	return &cp, nil
}

// ---------------------------------------------------------------------------
// fakeOTPRepo
// ---------------------------------------------------------------------------

type fakeOTPRepo struct {
	lastCreated       *auth.UserOTP
	activeByRequestID map[string]*auth.UserOTP
	incrementedID     uint64
	// OTP create error injection
	createErr       error
	createErrCount  int // if > 0, return createErr for this many calls, then succeed
	createCallCount int
	// MarkUsed error injection
	markUsedErr error
	// Rate limiting stubs
	latestByPhone    *auth.UserOTP
	latestByPhoneErr error
	countByPhone     int64
	countByPhoneErr  error
}

func (f *fakeOTPRepo) Create(_ context.Context, entity *auth.UserOTP) (*auth.UserOTP, error) {
	f.createCallCount++
	if f.createErr != nil {
		if f.createErrCount <= 0 || f.createCallCount <= f.createErrCount {
			return nil, f.createErr
		}
	}
	cp := *entity
	cp.ID = 100
	f.lastCreated = &cp
	f.activeByRequestID[cp.RequestID] = &cp
	return &cp, nil
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

func (f *fakeOTPRepo) FindLatestByPhone(_ context.Context, _, _ string) (*auth.UserOTP, error) {
	return f.latestByPhone, f.latestByPhoneErr
}

func (f *fakeOTPRepo) CountRecentByPhone(_ context.Context, _, _ string, _ time.Time) (int64, error) {
	if f.countByPhoneErr != nil {
		return 0, f.countByPhoneErr
	}
	return f.countByPhone, nil
}

// ---------------------------------------------------------------------------
// fakeSessionRepo
// ---------------------------------------------------------------------------

type fakeSessionRepo struct {
	sessionByHash map[string]*auth.UserSession
	rotatedID     uint64
	revokedID     uint64
	revokedUserID uint64
	revokedOnPlat auth.OTPPlatform
	createErr     error
	rotateErr     error
	revokeAllErr  error
}

func (f *fakeSessionRepo) Create(_ context.Context, entity *auth.UserSession) (*auth.UserSession, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	cp := *entity
	cp.ID = 501
	f.sessionByHash[cp.RefreshTokenHash] = &cp
	return &cp, nil
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

// ---------------------------------------------------------------------------
// fakeOTPProvider
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// fakeTokenProvider
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// fakeTokenProviderWithError
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
// helpers
// ---------------------------------------------------------------------------

func newSvc(userR *fakeUserRepo, otpR *fakeOTPRepo, sesR *fakeSessionRepo, sender *fakeOTPProvider, tokens tokenprovider.Provider, cfg config.AuthConfig) auth.Service {
	return auth.NewService(userR, otpR, sesR, sender, tokens, cfg, &gorm.DB{})
}

func otpRecord(requestID string, code string, platform auth.OTPPlatform) *auth.UserOTP {
	return &auth.UserOTP{
		ID:          100,
		CountryCode: "+91",
		PhoneNumber: "9999999999",
		RequestID:   requestID,
		OTPCode:     code,
		Platform:    platform,
		OTPFor:      auth.OTPForMobile,
		ExpiresAt:   time.Now().Add(2 * time.Minute),
	}
}

// ---------------------------------------------------------------------------
// SendOTP / Register / Login — triggerOTP path
// ---------------------------------------------------------------------------

func TestRegisterTriggersOTP(t *testing.T) {
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sender := &fakeOTPProvider{}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	resp, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91", PhoneNumber: "9999999999", Platform: "web", DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || resp.Message != "OTP sent successfully" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if sender.lastDestination != "+919999999999" {
		t.Fatalf("expected destination +919999999999, got %q", sender.lastDestination)
	}
	if otpRepo.lastCreated == nil {
		t.Fatal("expected OTP to be created")
	}
	if otpRepo.lastCreated.CountryCode != "+91" || otpRepo.lastCreated.PhoneNumber != "9999999999" {
		t.Fatalf("OTP record missing phone data: %+v", otpRepo.lastCreated)
	}
	if len(resp.RequestID) != 8 {
		t.Fatalf("expected requestId length 8, got %q", resp.RequestID)
	}
}

func TestLoginTriggersOTP(t *testing.T) {
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sender := &fakeOTPProvider{}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	resp, err := svc.Login(context.Background(), auth.LoginRequest{
		CountryCode: "+91", PhoneNumber: "8888888888", Platform: "ios_mobile", DeviceID: "device-login",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sender.lastDestination != "+918888888888" {
		t.Fatalf("expected destination +918888888888, got %q", sender.lastDestination)
	}
	if len(resp.RequestID) != 8 {
		t.Fatalf("expected requestId length 8, got %q", resp.RequestID)
	}
}

func TestSendOTPDelegatesToTriggerOTP(t *testing.T) {
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sender := &fakeOTPProvider{}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	resp, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "7777777777", Platform: "android_mobile", DeviceID: "device-otp",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Message != "OTP sent successfully" {
		t.Fatalf("unexpected message: %q", resp.Message)
	}
	if sender.lastDestination != "+917777777777" {
		t.Fatalf("expected destination +917777777777, got %q", sender.lastDestination)
	}
}

func TestTriggerOTP_NoUserTableAccess(t *testing.T) {
	// triggerOTP must not call FindByPhone or Create on userRepo
	userRepo := &fakeUserRepo{
		records:      map[string]*user.User{},
		findPhoneErr: errors.New("user repo should not be called"),
	}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	svc := newSvc(userRepo, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "5555555555", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("triggerOTP must not call userRepo; got error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// OTP rate limiting
// ---------------------------------------------------------------------------

func TestTriggerOTP_CooldownEnforced(t *testing.T) {
	recent := &auth.UserOTP{CreatedAt: time.Now().Add(-30 * time.Second)}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		latestByPhone:     recent,
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL:             5 * time.Minute,
		OTPCooldownSeconds: 60,
	})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "9999999999", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, auth.ErrOTPCooldown) {
		t.Fatalf("expected ErrOTPCooldown, got %v", err)
	}
}

func TestTriggerOTP_CooldownNotEnforcedAfterExpiry(t *testing.T) {
	old := &auth.UserOTP{CreatedAt: time.Now().Add(-2 * time.Minute)}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		latestByPhone:     old,
	}
	sender := &fakeOTPProvider{}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL:             5 * time.Minute,
		OTPCooldownSeconds: 60,
		OTPMaxDailySends:   10,
	})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "9999999999", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error after cooldown expiry, got %v", err)
	}
}

func TestTriggerOTP_NoPreviousOTPSkipsCooldown(t *testing.T) {
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		latestByPhone:     nil, // no previous OTP
	}
	sender := &fakeOTPProvider{}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL:             5 * time.Minute,
		OTPCooldownSeconds: 60,
		OTPMaxDailySends:   10,
	})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "1111111111", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error for new phone, got %v", err)
	}
}

func TestTriggerOTP_DailyCapEnforced(t *testing.T) {
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		latestByPhone:     nil,
		countByPhone:      10, // at daily limit
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL:           5 * time.Minute,
		OTPMaxDailySends: 10,
	})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "9999999999", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, auth.ErrOTPRateLimitExceeded) {
		t.Fatalf("expected ErrOTPRateLimitExceeded, got %v", err)
	}
}

func TestTriggerOTP_DailyCapNotEnforcedBelowLimit(t *testing.T) {
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		latestByPhone:     nil,
		countByPhone:      9, // below limit
	}
	sender := &fakeOTPProvider{}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{
		OTPTTL:           5 * time.Minute,
		OTPMaxDailySends: 10,
	})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "9999999999", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error below daily cap, got %v", err)
	}
}

func TestTriggerOTP_FindLatestByPhoneError(t *testing.T) {
	dbErr := errors.New("db connection refused")
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		latestByPhoneErr:  dbErr,
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "9999999999", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got %v", err)
	}
}

func TestTriggerOTP_CountRecentByPhoneError(t *testing.T) {
	dbErr := errors.New("count query failed")
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		latestByPhone:     nil,
		countByPhoneErr:   dbErr,
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	_, err := svc.SendOTP(context.Background(), auth.SendOTPRequest{
		CountryCode: "+91", PhoneNumber: "9999999999", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// triggerOTP: requestID retry / exhaustion
// ---------------------------------------------------------------------------

func TestTriggerOTP_DuplicateRequestIDRetry(t *testing.T) {
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		createErr:         gorm.ErrDuplicatedKey,
		createErrCount:    1,
	}
	sender := &fakeOTPProvider{}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	resp, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91", PhoneNumber: "1234567890", Platform: "web", DeviceID: "device-x",
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
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		createErr:         gorm.ErrDuplicatedKey,
		createErrCount:    0,
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91", PhoneNumber: "1111111111", Platform: "web", DeviceID: "device-y",
	})
	if err == nil {
		t.Fatal("expected error when all retries exhausted")
	}
	if otpRepo.createCallCount != 5 {
		t.Fatalf("expected exactly 5 Create calls, got %d", otpRepo.createCallCount)
	}
}

func TestTriggerOTP_NonDuplicateCreateError(t *testing.T) {
	nonDupErr := errors.New("some db constraint error")
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{},
		createErr:         nonDupErr,
		createErrCount:    0,
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91", PhoneNumber: "2222222222", Platform: "web", DeviceID: "device-z",
	})
	if !errors.Is(err, nonDupErr) {
		t.Fatalf("expected nonDupErr, got %v", err)
	}
	if otpRepo.createCallCount != 1 {
		t.Fatalf("expected exactly 1 Create call, got %d", otpRepo.createCallCount)
	}
}

func TestTriggerOTP_SendError(t *testing.T) {
	sendErr := errors.New("sms gateway down")
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	sender := &fakeOTPProvider{sendErr: sendErr}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, sender, &fakeTokenProvider{}, config.AuthConfig{OTPTTL: 5 * time.Minute})

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		CountryCode: "+91", PhoneNumber: "5544332211", Platform: "web", DeviceID: "device-send",
	})
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected sendErr, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// VerifyOTP: validation paths (before user lookup)
// ---------------------------------------------------------------------------

func TestVerifyOTPRejectsInvalidCode(t *testing.T) {
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Ab12Cd34": otpRecord("Ab12Cd34", "123456", auth.OTPPlatformWeb),
		},
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Ab12Cd34", OTPCode: "000000", Platform: "web",
	})
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
	if otpRepo.incrementedID != 100 {
		t.Fatalf("expected attempt increment for id 100, got %d", otpRepo.incrementedID)
	}
}

func TestVerifyOTP_OTPAlreadyUsed(t *testing.T) {
	rec := otpRecord("UsedOTP1", "101010", auth.OTPPlatformWeb)
	rec.IsUsed = true
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"UsedOTP1": rec}}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "UsedOTP1", OTPCode: "101010", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, auth.ErrOTPAlreadyUsed) {
		t.Fatalf("expected ErrOTPAlreadyUsed, got %v", err)
	}
}

func TestVerifyOTP_OTPExpired(t *testing.T) {
	rec := otpRecord("ExpOTP01", "111222", auth.OTPPlatformWeb)
	rec.ExpiresAt = time.Now().Add(-1 * time.Minute)
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"ExpOTP01": rec}}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "ExpOTP01", OTPCode: "111222", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, auth.ErrOTPExpired) {
		t.Fatalf("expected ErrOTPExpired, got %v", err)
	}
}

func TestVerifyOTP_AttemptsExceeded(t *testing.T) {
	rec := otpRecord("MaxAtt01", "121212", auth.OTPPlatformWeb)
	rec.AttemptCount = 5
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"MaxAtt01": rec}}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{OTPMaxAttempts: 5})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "MaxAtt01", OTPCode: "121212", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, auth.ErrOTPAttemptsExceeded) {
		t.Fatalf("expected ErrOTPAttemptsExceeded, got %v", err)
	}
}

func TestVerifyOTP_OTPNotFound(t *testing.T) {
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "NotFound", OTPCode: "000000", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP for missing requestID, got %v", err)
	}
}

func TestVerifyOTP_MarkUsedError(t *testing.T) {
	markErr := errors.New("mark used failed")
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{"MarkUsed1": otpRecord("MarkUsed1", "555555", auth.OTPPlatformWeb)},
		markUsedErr:       markErr,
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "MarkUsed1", OTPCode: "555555", Platform: "web", DeviceID: "device-5",
	})
	if !errors.Is(err, markErr) {
		t.Fatalf("expected markErr, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// VerifyOTP: user find-or-create (after OTP validation)
// ---------------------------------------------------------------------------

func TestVerifyOTP_ExistingUser(t *testing.T) {
	// Phone already exists in users table → user found, not created
	u := &user.User{ID: 1, CountryCode: "+91", PhoneNumber: "9999999999", Name: "Alice"}
	userRepo := &fakeUserRepo{records: map[string]*user.User{"+91|9999999999": u}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"Exist001": otpRecord("Exist001", "123456", auth.OTPPlatformWeb)}}
	sesRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := newSvc(userRepo, otpRepo, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	resp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Exist001", OTPCode: "123456", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.RequiredName {
		t.Fatal("expected required_name false for user with name set")
	}
	if userRepo.findPhoneCallCount != 1 {
		t.Fatalf("expected 1 FindByPhone call, got %d", userRepo.findPhoneCallCount)
	}
}

func TestVerifyOTP_NewUserCreated(t *testing.T) {
	// Phone not in users table → user created
	userRepo := &fakeUserRepo{records: map[string]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"New00001": otpRecord("New00001", "654321", auth.OTPPlatformWeb)}}
	sesRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := newSvc(userRepo, otpRepo, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	resp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "New00001", OTPCode: "654321", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.RequiredName {
		t.Fatal("expected required_name true for newly created user (no name)")
	}
	if userRepo.nextID != 1 {
		t.Fatalf("expected user to have been created, nextID=%d", userRepo.nextID)
	}
}

func TestVerifyOTP_ConcurrentCreationRace(t *testing.T) {
	// Simulates: FindByPhone → not found (1st call), Create → ErrDuplicatedKey (concurrent race),
	// FindByPhone → user found (2nd call). Users table is the canonical source.
	winner := &user.User{ID: 99, CountryCode: "+91", PhoneNumber: "9999999999", Name: ""}
	userRepo := &fakeUserRepo{
		records:                    map[string]*user.User{"+91|9999999999": winner},
		findPhoneNotFoundFirstTime: true, // 1st call returns not found
		createUserErr:              gorm.ErrDuplicatedKey,
	}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"Race0001": otpRecord("Race0001", "111111", auth.OTPPlatformWeb)}}
	sesRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := newSvc(userRepo, otpRepo, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	resp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Race0001", OTPCode: "111111", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error after race recovery, got %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatal("expected valid token response")
	}
	if userRepo.findPhoneCallCount != 2 {
		t.Fatalf("expected 2 FindByPhone calls (initial not-found + re-fetch), got %d", userRepo.findPhoneCallCount)
	}
}

func TestVerifyOTP_FindByPhoneError(t *testing.T) {
	dbErr := errors.New("users db error")
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, findPhoneErr: dbErr}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"PhErr001": otpRecord("PhErr001", "222222", auth.OTPPlatformWeb)}}
	svc := newSvc(userRepo, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "PhErr001", OTPCode: "222222", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr from FindByPhone, got %v", err)
	}
}

func TestVerifyOTP_CreateUserError(t *testing.T) {
	createErr := errors.New("user create failed")
	userRepo := &fakeUserRepo{records: map[string]*user.User{}, createUserErr: createErr}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"CrErr001": otpRecord("CrErr001", "333333", auth.OTPPlatformWeb)}}
	svc := newSvc(userRepo, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "CrErr001", OTPCode: "333333", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected createErr, got %v", err)
	}
}

func TestVerifyOTP_RaceRefetchError(t *testing.T) {
	// Covers: Create returns ErrDuplicatedKey → re-fetch FindByPhone also returns an error
	refetchErr := errors.New("refetch failed")
	userRepo := &fakeUserRepo{
		records:                    map[string]*user.User{},
		findPhoneNotFoundFirstTime: true, // 1st FindByPhone returns ErrUserNotFound
		createUserErr:              gorm.ErrDuplicatedKey,
		findPhoneErrOnSecondCall:   refetchErr, // 2nd FindByPhone (re-fetch) returns error
	}

	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"RaceErr1": otpRecord("RaceErr1", "444444", auth.OTPPlatformWeb)}}
	svc := newSvc(userRepo, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "RaceErr1", OTPCode: "444444", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, refetchErr) {
		t.Fatalf("expected refetchErr, got %v", err)
	}
	if userRepo.findPhoneCallCount != 2 {
		t.Fatalf("expected 2 FindByPhone calls, got %d", userRepo.findPhoneCallCount)
	}
}

func TestVerifyOTP_RequiredNameTrueForNewUser(t *testing.T) {
	userRepo := &fakeUserRepo{records: map[string]*user.User{}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"NewName1": otpRecord("NewName1", "555666", auth.OTPPlatformWeb)}}
	svc := newSvc(userRepo, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	resp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "NewName1", OTPCode: "555666", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.RequiredName {
		t.Fatal("expected required_name true for new user (no name)")
	}
}

func TestVerifyOTP_RequiredNameFalseWhenNameSet(t *testing.T) {
	u := &user.User{ID: 2, CountryCode: "+91", PhoneNumber: "8888888888", Name: "Bob"}
	userRepo := &fakeUserRepo{records: map[string]*user.User{"+91|8888888888": u}}
	rec := &auth.UserOTP{
		ID: 100, CountryCode: "+91", PhoneNumber: "8888888888",
		RequestID: "Named001", OTPCode: "222222", Platform: auth.OTPPlatformWeb,
		OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute),
	}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{"Named001": rec}}
	svc := newSvc(userRepo, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	resp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Named001", OTPCode: "222222", Platform: "web", DeviceID: "d",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.RequiredName {
		t.Fatal("expected required_name false for user with name set")
	}
}

// ---------------------------------------------------------------------------
// VerifyOTP: session/token paths (after user lookup)
// ---------------------------------------------------------------------------

func TestVerifyOTP_IssueTokenError(t *testing.T) {
	issueErr := errors.New("token issue failed")
	u := &user.User{ID: 13, CountryCode: "+91", PhoneNumber: "1333444555"}
	userRepo := &fakeUserRepo{records: map[string]*user.User{"+91|1333444555": u}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"IssErr01": {ID: 130, CountryCode: "+91", PhoneNumber: "1333444555", RequestID: "IssErr01", OTPCode: "133133", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	tokenProv := &fakeTokenProviderWithError{issueErr: issueErr}
	svc := newSvc(userRepo, otpRepo, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, tokenProv, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "IssErr01", OTPCode: "133133", Platform: "web", DeviceID: "d",
	})
	if !errors.Is(err, issueErr) {
		t.Fatalf("expected issueErr, got %v", err)
	}
}

func TestVerifyOTP_RevokeAllError(t *testing.T) {
	revokeErr := errors.New("revoke all failed")
	u := &user.User{ID: 7, CountryCode: "+91", PhoneNumber: "7654321098"}
	userRepo := &fakeUserRepo{records: map[string]*user.User{"+91|7654321098": u}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"RevAll01": {ID: 70, CountryCode: "+91", PhoneNumber: "7654321098", RequestID: "RevAll01", OTPCode: "777777", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sesRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}, revokeAllErr: revokeErr}
	svc := newSvc(userRepo, otpRepo, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "RevAll01", OTPCode: "777777", Platform: "web", DeviceID: "device-7",
	})
	if !errors.Is(err, revokeErr) {
		t.Fatalf("expected revokeErr, got %v", err)
	}
}

func TestVerifyOTP_SessionCreateError(t *testing.T) {
	createErr := errors.New("session create failed")
	u := &user.User{ID: 6, CountryCode: "+91", PhoneNumber: "6666666666"}
	userRepo := &fakeUserRepo{records: map[string]*user.User{"+91|6666666666": u}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"SesCreate": {ID: 60, CountryCode: "+91", PhoneNumber: "6666666666", RequestID: "SesCreate", OTPCode: "666666", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
		},
	}
	sesRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}, createErr: createErr}
	svc := newSvc(userRepo, otpRepo, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "SesCreate", OTPCode: "666666", Platform: "web", DeviceID: "device-6",
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected createErr, got %v", err)
	}
}

func TestVerifyOTPRevokesExistingSessionsForSamePlatform(t *testing.T) {
	u := &user.User{ID: 1, CountryCode: "+91", PhoneNumber: "9999999999"}
	userRepo := &fakeUserRepo{records: map[string]*user.User{"+91|9999999999": u}}
	otpRepo := &fakeOTPRepo{
		activeByRequestID: map[string]*auth.UserOTP{
			"Mn34Rt78": otpRecord("Mn34Rt78", "123456", auth.OTPPlatformWeb),
		},
	}
	sesRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"old-web": {ID: 22, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "old-web"},
			"old-ios": {ID: 23, UserID: 1, Platform: auth.OTPPlatformIOSMobile, RefreshTokenHash: "old-ios"},
		},
	}
	svc := newSvc(userRepo, otpRepo, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Mn34Rt78", OTPCode: "123456", Platform: "web", DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sesRepo.revokedUserID != 1 || sesRepo.revokedOnPlat != auth.OTPPlatformWeb {
		t.Fatalf("expected revoke for user 1 on web, got userID=%d plat=%s", sesRepo.revokedUserID, sesRepo.revokedOnPlat)
	}
	if !sesRepo.sessionByHash["old-web"].Revoked {
		t.Fatal("expected old web session to be revoked")
	}
	if sesRepo.sessionByHash["old-ios"].Revoked {
		t.Fatal("expected ios session to remain active")
	}
}

// ---------------------------------------------------------------------------
// RefreshToken + Logout
// ---------------------------------------------------------------------------

func TestRefreshAndLogout(t *testing.T) {
	u := &user.User{ID: 1, CountryCode: "+91", PhoneNumber: "9999999999", Name: "Alice"}
	userRepo := &fakeUserRepo{records: map[string]*user.User{"+91|9999999999": u}}
	otpRepo := &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{
		"Zx90Qw12": {ID: 10, CountryCode: "+91", PhoneNumber: "9999999999", RequestID: "Zx90Qw12", OTPCode: "123456", Platform: auth.OTPPlatformWeb, OTPFor: auth.OTPForMobile, ExpiresAt: time.Now().Add(2 * time.Minute)},
	}}
	sesRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := newSvc(userRepo, otpRepo, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	verifyResp, err := svc.VerifyOTP(context.Background(), auth.VerifyOTPRequest{
		RequestID: "Zx90Qw12", OTPCode: "123456", Platform: "web", DeviceID: "device-1",
	})
	if err != nil {
		t.Fatalf("verify otp should succeed, got %v", err)
	}
	if verifyResp.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	if verifyResp.RequiredName {
		t.Fatal("expected required_name false for user with name set")
	}

	refreshResp, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("refresh should succeed, got %v", err)
	}
	if refreshResp.AccessToken != "new-access-token" {
		t.Fatalf("unexpected access token: %s", refreshResp.AccessToken)
	}
	if sesRepo.rotatedID == 0 {
		t.Fatal("expected rotate to be called")
	}

	if err := svc.Logout(context.Background(), auth.LogoutRequest{AccessToken: "new-access-token", Platform: "web"}); err != nil {
		t.Fatalf("logout should succeed, got %v", err)
	}
	if sesRepo.revokedUserID == 0 {
		t.Fatal("expected revoke by user id to be called")
	}
	if sesRepo.revokedOnPlat != auth.OTPPlatformWeb {
		t.Fatalf("expected revoke on web platform, got %s", sesRepo.revokedOnPlat)
	}
}

func TestLogoutRejectsInvalidAccessToken(t *testing.T) {
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}, &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	err := svc.Logout(context.Background(), auth.LogoutRequest{AccessToken: "bad-token", Platform: "web"})
	if !errors.Is(err, auth.ErrInvalidAccessToken) {
		t.Fatalf("expected ErrInvalidAccessToken, got %v", err)
	}
}

func TestRefreshToken_RevokedSession(t *testing.T) {
	expiry := time.Now().Add(7 * 24 * time.Hour)
	sesRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 1, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: true, ExpiresAt: &expiry},
		},
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, auth.ErrSessionRevoked) {
		t.Fatalf("expected ErrSessionRevoked, got %v", err)
	}
}

func TestRefreshToken_ExpiredSession(t *testing.T) {
	expiry := time.Now().Add(-1 * time.Hour)
	sesRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 2, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: &expiry},
		},
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshToken_TokenProviderRotateError(t *testing.T) {
	rotErr := errors.New("rotate failed")
	expiry := time.Now().Add(7 * 24 * time.Hour)
	sesRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 3, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: &expiry},
		},
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}, sesRepo, &fakeOTPProvider{}, &fakeTokenProviderWithError{rotateErr: rotErr}, config.AuthConfig{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, rotErr) {
		t.Fatalf("expected rotateErr, got %v", err)
	}
}

func TestRefreshToken_RotateRefreshTokenError(t *testing.T) {
	rotateErr := errors.New("db rotate failed")
	expiry := time.Now().Add(7 * 24 * time.Hour)
	sesRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 4, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: &expiry},
		},
		rotateErr: rotateErr,
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if !errors.Is(err, rotateErr) {
		t.Fatalf("expected rotateErr, got %v", err)
	}
}

func TestRefreshToken_NilExpiresAt(t *testing.T) {
	sesRepo := &fakeSessionRepo{
		sessionByHash: map[string]*auth.UserSession{
			"refresh-token-hash": {ID: 5, UserID: 1, Platform: auth.OTPPlatformWeb, RefreshTokenHash: "refresh-token-hash", Revoked: false, ExpiresAt: nil},
		},
	}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	resp, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("expected success for nil ExpiresAt session, got %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatal("expected non-empty token response")
	}
}

func TestRefreshToken_SessionNotFound(t *testing.T) {
	sesRepo := &fakeSessionRepo{sessionByHash: map[string]*auth.UserSession{}}
	svc := newSvc(&fakeUserRepo{records: map[string]*user.User{}}, &fakeOTPRepo{activeByRequestID: map[string]*auth.UserOTP{}}, sesRepo, &fakeOTPProvider{}, &fakeTokenProvider{}, config.AuthConfig{})

	_, err := svc.RefreshToken(context.Background(), auth.RefreshTokenRequest{RefreshToken: "nonexistent-token"})
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}
