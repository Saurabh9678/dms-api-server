package subscription

import (
	"time"

	"github.com/google/uuid"
	"infiour.local/dms-api-server/pkg/database"
)

type SubscriptionStatus string

const (
	SubscriptionStatusPending   SubscriptionStatus = "PENDING"
	SubscriptionStatusActive    SubscriptionStatus = "ACTIVE"
	SubscriptionStatusExpired   SubscriptionStatus = "EXPIRED"
	SubscriptionStatusCancelled SubscriptionStatus = "CANCELLED"
	SubscriptionStatusSuspended SubscriptionStatus = "SUSPENDED"
)

type DealerSubscription struct {
	ID                    uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DealerID              uint64             `gorm:"not null" json:"dealer_id"`
	SubscriptionPlanID    uuid.UUID          `gorm:"type:uuid;not null" json:"subscription_plan_id"`
	SubscriptionPricingID uuid.UUID          `gorm:"type:uuid;not null" json:"subscription_pricing_id"`
	Status                SubscriptionStatus `gorm:"type:subscription_status_type;not null" json:"status"`
	StartsAt              time.Time          `gorm:"type:timestamptz;not null" json:"starts_at"`
	ExpiresAt             time.Time          `gorm:"type:timestamptz;not null" json:"expires_at"`
	AutoRenew             bool               `gorm:"not null;default:true" json:"auto_renew"`
	database.TimestampedModel
}

func (DealerSubscription) TableName() string {
	return "dealer_subscriptions"
}
