package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/service"
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
func ShortenURL(svc service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}

		URL, err := svc.ShortenURL(r.Context(), string(body))
		if err != nil {
			http.Error(w, "Failed to shorten URL", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", ContentTypePlainText)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(config.Options.FlagBaseURL + "/" + URL.ShortURL))
	}
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
func GetOriginalURL(svc service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		originalURL, err := svc.GetOriginalURL(r.Context(), id)
		if err != nil {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
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
func GetUserURLs(svc service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urls, err := svc.GetUserURLs(r.Context())
		if err != nil || len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		if err := json.NewEncoder(w).Encode(urls); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
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
func Ping(svc service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := svc.PingDatabase(r.Context()); err != nil {
			http.Error(w, "Database is unreachable", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Database is healthy"))
	}
}
