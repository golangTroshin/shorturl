package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

const ContentTypeJSON = "application/json"

func APIPostHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var url storage.RequestURL

		if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
			http.Error(w, "Wrong request body", http.StatusBadRequest)
			return
		}

		urlObj, err := store.Set(r.Context(), url.URL)
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusCreated)

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

		var responseBodies []storage.ResponseBodyBanch
		for _, url := range urlObjs {
			responseBody := storage.ResponseBodyBanch{
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
