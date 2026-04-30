package showroom

import (
	"encoding/json"

	"infiour.local/dms-api-server/internal/repository"
)

type Showroom struct {
	ID                  uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                string          `gorm:"type:varchar(255);not null" json:"name"`
	ShowroomLogo        string          `gorm:"type:text" json:"showroom_logo"`
	ShowroomGeolocation json.RawMessage `gorm:"type:json" json:"showroom_geolocation"`
	repository.SoftDeleteableModel
}

func (Showroom) TableName() string {
	return "showrooms"
}
