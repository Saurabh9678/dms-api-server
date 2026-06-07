CREATE UNIQUE INDEX IF NOT EXISTS idx_users_country_code_phone_number
    ON users(country_code, phone_number);
