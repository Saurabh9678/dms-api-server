package vehiclestatus

import (
	"infiour.local/dms-api-server/internal/repository"
	"infiour.local/dms-api-server/internal/repository/users"
	"infiour.local/dms-api-server/internal/repository/vehicles"
	"time"
)

type VehicleStatus struct {
	ID          uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID   uint64            `gorm:"not null" json:"vehicle_id"`
	Status      VehicleStatusType `gorm:"type:varchar(255);not null" json:"status"`
	Description string            `gorm:"type:text;not null" json:"description"`
	StartedAt   time.Time         `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"started_at"`
	EndedAt     time.Time         `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"ended_at"`
	AddedBy     uint64            `gorm:"not null" json:"added_by"`
	User        users.User        `gorm:"foreignKey:AddedBy;references:ID" json:"user"`
	Vehicle     vehicles.Vehicle  `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	repository.SoftDeleteableModel
}

type VehicleStatusType string

const (
	VehicleStatusTypeGarage       VehicleStatusType = "garage"
	VehicleStatusTypeInspection   VehicleStatusType = "inspection"
	VehicleStatusTypeReadyForSale VehicleStatusType = "ready_for_sale"
	VehicleStatusTypeSold         VehicleStatusType = "sold"
)

func (VehicleStatus) TableName() string {
	return "vehicle_statuses"
}
