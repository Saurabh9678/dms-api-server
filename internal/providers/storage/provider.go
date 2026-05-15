package storage

import "context"

type Provider interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
}
