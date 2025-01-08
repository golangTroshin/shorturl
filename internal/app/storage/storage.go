package storage

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/golangTroshin/shorturl/internal/app/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Storage defines the interface for a URL storage system. It supports CRUD
// operations for URLs and batch operations for managing multiple URLs.
type Storage interface {
	Get(ctx context.Context, key string) (string, error)                   // Get retrieves the original URL corresponding to the given short URL.
	GetByUserID(ctx context.Context, userID string) ([]URL, error)         // GetByUserID retrieves all URLs associated with the specified user ID.
	Set(ctx context.Context, value string) (URL, error)                    // Set creates and stores a new short URL for the given original URL.
	SetBatch(ctx context.Context, batch []RequestBodyBanch) ([]URL, error) // SetBatch stores multiple URLs in a single operation.
	BatchDeleteURLs(userID string, batch []string) error                   // BatchDeleteURLs marks multiple URLs as deleted for a specific user.
	GetStats(ctx context.Context) (Stats, error)                           // GetStats retrieves service statistic
}

// URL represents a mapping between a short URL and its original URL.
// It includes metadata such as user ownership and deletion status.
type URL struct {
	UUID        string `json:"uuid"`         // Unique identifier for the URL
	ShortURL    string `json:"short_url"`    // Shortened URL key
	OriginalURL string `json:"original_url"` // Original URL
	UserID      string // User who owns the URL
	DeletedFlag bool   `db:"is_deleted"` // Indicates if the URL has been deleted
}

// Stats holds statistical information about saved URLs and users.
type Stats struct {
	Urls  int `json:"urls"`  // number of all saved urls
	Users int `json:"users"` // number of all users
}

// RequestURL represents the structure for incoming API requests to shorten a URL.
type RequestURL struct {
	URL string `json:"url"` // The original URL to be shortened
}

// ResponseShortURL represents the structure of the API response for a shortened URL.
type ResponseShortURL struct {
	ShortURL string `json:"result"` // The generated short URL
}

// RequestBodyBanch represents the structure of a batch request for shortening multiple URLs.
type RequestBodyBanch struct {
	CorrelationID string `json:"correlation_id"` // Identifier for the batch request
	OriginalURL   string `json:"original_url"`   // The original URL to be shortened
}

// GetStorageByConfig initializes and returns the appropriate storage system
// based on the application configuration (e.g., database, file, or memory storage).
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
