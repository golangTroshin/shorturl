package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/stores"
)

const ContentTypeJSON = "application/json"

type RequestURL struct {
	URL string `json:"url"`
}

type ResponseShortURL struct {
	ShortURL string `json:"result"`
}

func APIPostHandler(store *stores.URLStore) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var url RequestURL

		if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
			http.Error(w, "Wrong request body", http.StatusBadRequest)
			return
		}

		key := store.Set([]byte(url.URL))

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusCreated)

		var result ResponseShortURL
		result.ShortURL = config.Options.FlagBaseURL + "/" + key

		if err := json.NewEncoder(w).Encode(&result); err != nil {
			log.Panicf("Unable to write reponse: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	return http.HandlerFunc(fn)
}
