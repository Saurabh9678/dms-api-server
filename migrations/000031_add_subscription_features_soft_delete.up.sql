ALTER TABLE subscription_features
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE subscription_features
    DROP CONSTRAINT IF EXISTS subscription_features_key_key;

DROP INDEX IF EXISTS subscription_features_key_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_features_key_active
    ON subscription_features(key)
    WHERE deleted_at IS NULL;
