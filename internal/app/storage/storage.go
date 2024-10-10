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
	GetByUserID(ctx context.Context, userID string) ([]URL, error)
	Set(ctx context.Context, value string) (URL, error)
	SetBatch(ctx context.Context, batch []RequestBodyBanch) ([]URL, error)
	BatchDeleteURLs(userID string, batch []string) error
}

type URL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string
}

type RequestURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	ShortURL string `json:"result"`
}

type RequestBodyBanch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
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

func getURLObject(url string, userID string) URL {
	key := generateShortURL(url)
	return URL{
		UUID:        "uuid_" + key,
		ShortURL:    key,
		OriginalURL: url,
		UserID:      userID,
	}
}

func getURLObjectWithID(uuid string, url string, userID string) URL {
	key := generateShortURL(url)

	return URL{
		UUID:        uuid,
		ShortURL:    key,
		OriginalURL: url,
		UserID:      userID,
	}
}

func generateShortURL(body string) string {
	hash := sha256.Sum256([]byte(body))
	return base64.URLEncoding.EncodeToString(hash[:])[:8]
}
