package vehicleexpenses

import (
	"time"

	"infiour.local/dms-api-server/internal/repository"
	"infiour.local/dms-api-server/internal/repository/vehicle_status"
	"infiour.local/dms-api-server/internal/repository/vehicles"
)

type VehicleExpenses struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID uint64 `gorm:"not null" json:"vehicle_id"`
	StatusID uint64 `gorm:"not null" json:"status_id"`
	Type VehicleExpensesType `gorm:"type:varchar(255);not null" json:"type"`
	Amount float64 `gorm:"type:numeric(10,2);not null" json:"amount"`
	PaidTo string `gorm:"type:varchar(255);not null" json:"paid_to"`
	Description string `gorm:"type:text;not null" json:"description"`
	Date time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"date"`
	Vehicle vehicles.Vehicle `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	Status vehiclestatus.VehicleStatus `gorm:"foreignKey:StatusID;references:ID" json:"status"`
	repository.SoftDeleteableModel
}	

type VehicleExpensesType string

func (VehicleExpenses) TableName() string {
	return "vehicle_expenses"
}