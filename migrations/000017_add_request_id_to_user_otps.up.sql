ALTER TABLE user_otps
    ADD COLUMN request_id VARCHAR(8);

UPDATE user_otps
SET request_id = RIGHT(LPAD(UPPER(TO_HEX(id)), 8, '0'), 8)
WHERE request_id IS NULL;

ALTER TABLE user_otps
    ALTER COLUMN request_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_otps_request_id ON user_otps(request_id);
