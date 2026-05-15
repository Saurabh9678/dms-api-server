package config

import (
	"os"
	"strconv"
	"time"
)

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
