package userotp

import (
	"time"

	"infiour.local/dms-api-server/internal/repository/users"
)

type UserOTP struct {
	ID           uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64       `gorm:"not null" json:"user_id"`
	OTPCode      string       `gorm:"type:varchar(6);not null" json:"otp_code"`
	Platform     PlatformType `gorm:"type:platform_type;not null" json:"platform"`
	OTPFor       OTPForType   `gorm:"type:otp_for_type;not null" json:"otp_for"`
	DeviceID     string       `gorm:"type:varchar(255)" json:"device_id"`
	AttemptCount int          `gorm:"not null;default:0" json:"attempt_count"`
	ResendCount  int          `gorm:"not null;default:0" json:"resend_count"`
	IsUsed       bool         `gorm:"not null;default:false" json:"is_used"`
	ExpiresAt    time.Time    `gorm:"type:timestamptz;not null" json:"expires_at"`
	CreatedAt    time.Time    `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	VerifiedAt   *time.Time   `gorm:"type:timestamptz" json:"verified_at,omitempty"`
	User         users.User   `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

type PlatformType string

const (
	PlatformTypeWeb           PlatformType = "web"
	PlatformTypeIOSMobile     PlatformType = "ios_mobile"
	PlatformTypeAndroidMobile PlatformType = "android_mobile"
	PlatformTypeDesktop       PlatformType = "desktop"
)

type OTPForType string

const (
	OTPForTypeMobile OTPForType = "mobile"
	OTPForTypeEmail  OTPForType = "email"
)

func (UserOTP) TableName() string {
	return "user_otps"
}
