package subscription_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/subscription"
)

func TestSubscriptionPlanTableName(t *testing.T) {
	assert.Equal(t, "subscription_plans", subscription.SubscriptionPlan{}.TableName())
}

func TestSubscriptionFeatureTableName(t *testing.T) {
	assert.Equal(t, "subscription_features", subscription.SubscriptionFeature{}.TableName())
}

func TestSubscriptionPricingTableName(t *testing.T) {
	assert.Equal(t, "subscription_pricing", subscription.SubscriptionPricing{}.TableName())
}

func TestPlanFeatureTableName(t *testing.T) {
	assert.Equal(t, "plan_features", subscription.PlanFeature{}.TableName())
}

func TestDealerSubscriptionTableName(t *testing.T) {
	assert.Equal(t, "dealer_subscriptions", subscription.DealerSubscription{}.TableName())
}

func TestBillingCycleConstants(t *testing.T) {
	assert.Equal(t, subscription.BillingCycle("MONTHLY"), subscription.BillingCycleMonthly)
	assert.Equal(t, subscription.BillingCycle("YEARLY"), subscription.BillingCycleYearly)
}

func TestFeatureValueTypeConstants(t *testing.T) {
	assert.Equal(t, subscription.FeatureValueType("BOOLEAN"), subscription.FeatureValueTypeBoolean)
	assert.Equal(t, subscription.FeatureValueType("STRING"), subscription.FeatureValueTypeString)
}

func TestSubscriptionStatusConstants(t *testing.T) {
	assert.Equal(t, subscription.SubscriptionStatus("ACTIVE"), subscription.SubscriptionStatusActive)
	assert.Equal(t, subscription.SubscriptionStatus("SUSPENDED"), subscription.SubscriptionStatusSuspended)
}
