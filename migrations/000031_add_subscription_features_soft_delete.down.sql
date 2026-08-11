DROP INDEX IF EXISTS idx_subscription_features_key_active;

ALTER TABLE subscription_features
    DROP COLUMN IF EXISTS deleted_at;

CREATE UNIQUE INDEX IF NOT EXISTS subscription_features_key_key
    ON subscription_features(key);
