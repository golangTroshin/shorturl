package stores

import (
	"crypto/sha256"
	"encoding/base64"
	"sync"
)

type URLStore struct {
	mu     sync.Mutex
	urlMap map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{
		urlMap: make(map[string]string),
	}
}

func (store *URLStore) Set(key string, value string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.urlMap[key] = value
}

func (store *URLStore) Get(key string) (string, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	val, ok := store.urlMap[key]
	return val, ok
}

func GenerateKey(body []byte) string {
	hash := sha256.Sum256(body)

	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}
