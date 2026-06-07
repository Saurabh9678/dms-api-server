package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"infiour.local/dms-api-server/pkg/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset all relevant env vars so defaults apply
	vars := []string{
		"APP_ENV", "APP_PORT", "DB_URL",
		"AUTH_ACCESS_TOKEN_SECRET", "AUTH_ACCESS_TOKEN_TTL_SECONDS",
		"AUTH_REFRESH_TOKEN_TTL_SECONDS", "AUTH_OTP_TTL_SECONDS",
		"AUTH_OTP_MAX_ATTEMPTS", "AUTH_OTP_COOLDOWN_SECONDS", "AUTH_OTP_MAX_DAILY_SENDS",
	}
	for _, v := range vars {
		t.Setenv(v, "")
	}

	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, 900*time.Second, cfg.Auth.AccessTokenTTL)
	assert.Equal(t, 604800*time.Second, cfg.Auth.RefreshTokenTTL)
	assert.Equal(t, 300*time.Second, cfg.Auth.OTPTTL)
	assert.Equal(t, 5, cfg.Auth.OTPMaxAttempts)
	assert.Equal(t, 60, cfg.Auth.OTPCooldownSeconds)
	assert.Equal(t, 10, cfg.Auth.OTPMaxDailySends)
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("AUTH_OTP_COOLDOWN_SECONDS", "120")
	t.Setenv("AUTH_OTP_MAX_DAILY_SENDS", "5")
	t.Setenv("AUTH_OTP_MAX_ATTEMPTS", "3")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, 120, cfg.Auth.OTPCooldownSeconds)
	assert.Equal(t, 5, cfg.Auth.OTPMaxDailySends)
	assert.Equal(t, 3, cfg.Auth.OTPMaxAttempts)
}

func TestLoad_InvalidIntFallsBackToDefault(t *testing.T) {
	t.Setenv("AUTH_OTP_COOLDOWN_SECONDS", "not-a-number")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, 60, cfg.Auth.OTPCooldownSeconds)
}

func TestMustLoad_ReturnsConfig(t *testing.T) {
	cfg := config.MustLoad()
	assert.NotNil(t, cfg)
}

func TestLoad_EnvVarPresent(t *testing.T) {
	t.Setenv("AUTH_ACCESS_TOKEN_SECRET", "test-secret")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "test-secret", cfg.Auth.AccessTokenSecret)
}
