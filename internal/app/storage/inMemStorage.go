package storage

import (
	"context"
	"errors"
	"sync"
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

func (store *MemoryStore) Set(ctx context.Context, value []byte) (URL, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	key := generateShortUrl(value)
	url := URL{
		UUID:        len(store.urlList) + 1,
		ShortURL:    key,
		OriginalURL: string(value),
	}
	store.urlList[key] = url

	return url, nil
}
