package storage

import (
	"context"
	"fmt"
	"strings"

	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
	"infiour.local/dms-api-server/pkg/config"
)

// NewProvider selects a storage backend from config.
// Defaults to local when STORAGE_PROVIDER is empty or "local".
func NewProvider(cfg config.StorageConfig) (storageprovider.Provider, error) {
	switch storageprovider.NormalizeProviderName(cfg.Provider) {
	case "", "local":
		return NewLocalProvider(cfg.BasePath), nil
	case "gcs":
		if strings.TrimSpace(cfg.GCSBucket) == "" {
			return nil, fmt.Errorf("GCS_BUCKET_NAME is required when STORAGE_PROVIDER=gcs")
		}
		return NewGCSProvider(context.Background(), cfg.GCSBucket)
	default:
		return nil, fmt.Errorf("unknown STORAGE_PROVIDER %q (supported: local, gcs)", cfg.Provider)
	}
}
