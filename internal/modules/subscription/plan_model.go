package subscription

import (
	"github.com/google/uuid"
	"infiour.local/dms-api-server/pkg/database"
)

type SubscriptionPlan struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code         string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	Description  *string    `gorm:"type:text" json:"description,omitempty"`
	IsFree       bool       `gorm:"not null;default:false" json:"is_free"`
	IsGlobal     bool       `gorm:"not null;default:true" json:"is_global"`
	CustomerID   *uuid.UUID `gorm:"type:uuid" json:"customer_id,omitempty"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	DisplayOrder int        `gorm:"not null;default:0" json:"display_order"`
	database.TimestampedModel
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}
