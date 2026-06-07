-- Drop user_id (PostgreSQL automatically drops idx_user_otps_user_id with the column)
ALTER TABLE user_otps DROP COLUMN IF EXISTS user_id;

-- Add phone identity snapshot columns (DEFAULT '' satisfies NOT NULL for any pre-existing rows)
ALTER TABLE user_otps
    ADD COLUMN country_code VARCHAR NOT NULL DEFAULT '',
    ADD COLUMN phone_number VARCHAR NOT NULL DEFAULT '';

ALTER TABLE user_otps
    ALTER COLUMN country_code DROP DEFAULT,
    ALTER COLUMN phone_number DROP DEFAULT;

-- Composite index supports FindLatestByPhone (ORDER BY created_at DESC) and CountRecentByPhone (created_at >= ?)
CREATE INDEX IF NOT EXISTS idx_user_otps_phone_created
    ON user_otps(country_code, phone_number, created_at);
