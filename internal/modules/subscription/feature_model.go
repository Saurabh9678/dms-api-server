package subscription

import (
	"github.com/google/uuid"
	"infiour.local/dms-api-server/pkg/database"
)

type FeatureValueType string

const (
	FeatureValueTypeBoolean FeatureValueType = "BOOLEAN"
	FeatureValueTypeNumber  FeatureValueType = "NUMBER"
	FeatureValueTypeString  FeatureValueType = "STRING"
)

type SubscriptionFeature struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Key         string           `gorm:"column:key;type:varchar(100);not null" json:"key"`
	Name        string           `gorm:"type:varchar(100);not null" json:"name"`
	Description *string          `gorm:"type:text" json:"description,omitempty"`
	ValueType   FeatureValueType `gorm:"type:feature_value_type;not null" json:"value_type"`
	Category    *string          `gorm:"type:varchar(100)" json:"category,omitempty"`
	database.SoftDeleteableModel
}

func (SubscriptionFeature) TableName() string {
	return "subscription_features"
}
