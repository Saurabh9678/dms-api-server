package customervehiclesale

import (
	"time"
	"infiour.local/dms-api-server/internal/repository"
)

type CustomerVehicleSale struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerID uint64 `gorm:"not null" json:"customer_id"`
	VehicleID uint64 `gorm:"not null" json:"vehicle_id"`
	SalePrice float64 `gorm:"type:numeric(10,2);not null" json:"sale_price"`
	SaleDate time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"sale_date"`
	PaymentMode PaymentMode `gorm:"type:varchar(255);not null" json:"payment_mode"`
	ReceiptUrl string `gorm:"type:text;not null" json:"receipt_url"`
	Remarks string `gorm:"type:text;not null" json:"remarks"`
	repository.SoftDeleteableModel
}

type PaymentMode string

const (
	PaymentModeCash PaymentMode = "cash"
	PaymentModeCheque PaymentMode = "cheque"
	PaymentModeBankTransfer PaymentMode = "bank_transfer"
	PaymentModeOnline PaymentMode = "online"
	PaymentModeCredit PaymentMode = "credit"
	PaymentModeDebit PaymentMode = "debit"
	PaymentModeOther PaymentMode = "other"
)

func (CustomerVehicleSale) TableName() string {
	return "customer_vehicle_sales"
}