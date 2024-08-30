package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golangTroshin/shorturl/internal/app/config"
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
	query := `SELECT originUrl FROM urls WHERE shortUrl = $1;`

	var originalURL string
	err := DB.QueryRowContext(ctx, query, key).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("there is no rows: %v", err)
			return "", nil
		}

		log.Printf("error getting row: %v", err)
		return "", err
	}

	log.Printf("url is found: %v", originalURL)
	return originalURL, nil
}

func (store *DatabaseStore) Set(ctx context.Context, value string) (URL, error) {
	url := getURLObject(value)

	tx, err := DB.Begin()
	if err != nil {
		return url, err
	}

	defer tx.Rollback()

	result, err := DB.ExecContext(ctx, `
        INSERT INTO urls (originUrl, shortUrl)
        VALUES ($1, $2)
        ON CONFLICT (originUrl) DO NOTHING`, url.OriginalURL, url.ShortURL)

	if err != nil {
		return url, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return url, err
	}

	if rowsAffected == 0 {
		var existingShortURL string

		queryErr := DB.QueryRowContext(ctx, `
            SELECT shortUrl FROM urls WHERE originUrl = $1`, url.OriginalURL).Scan(&existingShortURL)

		if queryErr != nil {
			return url, queryErr
		}

		url.ShortURL = existingShortURL
		log.Printf("conflict: originUrl %v already exists", url.OriginalURL)

		return url, NewInsertConflictError()
	}

	if err = tx.Commit(); err != nil {
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
		"INSERT INTO urls (originUrl, shortUrl) "+
			"VALUES ($1, $2);")

	if err != nil {
		log.Printf("error preparing context: %v", err)
		return URLs, err
	}
	defer stmt.Close()

	for _, url := range batch {
		urlObj := getURLObjectWithID(url.CorrelationID, url.OriginalURL)
		_, err = stmt.ExecContext(ctx, urlObj.OriginalURL, urlObj.ShortURL)
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

func createTableIfNotExists() error {
	createTableSQL := "CREATE TABLE IF NOT EXISTS urls (" +
		" id SERIAL PRIMARY KEY," +
		" originUrl VARCHAR(250) NOT NULL UNIQUE," +
		" shortUrl VARCHAR(250) NOT NULL," +
		" created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP)"

	if _, err := DB.ExecContext(context.Background(), createTableSQL); err != nil {
		log.Printf("error creating table: %v", err)
		return err
	}

	log.Println("table created or already exists.")
	return nil
}
