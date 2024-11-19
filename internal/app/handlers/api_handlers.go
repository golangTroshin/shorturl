package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/middleware"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

// ContentTypeJSON defines the Content-Type for JSON responses.
const ContentTypeJSON = "application/json"

// APIPostHandler returns an HTTP handler for creating a shortened URL.
//
// This handler processes a POST request with a JSON payload containing the original URL.
// It generates a shortened URL and returns it in the response.
//
// Parameters:
//   - store: The storage interface for managing URL persistence.
//
// Returns:
//   - An `http.HandlerFunc` that handles the URL shortening request.
func APIPostHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var url storage.RequestURL

		if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
			http.Error(w, "Wrong request body", http.StatusBadRequest)
			return
		}

		status := http.StatusCreated

		urlObj, err := store.Set(r.Context(), url.URL)
		if err != nil {
			var target *storage.InsertConflictError

			if errors.As(err, &target) {
				status = http.StatusConflict
			}
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(status)

		var result storage.ResponseShortURL
		result.ShortURL = config.Options.FlagBaseURL + "/" + urlObj.ShortURL

		if err := json.NewEncoder(w).Encode(&result); err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	return http.HandlerFunc(fn)
}

// APIPostBatchHandler returns an HTTP handler for creating multiple shortened URLs in a batch.
//
// This handler processes a POST request with a JSON array payload containing multiple original URLs.
// It generates shortened URLs for each input and returns them in the response.
//
// Parameters:
//   - store: The storage interface for managing URL persistence.
//
// Returns:
//   - An `http.HandlerFunc` that handles the batch URL shortening request.
func APIPostBatchHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestBodies []storage.RequestBodyBanch
		err := json.NewDecoder(r.Body).Decode(&requestBodies)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		urlObjs, err := store.SetBatch(r.Context(), requestBodies)
		log.Printf("urlObjs %v", urlObjs)
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusCreated)

		type responseBodyBatch struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}
		var responseBodies []responseBodyBatch
		for _, url := range urlObjs {
			responseBody := responseBodyBatch{
				CorrelationID: url.UUID,
				ShortURL:      config.Options.FlagBaseURL + "/" + url.ShortURL,
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

// APIDeleteUrlsHandler returns an HTTP handler for deleting a batch of URLs for a user.
//
// This handler processes a DELETE request with a JSON array payload containing URL IDs to delete.
// The URLs are removed from the storage associated with the authenticated user.
//
// Parameters:
//   - store: The storage interface for managing URL persistence.
//
// Returns:
//   - An `http.HandlerFunc` that handles the batch URL deletion request.
func APIDeleteUrlsHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var urlIDs []string
		err := json.NewDecoder(r.Body).Decode(&urlIDs)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value(middleware.UserIDKey).(string)
		log.Printf("urls: %v, user: %v", urlIDs, userID)

		deleteChan <- deleteRequest{URLIDs: urlIDs, UserID: userID}

		w.WriteHeader(http.StatusAccepted)
	}

	return http.HandlerFunc(fn)
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
