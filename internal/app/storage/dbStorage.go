package storage

import (
	"context"
	"database/sql"
	"log"

	"github.com/golangTroshin/shorturl/internal/app/config"
)

type DatabaseStore struct{}

var DB *sql.DB

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

func (store *DatabaseStore) Set(ctx context.Context, value []byte) (URL, error) {
	key := generateKey(value)
	url := URL{
		UUID:        0,
		ShortURL:    key,
		OriginalURL: string(value),
	}

	query := `
		INSERT INTO urls (originUrl, shortUrl)
		VALUES ($1, $2)
		ON CONFLICT (originUrl)
		DO UPDATE SET shortUrl = EXCLUDED.shortUrl
		RETURNING id;`

	err := DB.QueryRow(query, url.OriginalURL, url.ShortURL).Scan(&url.UUID)
	if err != nil {
		log.Printf("error inserting row: %v", err)
		return url, err
	}

	log.Printf("url %v saved in db", url)
	return url, nil
}

func createTableIfNotExists() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY,
			originUrl VARCHAR(250) NOT NULL,
			shortUrl VARCHAR(250) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := DB.ExecContext(context.Background(), createTableSQL); err != nil {
		log.Printf("error creating table: %v", err)
		return err
	}

	log.Println("table created or already exists.")
	return nil
}
