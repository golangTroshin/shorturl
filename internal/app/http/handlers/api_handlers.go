package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/service"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

// ContentTypeJSON defines the Content-Type for JSON responses.
const ContentTypeJSON = "application/json"

// APIShortenURL returns an HTTP handler for creating a shortened URL.
//
// This handler processes a POST request with a JSON payload containing the original URL.
// It generates a shortened URL and returns it in the response using the provided service.
//
// Parameters:
//   - svc: The URL service for handling business logic.
//
// Returns:
//   - An `http.HandlerFunc` that handles the URL shortening request.
func APIShortenURL(svc service.Service) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var url storage.RequestURL

		if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
			http.Error(w, "Wrong request body", http.StatusBadRequest)
			return
		}

		status := http.StatusCreated

		urlObj, err := svc.ShortenURL(r.Context(), url.URL)
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
// It generates shortened URLs for each input and returns them in the response using the provided service.
//
// Parameters:
//   - svc: The URL service for handling business logic.
//
// Returns:
//   - An `http.HandlerFunc` that handles the batch URL shortening request.
func APIPostBatchHandler(svc service.Service) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestBodies []storage.RequestBodyBanch
		err := json.NewDecoder(r.Body).Decode(&requestBodies)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		urlObjs, err := svc.BatchShortenURLs(r.Context(), requestBodies)
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
// The URLs are removed using the provided service.
//
// Parameters:
//   - svc: The URL service for handling business logic.
//
// Returns:
//   - An `http.HandlerFunc` that handles the batch URL deletion request.
func APIDeleteUrlsHandler(svc service.Service) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var urlIDs []string
		err := json.NewDecoder(r.Body).Decode(&urlIDs)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err = svc.DeleteUserURLs(r.Context(), urlIDs)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}

	return http.HandlerFunc(fn)
}

// APIInternalGetStatsHandler returns an HTTP handler that provides statistics
// about stored URLs and users.
//
// This handler fetches statistics from the service and returns them in JSON format.
//
// Parameters:
//   - svc: The URL service for handling business logic.
//
// Returns:
//   - An `http.HandlerFunc` that handles the statistics request.
func APIInternalGetStatsHandler(svc service.Service) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		stats, err := svc.GetStats(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", ContentTypeJSON)

		log.Printf("responseBodies %v", stats)

		if err := json.NewEncoder(w).Encode(&stats); err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	return http.HandlerFunc(fn)
}
