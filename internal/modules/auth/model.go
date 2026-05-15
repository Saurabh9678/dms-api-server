package auth

import (
	"time"

	"infiour.local/dms-api-server/internal/modules/user"
)

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

type UserOTP struct {
	ID           uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64      `gorm:"not null" json:"user_id"`
	RequestID    string      `gorm:"type:varchar(8);uniqueIndex;not null" json:"request_id"`
	OTPCode      string      `gorm:"type:varchar(6);not null" json:"otp_code"`
	Platform     OTPPlatform `gorm:"type:platform_type;not null" json:"platform"`
	OTPFor       OTPFor      `gorm:"type:otp_for_type;not null" json:"otp_for"`
	DeviceID     string      `gorm:"type:varchar(255)" json:"device_id"`
	AttemptCount int         `gorm:"not null;default:0" json:"attempt_count"`
	ResendCount  int         `gorm:"not null;default:0" json:"resend_count"`
	IsUsed       bool        `gorm:"not null;default:false" json:"is_used"`
	ExpiresAt    time.Time   `gorm:"type:timestamptz;not null" json:"expires_at"`
	CreatedAt    time.Time   `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	VerifiedAt   *time.Time  `gorm:"type:timestamptz" json:"verified_at,omitempty"`
	User         user.User   `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (UserOTP) TableName() string {
	return "user_otps"
}

type UserSession struct {
	ID               uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64      `gorm:"not null" json:"user_id"`
	Platform         OTPPlatform `gorm:"type:platform_type;not null" json:"platform"`
	DeviceID         string      `gorm:"type:varchar(255)" json:"device_id"`
	IPAddress        string      `gorm:"type:varchar(45)" json:"ip_address"`
	RefreshTokenHash string      `gorm:"type:varchar(256)" json:"refresh_token_hash"`
	Revoked          bool        `gorm:"not null;default:false" json:"revoked"`
	Compromised      bool        `gorm:"not null;default:false" json:"compromised"`
	RevokedReason    string      `gorm:"type:varchar(255)" json:"revoked_reason"`
	CreatedAt        time.Time   `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastUsedAt       time.Time   `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"last_used_at"`
	ExpiresAt        *time.Time  `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	User             user.User   `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
