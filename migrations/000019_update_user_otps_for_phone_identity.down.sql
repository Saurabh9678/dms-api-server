DROP INDEX IF EXISTS idx_user_otps_phone_created;

ALTER TABLE user_otps
    DROP COLUMN IF EXISTS country_code,
    DROP COLUMN IF EXISTS phone_number;

-- Restore user_id as nullable (cannot restore NOT NULL without a data backfill)
ALTER TABLE user_otps ADD COLUMN user_id BIGINT REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_user_otps_user_id ON user_otps(user_id);
