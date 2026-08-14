package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
)

var _ storageprovider.Provider = (*GCSProvider)(nil)

// storageNewClient is swapped in tests to avoid requiring live GCP credentials.
var storageNewClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
	return storage.NewClient(ctx, opts...)
}

type gcsObjectStore interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) error
	SignedURL(key string, ttl time.Duration) (string, error)
	Close() error
}

type GCSProvider struct {
	store gcsObjectStore
}

// NewGCSProvider creates a GCS-backed storage provider using Application Default Credentials
// (including ADC with service-account impersonation).
func NewGCSProvider(ctx context.Context, bucket string) (*GCSProvider, error) {
	if bucket == "" {
		return nil, fmt.Errorf("gcs bucket name is required")
	}
	client, err := storageNewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create gcs client: %w", err)
	}
	return &GCSProvider{store: newGCSClientStore(client, bucket)}, nil
}

func newGCSProviderWithStore(store gcsObjectStore) *GCSProvider {
	return &GCSProvider{store: store}
}

func (p *GCSProvider) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty object key")
	}
	if err := p.store.Upload(ctx, key, data, contentType); err != nil {
		return "", err
	}
	return key, nil
}

func (p *GCSProvider) SignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty object key")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("signed url ttl must be positive")
	}
	url, err := p.store.SignedURL(key, ttl)
	if err != nil {
		return "", fmt.Errorf("sign gcs url: %w", err)
	}
	return url, nil
}

// Close releases the underlying GCS client.
func (p *GCSProvider) Close() error {
	if p.store == nil {
		return nil
	}
	return p.store.Close()
}

type gcsWriteCloser interface {
	io.WriteCloser
	SetContentType(contentType string)
}

type gcsClientStore struct {
	bucket      string
	closeFn     func() error
	newWriterFn func(ctx context.Context, key string) gcsWriteCloser
	signedURLFn func(key string, ttl time.Duration) (string, error)
}

func newGCSClientStore(client *storage.Client, bucket string) *gcsClientStore {
	return &gcsClientStore{
		bucket:  bucket,
		closeFn: client.Close,
		newWriterFn: func(ctx context.Context, key string) gcsWriteCloser {
			return &gcsObjectWriter{Writer: client.Bucket(bucket).Object(key).NewWriter(ctx)}
		},
		signedURLFn: func(key string, ttl time.Duration) (string, error) {
			return client.Bucket(bucket).SignedURL(key, &storage.SignedURLOptions{
				Scheme:  storage.SigningSchemeV4,
				Method:  "GET",
				Expires: time.Now().Add(ttl),
			})
		},
	}
}

type gcsObjectWriter struct {
	*storage.Writer
}

func (w *gcsObjectWriter) SetContentType(contentType string) {
	w.ContentType = contentType
}

func (s *gcsClientStore) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	w := s.newWriterFn(ctx, key)
	if contentType != "" {
		w.SetContentType(contentType)
	}
	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		_ = w.Close()
		return fmt.Errorf("write gcs object: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close gcs writer: %w", err)
	}
	return nil
}

func (s *gcsClientStore) SignedURL(key string, ttl time.Duration) (string, error) {
	return s.signedURLFn(key, ttl)
}

func (s *gcsClientStore) Close() error {
	return s.closeFn()
}
