package config

import "time"

type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

type AuthConfig struct {
	AccessTokenSecret  string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	OTPTTL             time.Duration
	OTPMaxAttempts     int
	OTPCooldownSeconds int
	OTPMaxDailySends   int
}

// loaderFn is the underlying loader used by Load; swapped in whitebox tests to cover MustLoad's panic branch.
var loaderFn = defaultLoad

func Load() (*Config, error) {
	return loaderFn()
}

func defaultLoad() (*Config, error) {
	cfg := &Config{
		Env: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/dms?sslmode=disable"),
		},
		Auth: AuthConfig{
			AccessTokenSecret:  getEnv("AUTH_ACCESS_TOKEN_SECRET", "development-secret-change-me"),
			AccessTokenTTL:     getDurationFromSeconds("AUTH_ACCESS_TOKEN_TTL_SECONDS", 900),
			RefreshTokenTTL:    getDurationFromSeconds("AUTH_REFRESH_TOKEN_TTL_SECONDS", 604800),
			OTPTTL:             getDurationFromSeconds("AUTH_OTP_TTL_SECONDS", 300),
			OTPMaxAttempts:     getInt("AUTH_OTP_MAX_ATTEMPTS", 5),
			OTPCooldownSeconds: getInt("AUTH_OTP_COOLDOWN_SECONDS", 60),
			OTPMaxDailySends:   getInt("AUTH_OTP_MAX_DAILY_SENDS", 10),
		},
	}
	return cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
