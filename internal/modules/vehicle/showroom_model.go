package vehicle

import (
	"infiour.local/dms-api-server/internal/modules/showroom"
	"infiour.local/dms-api-server/pkg/database"
)

type VehicleShowroom struct {
	ID         uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID  uint64            `gorm:"not null" json:"vehicle_id"`
	ShowroomID uint64            `gorm:"not null" json:"showroom_id"`
	Vehicle    Vehicle           `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	Showroom   showroom.Showroom `gorm:"foreignKey:ShowroomID;references:ID" json:"showroom"`
	database.SoftDeleteableModel
}

func (VehicleShowroom) TableName() string {
	return "vehicle_showroom_relations"
}
