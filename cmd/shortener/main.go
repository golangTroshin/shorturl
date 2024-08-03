package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/handlers"
	"github.com/golangTroshin/shorturl/internal/app/logger"
	"github.com/golangTroshin/shorturl/internal/app/middleware"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

func main() {
	if err := config.ParseFlags(); err != nil {
		log.Fatalf("error ocured while parsing flags: %v", err)
	}

	store, err := storage.InitURLStore()
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}

	if err := http.ListenAndServe(config.Options.FlagServiceAddress, Router(store)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func Router(store *storage.URLStore) chi.Router {
	r := chi.NewRouter()

	r.Post("/", logger.LoggingWrapper(middleware.GzipMiddleware(handlers.PostRequestHandler(store))))
	r.Post("/api/shorten", logger.LoggingWrapper(middleware.GzipMiddleware(handlers.APIPostHandler(store))))

	r.Get("/{id}", logger.LoggingWrapper(middleware.GzipMiddleware(handlers.GetRequestHandler(store))))

	return r
}
