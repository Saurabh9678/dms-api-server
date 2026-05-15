package vehicledocument

import (
	"infiour.local/dms-api-server/internal/repository"
	"infiour.local/dms-api-server/internal/repository/users"
	"infiour.local/dms-api-server/internal/repository/vehicles"
	"time"
)

type VehicleDocument struct {
	ID           uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID    uint64              `gorm:"not null" json:"vehicle_id"`
	DocumentType VehicleDocumentType `gorm:"type:varchar(255);not null" json:"document_type"`
	DocumentURL  string              `gorm:"type:text;not null" json:"document_url"`
	ValidFrom    time.Time           `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"valid_from"`
	ValidTill    time.Time           `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"valid_till"`
	Remarks      string              `gorm:"type:text;not null" json:"remarks"`
	UploadedAt   time.Time           `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"uploaded_at"`
	UploadedBy   uint64              `gorm:"not null" json:"uploaded_by"`
	User         users.User          `gorm:"foreignKey:UploadedBy;references:ID" json:"user"`
	Vehicle      vehicles.Vehicle    `gorm:"foreignKey:VehicleID;references:ID" json:"vehicle"`
	repository.SoftDeleteableModel
}

type VehicleDocumentType string

const (
	VehicleDocumentTypeRegistrationCertificate VehicleDocumentType = "registration_certificate"
	VehicleDocumentTypeInsurance               VehicleDocumentType = "insurance"
	VehicleDocumentTypePollution               VehicleDocumentType = "pollution"
)

func (VehicleDocument) TableName() string {
	return "vehicle_documents"
}
