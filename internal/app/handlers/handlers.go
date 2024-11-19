package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/middleware"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ContentTypePlainText const for content type
const ContentTypePlainText = "text/plain"

// PostRequestHandler handles HTTP POST requests to shorten URLs.
// It accepts a request body containing the URL to be shortened and interacts with the storage
// to generate or retrieve a shortened version of the URL.
//
// If the URL is successfully shortened, it returns a 201 Created status with the shortened URL.
// If the URL already exists in the storage, it returns a 409 Conflict status with the existing shortened URL.
// If the request body is empty or cannot be read, it returns a 400 Bad Request status.
//
// Parameters:
//   - store: The storage interface for managing URL data.
//
// Returns:
//   - http.HandlerFunc: A handler function to process the request.
func PostRequestHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}

		status := http.StatusCreated

		url, err := store.Set(r.Context(), string(body))
		if err != nil {
			var target *storage.InsertConflictError

			if errors.As(err, &target) {
				status = http.StatusConflict
			}
		}

		w.Header().Set("Content-Type", ContentTypePlainText)
		w.WriteHeader(status)

		_, err = w.Write([]byte(config.Options.FlagBaseURL + "/" + url.ShortURL))
		if err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, "Unable to write reponse", http.StatusNotFound)
		}
	}

	return http.HandlerFunc(fn)
}

// GetRequestHandler handles HTTP GET requests to retrieve the original URL
// corresponding to a given shortened URL ID.
//
// It extracts the "id" parameter from the URL path, queries the storage for the
// original URL, and performs the following actions:
//   - If the shortened URL exists and is active, it responds with a 307 Temporary Redirect status,
//     setting the "Location" header to the original URL.
//   - If the shortened URL has been deleted, it responds with a 410 Gone status.
//   - If the shortened URL does not exist, it responds with a 404 Not Found status.
//   - If the "id" parameter is missing or invalid, it responds with a 400 Bad Request status.
//
// Parameters:
//   - store: The storage interface for managing URL data.
//
// Returns:
//   - http.HandlerFunc: A handler function to process the request.
func GetRequestHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		status := http.StatusTemporaryRedirect

		val, err := store.Get(r.Context(), id)
		if err != nil {
			var target *storage.DeletedURLError

			if errors.As(err, &target) {
				status = http.StatusGone
			} else {
				http.Error(w, "No info about requested route", http.StatusNotFound)
				return
			}
		} else {
			w.Header().Set("Content-Type", ContentTypePlainText)
			w.Header().Set("Location", val)
		}

		w.WriteHeader(status)
	}

	return http.HandlerFunc(fn)
}

// GetURLsByUserHandler handles HTTP GET requests to retrieve all shortened URLs
// associated with the currently authenticated user.
//
// It extracts the user ID from the request context and queries the storage for
// all URLs associated with that user. The response includes the original URL and
// its corresponding shortened URL.
//
// The function performs the following actions:
//   - If URLs are found for the user, it responds with a JSON-encoded list of URLs
//     and a 200 OK status.
//   - If no URLs are found, it responds with a 204 No Content status.
//   - If an error occurs during retrieval, it responds with the appropriate HTTP status.
//
// Parameters:
//   - store: The storage interface for managing URL data.
//
// Returns:
//   - http.HandlerFunc: A handler function to process the request.
func GetURLsByUserHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)
		urls, err := store.GetByUserID(r.Context(), userID)
		if err != nil || len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		type responseGetByUserID struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}
		w.Header().Set("Content-Type", ContentTypeJSON)
		var responseBodies []responseGetByUserID
		for _, url := range urls {
			responseBody := responseGetByUserID{
				ShortURL:    config.Options.FlagBaseURL + "/" + url.ShortURL,
				OriginalURL: url.OriginalURL,
			}
			responseBodies = append(responseBodies, responseBody)
		}

		log.Printf("responseBodies %v", responseBodies)

		if err := json.NewEncoder(w).Encode(&responseBodies); err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	return http.HandlerFunc(fn)
}

// DatabasePing handles HTTP GET requests to check the health of the database connection.
//
// It performs the following actions:
//   - Attempts to establish a connection to the database using the configured DSN.
//   - If the connection is successful and the database is reachable, it responds with
//     a 200 OK status.
//   - If the connection fails or the database is unreachable, it responds with a
//     500 Internal Server Error status.
//
// Returns:
//   - http.HandlerFunc: A handler function to process the request.
func DatabasePing() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("pgx", config.Options.DatabaseDsn)

		if err != nil {
			http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		}

		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, "Unable to reach database", http.StatusInternalServerError)
		}
	}

	return http.HandlerFunc(fn)
}
