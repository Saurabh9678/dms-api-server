package showroom

import (
	"encoding/json"

	"infiour.local/dms-api-server/pkg/database"
)

type Showroom struct {
	ID                  uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                string          `gorm:"type:varchar(255);not null" json:"name"`
	ShowroomLogo        *string         `gorm:"type:text" json:"showroom_logo"`
	ShowroomBanner      *string         `gorm:"type:text" json:"showroom_banner"`
	ShowroomGeolocation json.RawMessage `gorm:"type:json" json:"showroom_geolocation"`
	database.SoftDeleteableModel
}

func (Showroom) TableName() string {
	return "showrooms"
}

// MemberRecord is the raw result of a list-members query joining user_showroom_relations, users, and user_roles.
type MemberRecord struct {
	UserID      uint64 `gorm:"column:user_id"`
	Name        string `gorm:"column:name"`
	CountryCode string `gorm:"column:country_code"`
	PhoneNumber string `gorm:"column:phone_number"`
	Role        string `gorm:"column:role"`
}
