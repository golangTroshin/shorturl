package stores

import (
	"crypto/sha256"
	"encoding/base64"
	"sync"
)

type URLStore struct {
	mu     sync.RWMutex
	urlMap map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{
		urlMap: make(map[string]string),
	}
}

func (store *URLStore) Set(value []byte) string {
	store.mu.Lock()
	defer store.mu.Unlock()

	key := generateKey(value)
	store.urlMap[key] = string(value)

	return key
}

func (store *URLStore) Get(key string) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	val, ok := store.urlMap[key]
	return val, ok
}

func generateKey(body []byte) string {
	hash := sha256.Sum256(body)

	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}
