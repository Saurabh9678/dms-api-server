package subscription

import "errors"

var (
	ErrPlanNotFound               = errors.New("subscription plan not found")
	ErrFeatureNotFound            = errors.New("subscription feature not found")
	ErrPricingNotFound            = errors.New("subscription pricing not found")
	ErrPlanFeatureNotFound        = errors.New("plan feature not found")
	ErrDealerSubscriptionNotFound = errors.New("dealer subscription not found")
)
