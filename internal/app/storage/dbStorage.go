package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/lib/pq"
)

// DatabaseStore represents the structure for database operations.
// It encapsulates methods to interact with the PostgreSQL database for URL shortening service.
type DatabaseStore struct{}

// DB represents the global database connection instance used by DatabaseStore.
var DB *sql.DB

// InsertConflictError represents an error when a conflict occurs during an INSERT operation,
// typically due to a unique constraint violation.
type InsertConflictError struct {
	Time time.Time
	Err  error
}

// Error returns a formatted error message with the timestamp and details of the conflict.
func (te *InsertConflictError) Error() string {
	return fmt.Sprintf("%v %v", te.Time.Format("2006/01/02 15:04:05"), te.Err)
}

// NewInsertConflictError creates a new instance of InsertConflictError
// indicating a conflict due to an already existing origin URL.
func NewInsertConflictError() error {
	return &InsertConflictError{
		Time: time.Now(),
		Err:  fmt.Errorf("conflict: originUrl already exists"),
	}
}

// DeletedURLError represents an error when a requested URL has been marked as deleted.
type DeletedURLError struct {
	Time time.Time
	Err  error
}

// Error returns a formatted error message with the timestamp and details of the deletion.
func (te *DeletedURLError) Error() string {
	return fmt.Sprintf("%v %v", te.Time.Format("2006/01/02 15:04:05"), te.Err)
}

// NewDeletedURLError creates a new instance of DeletedURLError
// indicating the requested URL was deleted.
func NewDeletedURLError() error {
	return &DeletedURLError{
		Time: time.Now(),
		Err:  fmt.Errorf("url was deleted"),
	}
}

// initDB initializes the database connection using the configuration.
// Returns an error if the connection cannot be established.
func initDB() error {
	var err error
	DB, err = sql.Open("pgx", config.Options.DatabaseDsn)
	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connection established")
	return nil
}

// CloseDB closes the active database connection.
// Logs any error encountered during the closure.
func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Failed to close the database connection: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}
}

// NewDatabaseStore creates a new DatabaseStore instance and initializes the database connection.
// Ensures that the necessary database table exists.
func NewDatabaseStore() (*DatabaseStore, error) {
	store := &DatabaseStore{}

	if err := initDB(); err != nil {
		return store, err
	}

	if err := createTableIfNotExists(); err != nil {
		return store, err
	}

	return store, nil
}

// Get retrieves the original URL for a given short URL from the database.
// If the short URL is marked as deleted, it returns a DeletedURLError.
func (store *DatabaseStore) Get(ctx context.Context, key string) (string, error) {
	query := `SELECT origin_url, is_deleted FROM urls WHERE short_url = $1;`

	var originalURL string
	var isDeleted bool
	err := DB.QueryRowContext(ctx, query, key).Scan(&originalURL, &isDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("there is no rows: %v", err)
			return "", nil
		}

		log.Printf("error getting row: %v", err)
		return "", err
	}

	log.Printf("url is found: %v %v", originalURL, isDeleted)
	if isDeleted {
		log.Printf("url was deleted: %v", originalURL)
		return "", NewDeletedURLError()
	}

	return originalURL, nil
}

