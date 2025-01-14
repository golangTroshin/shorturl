package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

// Service defines the interface for the URL service.
type Service interface {
	ShortenURL(ctx context.Context, originalURL string) (storage.URL, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	BatchShortenURLs(ctx context.Context, urls []storage.RequestBodyBanch) ([]storage.URL, error)
	GetUserURLs(ctx context.Context) ([]storage.URL, error)
	DeleteUserURLs(ctx context.Context, shortURLs []string) error
	GetStats(ctx context.Context) (storage.Stats, error)
	PingDatabase(ctx context.Context) error
}

var _ Service = (*URLService)(nil) // Ensures URLService implements Service

// URLService is a struct that provides URL shortening and retrieval functionality.
// It implements the Service interface, ensuring compliance with all defined methods.
type URLService struct {
	store storage.Storage
}

// NewURLService initializes the service with the provided storage.
func NewURLService(store storage.Storage) *URLService {
	return &URLService{store: store}
}

// ShortenURL shortens a single URL.
func (s *URLService) ShortenURL(ctx context.Context, originalURL string) (storage.URL, error) {
	url, err := s.store.Set(ctx, originalURL)
	if err != nil {
		return storage.URL{}, err
	}
	return url, nil
}

// GetOriginalURL retrieves the original URL by its short URL.
func (s *URLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	return s.store.Get(ctx, shortURL)
}

// BatchShortenURLs shortens multiple URLs in a batch.
func (s *URLService) BatchShortenURLs(ctx context.Context, urls []storage.RequestBodyBanch) ([]storage.URL, error) {
	return s.store.SetBatch(ctx, urls)
}

// GetUserURLs retrieves all URLs for a given user ID.
func (s *URLService) GetUserURLs(ctx context.Context) ([]storage.URL, error) {
	var urls []storage.URL

	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		log.Printf("Wrong userID: %v", userID)
		return urls, errors.New("user ID is empty")
	}

	urls, err := s.store.GetByUserID(ctx, userID)

	if err != nil {
		return urls, err
	}

	if len(urls) == 0 {
		log.Printf("GetURLsByUser: No URLs found for user %s", userID)
		return urls, nil // Return empty response
	}

	log.Printf("GetURLsByUser: Found %d URLs for user %s", len(urls), userID)

	return urls, nil
}

// deleteRequest represents a request to delete URLs for a user.
//
// Fields:
//   - URLIDs: A slice of URL IDs to be deleted.
//   - UserID: The ID of the user requesting the deletion.
type deleteRequest struct {
	URLIDs []string
	UserID string
}

// deleteChan is a buffered channel used for queuing URL deletion requests.
var deleteChan = make(chan deleteRequest, 100)

// DeleteUserURLs deletes URLs for a user in a batch.
func (s *URLService) DeleteUserURLs(ctx context.Context, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return errors.New("no URLs provided for deletion")
	}

	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		log.Printf("Wrong userID: %v", userID)
		return errors.New("wrong userID")
	}
	log.Printf("urls: %v, user: %v", shortURLs, userID)

	deleteChan <- deleteRequest{URLIDs: shortURLs, UserID: userID}

	return nil
}

// StartDeleteWorker starts a worker that processes URL deletion requests from the `deleteChan`.
//
// This worker listens for `deleteRequest` objects on the channel and performs batch URL deletions
// using the provided storage interface.
//
// Parameters:
//   - store: The storage interface for managing URL persistence.
//
// Usage:
//
//	This function is typically started as a goroutine.
func StartDeleteWorker(store storage.Storage) {
	for req := range deleteChan {
		err := store.BatchDeleteURLs(req.UserID, req.URLIDs)

		if err != nil {
			log.Printf("Error deleting URLs for user %v: %v", req.UserID, err)
		}
	}
}

// GetStats retrieves URL and user statistics.
func (s *URLService) GetStats(ctx context.Context) (storage.Stats, error) {
	return s.store.GetStats(ctx)
}

// PingDatabase checks the health of the database connection.
func (s *URLService) PingDatabase(ctx context.Context) error {
	db, err := sql.Open("pgx", config.Options.DatabaseDsn)

	if err != nil {
		log.Printf("Database connection error: %v", err)
		return errors.New("unable to connect to database")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("Database ping error: %v", err)
		return errors.New("database is unreachable")
	}

	log.Println("Database is healthy")

	return nil
}
