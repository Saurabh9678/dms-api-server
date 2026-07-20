package subscription

import (
	"github.com/google/uuid"
	"infiour.local/dms-api-server/pkg/database"
)

type PlanFeature struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SubscriptionPlanID    uuid.UUID `gorm:"type:uuid;not null" json:"subscription_plan_id"`
	SubscriptionFeatureID uuid.UUID `gorm:"type:uuid;not null" json:"subscription_feature_id"`
	BoolValue             *bool     `json:"bool_value,omitempty"`
	NumericValue          *float64  `gorm:"type:numeric(18,2)" json:"numeric_value,omitempty"`
	StringValue           *string   `gorm:"type:text" json:"string_value,omitempty"`
	database.TimestampedModel
}

func (PlanFeature) TableName() string {
	return "plan_features"
}
