package usersession

import (
	"time"

	"infiour.local/dms-api-server/internal/repository/users"
)

type UserSession struct {
	ID               uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64              `gorm:"not null" json:"user_id"`
	Platform         SessionPlatformType `gorm:"type:platform_type;not null" json:"platform"`
	DeviceID         string              `gorm:"type:varchar(255)" json:"device_id"`
	IPAddress        string              `gorm:"type:varchar(45)" json:"ip_address"`
	RefreshTokenHash string              `gorm:"type:varchar(256)" json:"refresh_token_hash"`
	Revoked          bool                `gorm:"not null;default:false" json:"revoked"`
	Compromised      bool                `gorm:"not null;default:false" json:"compromised"`
	RevokedReason    string              `gorm:"type:varchar(255)" json:"revoked_reason"`
	CreatedAt        time.Time           `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastUsedAt       time.Time           `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"last_used_at"`
	ExpiresAt        *time.Time          `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	User             users.User          `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

type SessionPlatformType string

const (
	SessionPlatformTypeWeb           SessionPlatformType = "web"
	SessionPlatformTypeIOSMobile     SessionPlatformType = "ios_mobile"
	SessionPlatformTypeAndroidMobile SessionPlatformType = "android_mobile"
	SessionPlatformTypeDesktop       SessionPlatformType = "desktop"
)

func (UserSession) TableName() string {
	return "user_sessions"
}
