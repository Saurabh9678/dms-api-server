DROP INDEX IF EXISTS idx_user_otps_request_id;

ALTER TABLE user_otps
    DROP COLUMN IF EXISTS request_id;
