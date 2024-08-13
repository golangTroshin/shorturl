package handlers

import (
	"database/sql"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const ContentTypePlainText = "text/plain"

func PostRequestHandler(storage storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}

		url, err := storage.Set(r.Context(), string(body))
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Type", ContentTypePlainText)
		w.WriteHeader(http.StatusCreated)

		_, err = w.Write([]byte(config.Options.FlagBaseURL + "/" + url.ShortURL))
		if err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, "Unable to write reponse", http.StatusNotFound)
		}
	}

	return http.HandlerFunc(fn)
}

func GetRequestHandler(storage storage.Storage) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		val, err := storage.Get(r.Context(), id)
		if err != nil {
			http.Error(w, "No info about requested route", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", ContentTypePlainText)
		w.Header().Set("Location", val)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

	return http.HandlerFunc(fn)
}

func DatabasePing() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("pgx", config.Options.DatabaseDsn)

		if err != nil {
			http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		}

		err = db.Ping()
		if err != nil {
			log.Printf("Unable to write reponse: %v", err)
			http.Error(w, "Unable to reach database", http.StatusInternalServerError)
		}

		defer db.Close()
	}

	return http.HandlerFunc(fn)
}