// GetByUserID retrieves all URLs associated with a given user ID.
// Returns a slice of URLs or an error if the query fails.
func (store *DatabaseStore) GetByUserID(ctx context.Context, userID string) ([]URL, error) {
	var URLs []URL

	query := `SELECT origin_url, short_url FROM urls WHERE user_id = $1;`

	rows, err := DB.QueryContext(ctx, query, userID)

	if rows.Err() != nil || err != nil {
		log.Printf("error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url URL
		err := rows.Scan(&url.OriginalURL, &url.ShortURL)
		if err != nil {
			log.Printf("error scanning row: %v", err)
			return nil, err
		}
		URLs = append(URLs, url)
	}

	return URLs, nil
}

// Set inserts a new URL into the database with the provided original URL and user ID.
// If the original URL already exists, it retrieves the existing short URL and returns
// an InsertConflictError.
func (store *DatabaseStore) Set(ctx context.Context, value string) (URL, error) {
	ctxValue := ctx.Value(middleware.UserIDKey)
	if ctxValue == nil {
		return URL{}, fmt.Errorf("ctxValue is nil: %v", ctxValue)
	}

	userID := ctxValue.(string)
	log.Printf("userID: %v", userID)
	url := getURLObject(value, userID)

	result, err := DB.ExecContext(ctx, `
        INSERT INTO urls (origin_url, short_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (origin_url) DO NOTHING`, url.OriginalURL, url.ShortURL, userID)

	if err != nil {
		log.Printf("error %v", err)

		return url, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("error %v", err)

		return url, err
	}

	if rowsAffected == 0 {
		var existingShortURL string

		queryErr := DB.QueryRowContext(ctx, `
            SELECT short_url FROM urls WHERE origin_url = $1`, url.OriginalURL).Scan(&existingShortURL)

		if queryErr != nil {
			log.Printf("error %v", err)

			return url, queryErr
		}

		url.ShortURL = existingShortURL
		log.Printf("conflict: originUrl %v already exists", url.OriginalURL)

		return url, NewInsertConflictError()
	}

	log.Printf("url %v saved in db", url)
	return url, nil
}

// SetBatch inserts multiple URLs into the database within a single transaction.
// If any operation fails, the transaction is rolled back.
func (store *DatabaseStore) SetBatch(ctx context.Context, batch []RequestBodyBanch) ([]URL, error) {
	URLs := make([]URL, 0, len(batch))

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("error start transaction: %v", err)
		return URLs, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (origin_url, short_url, user_id) "+
			"VALUES ($1, $2, $3);")

	if err != nil {
		log.Printf("error preparing context: %v", err)
		return URLs, err
	}
	defer stmt.Close()

	userID := ctx.Value(middleware.UserIDKey).(string)
	for _, url := range batch {
		urlObj := getURLObjectWithID(url.CorrelationID, url.OriginalURL, userID)
		_, err = stmt.ExecContext(ctx, urlObj.OriginalURL, urlObj.ShortURL, userID)
		if err != nil {
			log.Printf("error inserting row: %v", err)
			return URLs, err
		}

		URLs = append(URLs, urlObj)
	}

	if err = tx.Commit(); err != nil {
		log.Printf("error commit transaction: %v", err)
		return URLs, err
	}

	return URLs, nil
}

// BatchDeleteURLs marks multiple URLs as deleted for a specific user ID.
// Returns an error if the operation fails.
func (store *DatabaseStore) BatchDeleteURLs(userID string, urlIDs []string) error {
	log.Printf("Start delete: %v %v", urlIDs, userID)
	query := `UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1) AND user_id = $2`

	result, err := DB.Exec(query, pq.Array(urlIDs), userID)

	if err != nil {
		log.Printf("Error deleting URLs: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error fetching rows affected: %v", err)
		return err
	}

	log.Printf("Rows affected: %d", rowsAffected)

	return nil
}

func createTableIfNotExists() error {
	createTableSQL := "CREATE TABLE IF NOT EXISTS urls (" +
		" id SERIAL PRIMARY KEY," +
		" origin_url VARCHAR(250) NOT NULL UNIQUE," +
		" short_url VARCHAR(250) NOT NULL," +
		" user_id VARCHAR(250) NOT NULL," +
		" is_deleted BOOL NOT NULL DEFAULT FALSE," +
		" created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP)"

	if _, err := DB.ExecContext(context.Background(), createTableSQL); err != nil {
		log.Printf("error creating table: %v", err)
		return err
	}

	log.Println("table created or already exists.")
	return nil
}

// GetStats retrieves service statistic
func (store *DatabaseStore) GetStats(ctx context.Context) (Stats, error) {
	query := `
	SELECT COUNT(*) AS total_urls, 
		   COUNT(DISTINCT user_id) AS unique_users 
	FROM urls;
`

	var totalURLs int
	var uniqueUsers int
	err := DB.QueryRowContext(ctx, query).Scan(&totalURLs, &uniqueUsers)
	if err != nil {
		log.Printf("error getting stats: %v", err)
		return Stats{}, err
	}

	log.Printf("Total URLs: %d, Unique Users: %d", totalURLs, uniqueUsers)

	return Stats{
		Urls:  totalURLs,
		Users: uniqueUsers,
	}, nil
}
