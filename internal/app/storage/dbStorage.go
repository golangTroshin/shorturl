package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/middleware"
)

type DatabaseStore struct{}

var DB *sql.DB

type InsertConflictError struct {
	Time time.Time
	Err  error
}

func (te *InsertConflictError) Error() string {
	return fmt.Sprintf("%v %v", te.Time.Format("2006/01/02 15:04:05"), te.Err)
}

func NewInsertConflictError() error {
	return &InsertConflictError{
		Time: time.Now(),
		Err:  fmt.Errorf("conflict: originUrl already exists"),
	}
}

type DeletedURLError struct {
	Time time.Time
	Err  error
}

func (te *DeletedURLError) Error() string {
	return fmt.Sprintf("%v %v", te.Time.Format("2006/01/02 15:04:05"), te.Err)
}

func NewDeletedURLError() error {
	return &DeletedURLError{
		Time: time.Now(),
		Err:  fmt.Errorf("url was deleted"),
	}
}

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

func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Failed to close the database connection: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}
}

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

func (store *DatabaseStore) Get(ctx context.Context, key string) (string, error) {
	query := `SELECT origin_url FROM urls WHERE short_url = $1;`

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

func (store *DatabaseStore) Set(ctx context.Context, value string) (URL, error) {
	ctxValue := ctx.Value(middleware.UserIDKey)
	if ctxValue == nil {
		return URL{}, fmt.Errorf("ctxValue is nil: %v", ctxValue)
	}

	userID := ctxValue.(string)
	log.Printf("userID: %v", userID)
	url := getURLObject(value, userID)

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("error %v", err)

		return url, err
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

	if err = tx.Commit(); err != nil {
		log.Printf("error %v", err)

		return url, err
	}

	log.Printf("url %v saved in db", url)
	return url, nil
}

func (store *DatabaseStore) SetBatch(ctx context.Context, batch []RequestBodyBanch) ([]URL, error) {
	var URLs []URL

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
		" created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP)"

	if _, err := DB.ExecContext(context.Background(), createTableSQL); err != nil {
		log.Printf("error creating table: %v", err)
		return err
	}

	log.Println("table created or already exists.")
	return nil
}
