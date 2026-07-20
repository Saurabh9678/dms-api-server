CREATE TYPE subscription_status_type AS ENUM (
    'PENDING',
    'ACTIVE',
    'EXPIRED',
    'CANCELLED',
    'SUSPENDED'
);

CREATE TABLE IF NOT EXISTS dealer_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id BIGINT NOT NULL REFERENCES users(id),
    subscription_plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    subscription_pricing_id UUID NOT NULL REFERENCES subscription_pricing(id),
    status subscription_status_type NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dealer_subscriptions_dealer_id ON dealer_subscriptions(dealer_id);
CREATE INDEX IF NOT EXISTS idx_dealer_subscriptions_subscription_plan_id ON dealer_subscriptions(subscription_plan_id);
CREATE INDEX IF NOT EXISTS idx_dealer_subscriptions_subscription_pricing_id ON dealer_subscriptions(subscription_pricing_id);
CREATE INDEX IF NOT EXISTS idx_dealer_subscriptions_status ON dealer_subscriptions(status);
