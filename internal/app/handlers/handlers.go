package handlers

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/stores"
)

const ContentTypePlainText = "text/plain"

func PostRequestHandler(store *stores.URLStore) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}

		key := store.Set(body)

		w.Header().Set("Content-Type", ContentTypePlainText)
		w.WriteHeader(http.StatusCreated)

		_, err = w.Write([]byte(config.Options.FlagBaseURL + "/" + key))
		if err != nil {
			log.Panicf("Unable to write reponse: %v", err)
			http.Error(w, "Unable to write reponse", http.StatusNotFound)
		}
	}

	return http.HandlerFunc(fn)
}

func GetRequestHandler(store *stores.URLStore) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		val, ok := store.Get(id)
		if !ok {
			http.Error(w, "No info about requested route", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", ContentTypePlainText)
		w.Header().Set("Location", val)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

	return http.HandlerFunc(fn)
}