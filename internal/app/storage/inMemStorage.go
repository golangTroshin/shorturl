package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/golangTroshin/shorturl/internal/app/middleware"
)

// MemoryStore represents an in-memory storage for URLs.
// It uses a thread-safe map to store and manage URL data.
type MemoryStore struct {
	mu      sync.RWMutex   // Ensures thread-safe access to the urlList map.
	urlList map[string]URL // Stores mapping of short URLs to full URL objects.
}

// NewMemoryStore initializes and returns a new MemoryStore instance.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urlList: make(map[string]URL),
	}
}

// Get retrieves the original URL corresponding to a given short URL.
// Returns an error if the short URL does not exist in the store.
func (store *MemoryStore) Get(ctx context.Context, key string) (string, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	val, ok := store.urlList[key]
	if !ok {
		return "", errors.New("no info about requested route")
	}

	return val.OriginalURL, nil
}

// GetByUserID retrieves all URLs associated with a given user ID.
// Currently, the implementation returns an empty list (placeholder).
func (store *MemoryStore) GetByUserID(ctx context.Context, userID string) ([]URL, error) {
	var URLs []URL

	return URLs, nil
}

// Set adds a new URL to the store, generating a unique short URL for it.
// If the user ID is present in the context, it associates the URL with the user.
func (store *MemoryStore) Set(ctx context.Context, value string) (URL, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	ctxValue := ctx.Value(middleware.UserIDKey)
	userID := "default"
	if ctxValue != nil {
		userID = ctxValue.(string)
	}

	url := getURLObject(value, userID)
	store.urlList[url.ShortURL] = url
	return url, nil
}

// SetBatch adds multiple URLs to the store in a single operation.
// Each URL is associated with a user ID derived from the context.
func (store *MemoryStore) SetBatch(ctx context.Context, urls []RequestBodyBanch) ([]URL, error) {
	var URLs []URL
	store.mu.Lock()
	defer store.mu.Unlock()

	userID := ctx.Value(middleware.UserIDKey).(string)
	for _, url := range urls {
		url := getURLObjectWithID(url.CorrelationID, url.OriginalURL, userID)
		store.urlList[url.ShortURL] = url
		URLs = append(URLs, url)
	}

	return URLs, nil
}

// BatchDeleteURLs marks multiple URLs as deleted for a specific user ID.
func (store *MemoryStore) BatchDeleteURLs(userID string, batch []string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.urlList) == 0 {
		return errors.New("no URLs in the store")
	}

	batchMap := make(map[string]struct{})
	for _, shortURL := range batch {
		batchMap[shortURL] = struct{}{}
	}

	for key, url := range store.urlList {
		if url.UserID == userID {
			if _, found := batchMap[url.ShortURL]; found {
				url.DeletedFlag = true
				store.urlList[key] = url
			}
		}
	}

	return nil
}
