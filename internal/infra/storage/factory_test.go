package storage

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"infiour.local/dms-api-server/pkg/config"
)

func TestNewProvider_GCSSuccess(t *testing.T) {
	orig := storageNewClient
	t.Cleanup(func() { storageNewClient = orig })
	storageNewClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		return storage.NewClient(ctx, option.WithoutAuthentication())
	}

	p, err := NewProvider(config.StorageConfig{Provider: "GCS", GCSBucket: "dms-dev-assets"})
	require.NoError(t, err)
	require.NotNil(t, p)
	if closer, ok := p.(interface{ Close() error }); ok {
		require.NoError(t, closer.Close())
	}
}
