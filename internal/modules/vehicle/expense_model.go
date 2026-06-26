package vehicle

import (
	"time"

	"infiour.local/dms-api-server/pkg/database"
)

type VehicleExpensesType string

const (
	VehicleExpensesTypeRepair        VehicleExpensesType = "repair"
	VehicleExpensesTypeService       VehicleExpensesType = "service"
	VehicleExpensesTypeInsurance     VehicleExpensesType = "insurance"
	VehicleExpensesTypeTax           VehicleExpensesType = "tax"
	VehicleExpensesTypeInspection    VehicleExpensesType = "inspection"
	VehicleExpensesTypeCleaning      VehicleExpensesType = "cleaning"
	VehicleExpensesTypeDocumentation VehicleExpensesType = "documentation"
	VehicleExpensesTypeOther         VehicleExpensesType = "other"
)

type VehicleExpenses struct {
	ID          uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID   uint64              `gorm:"not null" json:"vehicle_id"`
	StatusID    *uint64             `json:"status_id,omitempty"`
	Type        VehicleExpensesType `gorm:"type:varchar(255);not null" json:"type"`
	Amount      float64             `gorm:"type:numeric(10,2);not null" json:"amount"`
	PaidTo      string              `gorm:"type:varchar(255)" json:"paid_to"`
	Description string              `gorm:"type:text" json:"description"`
	Date        time.Time           `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"date"`
	Vehicle     Vehicle             `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	database.SoftDeleteableModel
}

func (VehicleExpenses) TableName() string {
	return "vehicle_expenses"
}
