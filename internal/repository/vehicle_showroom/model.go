package vehicleshowroom

import (
	"infiour.local/dms-api-server/internal/repository"
	"infiour.local/dms-api-server/internal/repository/vehicles"
	"infiour.local/dms-api-server/internal/repository/showroom"
)

type VehicleShowroom struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID uint64 `gorm:"not null" json:"vehicle_id"`
	ShowroomID uint64 `gorm:"not null" json:"showroom_id"`
	Vehicle vehicles.Vehicle `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	Showroom showroom.Showroom `gorm:"foreignKey:ShowroomID;references:ID" json:"showroom"`
	repository.SoftDeleteableModel
}

func (VehicleShowroom) TableName() string {
	return "vehicle_showroom_relations"
}