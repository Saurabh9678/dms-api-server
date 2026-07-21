DROP INDEX IF EXISTS idx_users_country_code_phone_number;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_country_code_phone_number
    ON users(country_code, phone_number)
    WHERE deleted_at IS NULL;
