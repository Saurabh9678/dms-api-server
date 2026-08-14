package storage

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

type fakeGCSStore struct {
	uploadErr error
	signURL   string
	signErr   error
	closed    bool
	lastKey   string
	lastTTL   time.Duration
	lastCT    string
	lastData  []byte
}

func (f *fakeGCSStore) Upload(_ context.Context, key string, data []byte, contentType string) error {
	f.lastKey = key
	f.lastData = append([]byte(nil), data...)
	f.lastCT = contentType
	return f.uploadErr
}

func (f *fakeGCSStore) SignedURL(key string, ttl time.Duration) (string, error) {
	f.lastKey = key
	f.lastTTL = ttl
	if f.signErr != nil {
		return "", f.signErr
	}
	return f.signURL, nil
}

func (f *fakeGCSStore) Close() error {
	f.closed = true
	return nil
}

type fakeWriteCloser struct {
	buf         bytes.Buffer
	contentType string
	writeErr    error
	closeErr    error
	closed      bool
}

func (f *fakeWriteCloser) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return f.buf.Write(p)
}

func (f *fakeWriteCloser) Close() error {
	f.closed = true
	return f.closeErr
}

func (f *fakeWriteCloser) SetContentType(contentType string) {
	f.contentType = contentType
}

func TestGCSProvider_Upload_Success(t *testing.T) {
	store := &fakeGCSStore{}
	p := newGCSProviderWithStore(store)

	key, err := p.Upload(context.Background(), "1/showroom/ABC/x.jpg", []byte("img"), "image/jpeg")
	require.NoError(t, err)
	assert.Equal(t, "1/showroom/ABC/x.jpg", key)
	assert.Equal(t, "image/jpeg", store.lastCT)
	assert.Equal(t, []byte("img"), store.lastData)
}

func TestGCSProvider_Upload_EmptyKey(t *testing.T) {
	p := newGCSProviderWithStore(&fakeGCSStore{})
	_, err := p.Upload(context.Background(), "", []byte("x"), "image/jpeg")
	assert.Error(t, err)
}

func TestGCSProvider_Upload_StoreError(t *testing.T) {
	p := newGCSProviderWithStore(&fakeGCSStore{uploadErr: errors.New("boom")})
	_, err := p.Upload(context.Background(), "k.jpg", []byte("x"), "image/jpeg")
	assert.Error(t, err)
}

func TestGCSProvider_SignedURL_Success(t *testing.T) {
	store := &fakeGCSStore{signURL: "https://signed.example/obj"}
	p := newGCSProviderWithStore(store)

	url, err := p.SignedURL(context.Background(), "obj", time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "https://signed.example/obj", url)
	assert.Equal(t, time.Hour, store.lastTTL)
}

func TestGCSProvider_SignedURL_EmptyKey(t *testing.T) {
	p := newGCSProviderWithStore(&fakeGCSStore{})
	_, err := p.SignedURL(context.Background(), "", time.Hour)
	assert.Error(t, err)
}

func TestGCSProvider_SignedURL_InvalidTTL(t *testing.T) {
	p := newGCSProviderWithStore(&fakeGCSStore{})
	_, err := p.SignedURL(context.Background(), "obj", 0)
	assert.Error(t, err)
}

func TestGCSProvider_SignedURL_StoreError(t *testing.T) {
	p := newGCSProviderWithStore(&fakeGCSStore{signErr: errors.New("sign failed")})
	_, err := p.SignedURL(context.Background(), "obj", time.Hour)
	assert.Error(t, err)
}

func TestGCSProvider_Close(t *testing.T) {
	store := &fakeGCSStore{}
	p := newGCSProviderWithStore(store)
	require.NoError(t, p.Close())
	assert.True(t, store.closed)
}

func TestGCSProvider_Close_NilStore(t *testing.T) {
	p := &GCSProvider{}
	require.NoError(t, p.Close())
}

func TestNewGCSProvider_EmptyBucket(t *testing.T) {
	_, err := NewGCSProvider(context.Background(), "")
	require.Error(t, err)
}

func TestNewGCSProvider_ClientError(t *testing.T) {
	orig := storageNewClient
	t.Cleanup(func() { storageNewClient = orig })
	storageNewClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		return nil, errors.New("dial failed")
	}

	_, err := NewGCSProvider(context.Background(), "bucket")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create gcs client")
}

func TestNewGCSProvider_Success(t *testing.T) {
	orig := storageNewClient
	t.Cleanup(func() { storageNewClient = orig })
	storageNewClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		return storage.NewClient(ctx, option.WithoutAuthentication())
	}

	p, err := NewGCSProvider(context.Background(), "bucket")
	require.NoError(t, err)
	require.NotNil(t, p)
	require.NoError(t, p.Close())
}

func TestGCSClientStore_UploadSuccessAndErrors(t *testing.T) {
	var writer *fakeWriteCloser
	store := &gcsClientStore{
		bucket:  "b",
		closeFn: func() error { return nil },
		newWriterFn: func(ctx context.Context, key string) gcsWriteCloser {
			writer = &fakeWriteCloser{}
			return writer
		},
		signedURLFn: func(key string, ttl time.Duration) (string, error) {
			return "https://example/" + key, nil
		},
	}

	require.NoError(t, store.Upload(context.Background(), "obj.jpg", []byte("data"), "image/jpeg"))
	require.NotNil(t, writer)
	assert.Equal(t, "image/jpeg", writer.contentType)
	assert.Equal(t, "data", writer.buf.String())
	assert.True(t, writer.closed)

	store.newWriterFn = func(ctx context.Context, key string) gcsWriteCloser {
		return &fakeWriteCloser{writeErr: errors.New("write fail")}
	}
	err := store.Upload(context.Background(), "obj.jpg", []byte("data"), "")
	assert.Error(t, err)

	store.newWriterFn = func(ctx context.Context, key string) gcsWriteCloser {
		return &fakeWriteCloser{closeErr: errors.New("close fail")}
	}
	err = store.Upload(context.Background(), "obj.jpg", []byte("data"), "image/png")
	assert.Error(t, err)

	url, err := store.SignedURL("obj.jpg", time.Hour)
	require.NoError(t, err)
	assert.Equal(t, "https://example/obj.jpg", url)
	require.NoError(t, store.Close())
}

func TestGCSObjectWriter_SetContentType(t *testing.T) {
	w := &gcsObjectWriter{Writer: &storage.Writer{}}
	w.SetContentType("image/png")
	assert.Equal(t, "image/png", w.ContentType)
}

func TestGCSClientStore_NewGCSClientStoreWiresFns(t *testing.T) {
	orig := storageNewClient
	t.Cleanup(func() { storageNewClient = orig })
	storageNewClient = func(ctx context.Context, opts ...option.ClientOption) (*storage.Client, error) {
		return storage.NewClient(ctx, option.WithoutAuthentication())
	}

	client, err := storageNewClient(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	store := newGCSClientStore(client, "bucket")
	require.NotNil(t, store.newWriterFn)
	require.NotNil(t, store.signedURLFn)
	require.NotNil(t, store.closeFn)

	w := store.newWriterFn(context.Background(), "k.jpg")
	require.NotNil(t, w)
	w.SetContentType("image/jpeg")
	_, _ = w.Write([]byte("x"))
	_ = w.Close()

	_, err = store.signedURLFn("k.jpg", time.Hour)
	assert.Error(t, err) // unauthenticated client cannot sign
}
