CREATE TYPE platform_type AS ENUM ('web', 'ios_mobile', 'android_mobile', 'desktop');
CREATE TYPE otp_for_type AS ENUM ('mobile', 'email');

CREATE TABLE IF NOT EXISTS user_otps (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    otp_code VARCHAR(6) NOT NULL,
    platform platform_type NOT NULL,
    otp_for otp_for_type NOT NULL,
    device_id VARCHAR(255),
    attempt_count INT NOT NULL DEFAULT 0,
    resend_count INT NOT NULL DEFAULT 0,
    is_used BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_user_otps_user_id ON user_otps(user_id);
CREATE INDEX IF NOT EXISTS idx_user_otps_otp_code ON user_otps(otp_code);
