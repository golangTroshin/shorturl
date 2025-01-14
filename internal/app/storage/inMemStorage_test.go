package storage

import (
	"context"
	"testing"

	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStore_SetAndGet(t *testing.T) {
	store := NewMemoryStore()
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

func TestMemoryStore_GetNonExistent(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Test Get for a non-existent key
	_, err := store.Get(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestMemoryStore_SetBatch(t *testing.T) {
	store := NewMemoryStore()
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

func TestMemoryStore_BatchDeleteURLs(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")

	// Set URLs
	url1, _ := store.Set(ctx, "https://example1.com")
	url2, _ := store.Set(ctx, "https://example2.com")

	// Batch delete URLs
	err := store.BatchDeleteURLs("test-user", []string{url1.ShortURL, url2.ShortURL})
	assert.NoError(t, err)

	// Verify deletion
	assert.True(t, store.urlList[url1.ShortURL].DeletedFlag)
	assert.True(t, store.urlList[url2.ShortURL].DeletedFlag)
}

func TestMemoryStore_SetAndGetByUserID(t *testing.T) {
	// Initialize the memory store
	store := NewMemoryStore()

	// Create a context with a user ID
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")

	// Add URLs to the store
	store.Set(ctx, "https://example1.com")
	store.Set(ctx, "https://example2.com")

	// Retrieve URLs by user ID
	urls, err := store.GetByUserID(ctx, "test-user")

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, urls, 2)

	// Convert results to a map for comparison
	urlMap := make(map[string]bool)
	for _, url := range urls {
		urlMap[url.OriginalURL] = true
	}

	// Check the presence of expected URLs
	expectedURLs := []string{"https://example1.com", "https://example2.com"}
	for _, expectedURL := range expectedURLs {
		assert.True(t, urlMap[expectedURL], "Expected URL not found: %s", expectedURL)
	}
}
