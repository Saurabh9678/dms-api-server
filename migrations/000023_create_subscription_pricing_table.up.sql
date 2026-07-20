CREATE TYPE billing_cycle_type AS ENUM (
    'MONTHLY',
    'QUARTERLY',
    'HALF_YEARLY',
    'YEARLY'
);

CREATE TABLE IF NOT EXISTS subscription_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    billing_cycle billing_cycle_type NOT NULL,
    duration_days INT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (subscription_plan_id, billing_cycle)
);

CREATE INDEX IF NOT EXISTS idx_subscription_pricing_subscription_plan_id ON subscription_pricing(subscription_plan_id);
