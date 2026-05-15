package vehicle

import (
	"time"

	"infiour.local/dms-api-server/internal/modules/user"
	"infiour.local/dms-api-server/pkg/database"
)

type VehicleImage struct {
	ID         uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID  uint64            `gorm:"not null" json:"vehicle_id"`
	ImageURL   string            `gorm:"type:text;not null" json:"image_url"`
	Label      VehicleImageLabel `gorm:"type:varchar(255);not null" json:"label"`
	UploadedAt time.Time         `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"uploaded_at"`
	UploadedBy uint64            `gorm:"not null" json:"uploaded_by"`
	Vehicle    Vehicle           `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	User       user.User         `gorm:"foreignKey:UploadedBy;references:ID" json:"user"`
	database.SoftDeleteableModel
}

type VehicleImageLabel string

const (
	VehicleImageLabelFront    VehicleImageLabel = "front"
	VehicleImageLabelInterior VehicleImageLabel = "interior"
	VehicleImageLabelExterior VehicleImageLabel = "exterior"
	VehicleImageLabelBack     VehicleImageLabel = "back"
	VehicleImageLabelWheel    VehicleImageLabel = "wheel"
)

func (VehicleImage) TableName() string {
	return "vehicle_images"
}
