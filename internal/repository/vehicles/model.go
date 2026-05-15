package vehicles

import "infiour.local/dms-api-server/internal/repository"

type Vehicle struct {
	ID                 uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleType        VehicleType      `gorm:"type:varchar(255);not null" json:"vehicle_type"`
	Manufacturer       string           `gorm:"type:varchar(255);not null" json:"manufacturer"`
	Model              string           `gorm:"type:varchar(255);not null" json:"model"`
	Variant            string           `gorm:"type:varchar(255);not null" json:"variant"`
	Color              string           `gorm:"type:varchar(255);not null" json:"color"`
	YearOfManufacture  int              `gorm:"type:int;not null" json:"year_of_manufacture"`
	RTOCode            string           `gorm:"type:varchar(255);not null" json:"rto_code"`
	RegistrationNumber string           `gorm:"type:varchar(255);not null" json:"registration_number"`
	RegistrationState  string           `gorm:"type:varchar(255);not null" json:"registration_state"`
	UsageKM            int              `gorm:"type:int;not null" json:"usage_km"`
	FuelType           FuelType         `gorm:"type:varchar(255);not null" json:"fuel_type"`
	TransmissionType   TransmissionType `gorm:"type:varchar(255);not null" json:"transmission_type"`
	repository.SoftDeleteableModel
}

type VehicleType string

const (
	VehicleTypeBike   VehicleType = "bike"
	VehicleTypeCar    VehicleType = "car"
	VehicleTypeScooty VehicleType = "scooty"
)

type FuelType string

const (
	FuelTypePetrol FuelType = "petrol"
	FuelTypeDiesel FuelType = "diesel"
	FuelTypeEV     FuelType = "ev"
)

type TransmissionType string

const (
	TransmissionTypeManual    TransmissionType = "manual"
	TransmissionTypeAutomatic TransmissionType = "automatic"
)

func (Vehicle) TableName() string {
	return "vehicles"
}
