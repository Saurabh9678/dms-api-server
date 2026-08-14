package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
)

func TestNormalizeProviderName(t *testing.T) {
	assert.Equal(t, "gcs", storageprovider.NormalizeProviderName(" GCS "))
	assert.Equal(t, "local", storageprovider.NormalizeProviderName("Local"))
	assert.Equal(t, "", storageprovider.NormalizeProviderName("  "))
}
