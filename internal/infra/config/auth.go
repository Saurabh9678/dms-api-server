package config

import (
	"os"
	"strconv"
	"time"
)

type AuthConfig struct {
	AccessTokenSecret string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	OTPTTL            time.Duration
	OTPMaxAttempts    int
}

func LoadAuthConfig() AuthConfig {
	return AuthConfig{
		AccessTokenSecret: getEnv("AUTH_ACCESS_TOKEN_SECRET", "development-secret-change-me"),
		AccessTokenTTL:    getDurationFromSeconds("AUTH_ACCESS_TOKEN_TTL_SECONDS", 900),
		RefreshTokenTTL:   getDurationFromSeconds("AUTH_REFRESH_TOKEN_TTL_SECONDS", 604800),
		OTPTTL:            getDurationFromSeconds("AUTH_OTP_TTL_SECONDS", 300),
		OTPMaxAttempts:    getInt("AUTH_OTP_MAX_ATTEMPTS", 5),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getDurationFromSeconds(key string, fallbackSeconds int) time.Duration {
	return time.Duration(getInt(key, fallbackSeconds)) * time.Second
}
