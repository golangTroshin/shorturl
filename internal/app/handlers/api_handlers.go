package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

const ContentTypeJSON = "application/json"

type RequestURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	ShortURL string `json:"result"`
}

func APIPostHandler(storage storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var url RequestURL

		if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
			http.Error(w, "Wrong request body", http.StatusBadRequest)
			return
		}

		urlObj, err := storage.Set(r.Context(), []byte(url.URL))
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusCreated)

		var result ResponseShortURL
		result.ShortURL = config.Options.FlagBaseURL + "/" + urlObj.ShortURL

		if err := json.NewEncoder(w).Encode(&result); err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	return http.HandlerFunc(fn)
}
