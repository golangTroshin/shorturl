package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/golangTroshin/shorturl/internal/app/middleware"
)

type MemoryStore struct {
	mu      sync.RWMutex
	urlList map[string]URL
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urlList: make(map[string]URL),
	}
}

func (store *MemoryStore) Get(ctx context.Context, key string) (string, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	val, ok := store.urlList[key]
	if !ok {
		return "", errors.New("no info about requested route")
	}

	return val.OriginalURL, nil
}

func (store *MemoryStore) GetByUserID(ctx context.Context, userID string) ([]URL, error) {
	var URLs []URL

	return URLs, nil
}

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
