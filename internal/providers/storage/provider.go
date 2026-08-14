package storage

import (
	"context"
	"strings"
	"time"
)

// Provider abstracts object storage used by business modules.
// Upload returns the object key to persist; SignedURL returns a time-limited access URL for clients.
type Provider interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// NormalizeProviderName lowercases and trims a STORAGE_PROVIDER value.
func NormalizeProviderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
