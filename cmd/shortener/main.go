package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/handlers"
	"github.com/golangTroshin/shorturl/internal/app/logger"
	"github.com/golangTroshin/shorturl/internal/app/stores"
)

func main() {
	if err := config.ParseFlags(); err != nil {
		log.Fatalf("error ocured while parsing flags: %v", err)
	}

	store := stores.NewURLStore()

	if err := http.ListenAndServe(config.Options.FlagServiceAddress, Router(store)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func Router(store *stores.URLStore) chi.Router {
	r := chi.NewRouter()

	r.Post("/", logger.LoggingWrapper(handlers.PostRequestHandler(store)))
	r.Get("/{id}", logger.LoggingWrapper(handlers.GetRequestHandler(store)))

	return r
}
