package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	infrastorage "infiour.local/dms-api-server/internal/infra/storage"
)

func TestLocalProvider_Upload_Success(t *testing.T) {
	base := t.TempDir()
	provider := infrastorage.NewLocalProvider(base)

	data := []byte("hello world")
	key := "user1/showroom1/file.jpg"

	path, err := provider.Upload(context.Background(), key, data, "image/jpeg")
	require.NoError(t, err)
	assert.Equal(t, key, path)

	written, err := os.ReadFile(filepath.Join(base, key))
	require.NoError(t, err)
	assert.Equal(t, data, written)
}

func TestLocalProvider_Upload_MkdirAllError(t *testing.T) {
	// Use a file as the base path so MkdirAll fails trying to create a directory inside a file.
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

	// Create the target path as a directory so WriteFile fails.
	targetDir := filepath.Join(base, "user1", "file.jpg")
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	_, err := provider.Upload(context.Background(), "user1/file.jpg", []byte("data"), "image/jpeg")
	assert.Error(t, err)
}
