package storage

import (
	"context"
	"os"
	"testing"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestFileStore_SetAndGet(t *testing.T) {
	// Setup temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_store_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	config.Options.StoragePath = tmpFile.Name()

	store, err := NewFileStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")

	// Test Set
	originalURL := "https://example.com"
	url, err := store.Set(ctx, originalURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, url.OriginalURL)

	// Test Get
	retrievedURL, err := store.Get(ctx, url.ShortURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)
}

func TestFileStore_BatchDeleteURLs(t *testing.T) {
	// Setup temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_store_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	config.Options.StoragePath = tmpFile.Name()

	store, err := NewFileStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")

	// Set URLs
	url1, _ := store.Set(ctx, "https://example1.com")
	url2, _ := store.Set(ctx, "https://example2.com")

	// Batch delete URLs
	err = store.BatchDeleteURLs("test-user", []string{url1.ShortURL, url2.ShortURL})
	assert.NoError(t, err)

	// Verify deletion
	assert.True(t, store.urlList[url1.ShortURL].DeletedFlag)
	assert.True(t, store.urlList[url2.ShortURL].DeletedFlag)
}

func TestFileStore_loadFromFile(t *testing.T) {
	// Setup temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_store_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	config.Options.StoragePath = tmpFile.Name()

	// Prepare test data
	producer, err := NewProducer(config.Options.StoragePath)
	assert.NoError(t, err)

	url := URL{
		UUID:        "uuid123",
		ShortURL:    "short123",
		OriginalURL: "https://example.com",
		UserID:      "test-user",
	}
	err = producer.WriteURL(&url)
	assert.NoError(t, err)
	producer.Close()

	// Load from file
	store, err := NewFileStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)

	retrievedURL, ok := store.urlList[url.ShortURL]
	assert.True(t, ok)
	assert.Equal(t, url.OriginalURL, retrievedURL.OriginalURL)
}

func TestFileStore_SetBatch(t *testing.T) {
	// Setup temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_store_*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	config.Options.StoragePath = tmpFile.Name()

	store, err := NewFileStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")

	batch := []RequestBodyBanch{
		{CorrelationID: "id1", OriginalURL: "https://example1.com"},
		{CorrelationID: "id2", OriginalURL: "https://example2.com"},
	}

	urls, err := store.SetBatch(ctx, batch)
	assert.NoError(t, err)
	assert.Equal(t, len(batch), len(urls))

	for i, url := range urls {
		assert.Equal(t, batch[i].OriginalURL, url.OriginalURL)
	}
}
