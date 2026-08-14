package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	infrastorage "infiour.local/dms-api-server/internal/infra/storage"
	"infiour.local/dms-api-server/pkg/config"
)

func TestLocalProvider_Upload_Success(t *testing.T) {
	base := t.TempDir()
	provider := infrastorage.NewLocalProvider(base)

	data := []byte("hello world")
	key := "user1/showroom/SHOP0001/file.jpg"

	path, err := provider.Upload(context.Background(), key, data, "image/jpeg")
	require.NoError(t, err)
	assert.Equal(t, key, path)

	written, err := os.ReadFile(filepath.Join(base, key))
	require.NoError(t, err)
	assert.Equal(t, data, written)
}

func TestLocalProvider_Upload_MkdirAllError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "basepath-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })
	require.NoError(t, tmpFile.Close())

	provider := infrastorage.NewLocalProvider(tmpFile.Name())

	_, err = provider.Upload(context.Background(), "user1/file.jpg", []byte("data"), "image/jpeg")
	assert.Error(t, err)
}

func TestLocalProvider_Upload_WriteFileError(t *testing.T) {
	base := t.TempDir()
	provider := infrastorage.NewLocalProvider(base)

	targetDir := filepath.Join(base, "user1", "file.jpg")
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	_, err := provider.Upload(context.Background(), "user1/file.jpg", []byte("data"), "image/jpeg")
	assert.Error(t, err)
}

func TestLocalProvider_SignedURL_ReturnsKey(t *testing.T) {
	provider := infrastorage.NewLocalProvider(t.TempDir())
	url, err := provider.SignedURL(context.Background(), "1/showroom/ABC12345/x.jpg", time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "1/showroom/ABC12345/x.jpg", url)
}

func TestLocalProvider_SignedURL_EmptyKey(t *testing.T) {
	provider := infrastorage.NewLocalProvider(t.TempDir())
	_, err := provider.SignedURL(context.Background(), "", time.Hour)
	assert.Error(t, err)
}

func TestNewProvider_LocalDefault(t *testing.T) {
	p, err := infrastorage.NewProvider(config.StorageConfig{Provider: "", BasePath: t.TempDir()})
	require.NoError(t, err)
	require.NotNil(t, p)

	key, err := p.Upload(context.Background(), "a/b.jpg", []byte("x"), "image/jpeg")
	require.NoError(t, err)
	assert.Equal(t, "a/b.jpg", key)
}

func TestNewProvider_LocalExplicit(t *testing.T) {
	p, err := infrastorage.NewProvider(config.StorageConfig{Provider: "local", BasePath: t.TempDir()})
	require.NoError(t, err)
	require.NotNil(t, p)
}

func TestNewProvider_GCSMissingBucket(t *testing.T) {
	_, err := infrastorage.NewProvider(config.StorageConfig{Provider: "gcs"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GCS_BUCKET_NAME")
}

func TestNewProvider_Unknown(t *testing.T) {
	_, err := infrastorage.NewProvider(config.StorageConfig{Provider: "s3"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown STORAGE_PROVIDER")
}
