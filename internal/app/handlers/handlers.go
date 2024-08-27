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

const ContentTypePlainText = "text/plain"

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

func GetURLsByUserHandler(store storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authToken, err := r.Cookie(middleware.CookieAuthToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if authToken.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		urls, err := store.GetByUserID(r.Context(), authToken.Value)
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
