package vehicle

import (
	"time"

	"infiour.local/dms-api-server/pkg/database"
)

type VehiclePricing struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID   uint64    `gorm:"not null" json:"vehicle_id"`
	BuyingPrice float64   `gorm:"type:numeric(10,2);not null" json:"buying_price"`
	BuyingDate  time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"buying_date"`
	PriceTag    float64   `gorm:"type:numeric(10,2);not null" json:"price_tag"`
	TaggedAt    time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"tagged_at"`
	Currency    Currency  `gorm:"type:varchar(10);not null;default: 'inr'" json:"currency"`
	Remarks     string    `gorm:"type:text;not null" json:"remarks"`
	Vehicle     Vehicle   `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	database.SoftDeleteableModel
}

type Currency string

const (
	CurrencyINR Currency = "inr"
	CurrencyUSD Currency = "usd"
)

func (VehiclePricing) TableName() string {
	return "vehicle_pricing"
}
