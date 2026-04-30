package customer

import (
	"infiour.local/dms-api-server/internal/repository"
)

type Customer struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	FirstName string `gorm:"type:varchar(255);not null" json:"first_name"`
	LastName string `gorm:"type:varchar(255);not null" json:"last_name"`
	Email string `gorm:"type:varchar(255);not null" json:"email"`
	PhoneNumber string `gorm:"type:varchar(255);not null" json:"phone_number"`
	AltPhoneNumber string `gorm:"type:varchar(255);not null" json:"alt_phone_number"`
	Address string `gorm:"type:text;not null" json:"address"`
	City string `gorm:"type:varchar(255);not null" json:"city"`
	State string `gorm:"type:varchar(255);not null" json:"state"`
	Pincode string `gorm:"type:varchar(255);not null" json:"pincode"`
	IdProofType string `gorm:"type:varchar(255);not null" json:"id_proof_type"`
	IdProofNumber string `gorm:"type:varchar(255);not null" json:"id_proof_number"`
	IdProofUrl string `gorm:"type:text;not null" json:"id_proof_url"`
	repository.SoftDeleteableModel
}

func (Customer) TableName() string {
	return "customers"
}