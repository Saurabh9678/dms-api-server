package subscription

import (
	"github.com/google/uuid"
	"infiour.local/dms-api-server/pkg/database"
)

type BillingCycle string

const (
	BillingCycleMonthly    BillingCycle = "MONTHLY"
	BillingCycleQuarterly  BillingCycle = "QUARTERLY"
	BillingCycleHalfYearly BillingCycle = "HALF_YEARLY"
	BillingCycleYearly     BillingCycle = "YEARLY"
)

type SubscriptionPricing struct {
	ID                 uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SubscriptionPlanID uuid.UUID    `gorm:"type:uuid;not null" json:"subscription_plan_id"`
	BillingCycle       BillingCycle `gorm:"type:billing_cycle_type;not null" json:"billing_cycle"`
	DurationDays       int          `gorm:"not null" json:"duration_days"`
	Price              float64      `gorm:"type:numeric(10,2);not null" json:"price"`
	Currency           string       `gorm:"type:char(3);not null;default:INR" json:"currency"`
	IsActive           bool         `gorm:"not null;default:true" json:"is_active"`
	database.TimestampedModel
}

func (SubscriptionPricing) TableName() string {
	return "subscription_pricing"
}
