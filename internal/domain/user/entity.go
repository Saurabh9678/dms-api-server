package user

import "time"

type UserEntity struct {
	ID          uint64
	Email       string
	PhoneNumber string
	CountryCode string
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type UserRoleEntity struct {
	ID        uint64
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type UserShowroomEntity struct {
	ID         uint64
	UserID     uint64
	ShowroomID uint64
	RoleID     uint64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type OTPPlatform string

const (
	OTPPlatformWeb           OTPPlatform = "web"
	OTPPlatformIOSMobile     OTPPlatform = "ios_mobile"
	OTPPlatformAndroidMobile OTPPlatform = "android_mobile"
	OTPPlatformDesktop       OTPPlatform = "desktop"
)

type OTPFor string

const (
	OTPForMobile OTPFor = "mobile"
	OTPForEmail  OTPFor = "email"
)

type UserOTPEntity struct {
	ID           uint64
	UserID       uint64
	OTPCode      string
	Platform     OTPPlatform
	OTPFor       OTPFor
	DeviceID     string
	AttemptCount int
	ResendCount  int
	IsUsed       bool
	ExpiresAt    time.Time
	CreatedAt    time.Time
	VerifiedAt   *time.Time
}

type UserSessionEntity struct {
	ID               uint64
	UserID           uint64
	Platform         OTPPlatform
	DeviceID         string
	IPAddress        string
	RefreshTokenHash string
	Revoked          bool
	Compromised      bool
	RevokedReason    string
	CreatedAt        time.Time
	LastUsedAt       time.Time
	ExpiresAt        *time.Time
}
