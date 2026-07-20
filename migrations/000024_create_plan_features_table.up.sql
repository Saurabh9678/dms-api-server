CREATE TABLE IF NOT EXISTS plan_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    subscription_feature_id UUID NOT NULL REFERENCES subscription_features(id),
    bool_value BOOLEAN,
    numeric_value NUMERIC(18, 2),
    string_value TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (subscription_plan_id, subscription_feature_id)
);

CREATE INDEX IF NOT EXISTS idx_plan_features_subscription_plan_id ON plan_features(subscription_plan_id);
CREATE INDEX IF NOT EXISTS idx_plan_features_subscription_feature_id ON plan_features(subscription_feature_id);
