package storage

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/golangTroshin/shorturl/internal/app/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, value []byte) (URL, error)
}

type URL struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func GetStorageByConfig() (Storage, error) {
	var store Storage
	var err error

	if config.Options.DatabaseDsn != "" {
		store, err = NewDatabaseStore()
		if err != nil {
			return store, err
		}

		return store, nil
	}

	if config.Options.StoragePath != "" {
		store, err = NewFileStore()
		if err != nil {
			return store, err
		}

		return store, nil
	}

	return NewMemoryStore(), nil
}

func generateKey(body []byte) string {
	hash := sha256.Sum256(body)
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}
