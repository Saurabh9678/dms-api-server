package vehicleimages

import (
	"time"
	"infiour.local/dms-api-server/internal/repository"
	"infiour.local/dms-api-server/internal/repository/users"
	"infiour.local/dms-api-server/internal/repository/vehicles"
)

type VehicleImage struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID uint64 `gorm:"not null" json:"vehicle_id"`
	ImageURL string `gorm:"type:text;not null" json:"image_url"`
	Label VehicleImageLabel `gorm:"type:varchar(255);not null" json:"label"`
	UploadedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"uploaded_at"`
	UploadedBy uint64 `gorm:"not null" json:"uploaded_by"`
	Vehicle vehicles.Vehicle `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	User users.User `gorm:"foreignKey:UploadedBy;references:ID" json:"user"`
	repository.SoftDeleteableModel
}

type VehicleImageLabel string

const (
	VehicleImageLabelFront VehicleImageLabel = "front"
	VehicleImageLabelInterior VehicleImageLabel = "interior"
	VehicleImageLabelExterior VehicleImageLabel = "exterior"
	VehicleImageLabelBack VehicleImageLabel = "back"
	VehicleImageLabelWheel VehicleImageLabel = "wheel"
)

func (VehicleImage) TableName() string {
	return "vehicle_images"
}