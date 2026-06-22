package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
)

var _ storageprovider.Provider = (*LocalProvider)(nil)

type LocalProvider struct {
	basePath string
}

func NewLocalProvider(basePath string) *LocalProvider {
	return &LocalProvider{basePath: basePath}
}

func (p *LocalProvider) Upload(_ context.Context, key string, data []byte, _ string) (string, error) {
	fullPath := filepath.Join(p.basePath, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("create storage directory: %w", err)
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return key, nil
}
