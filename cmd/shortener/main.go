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

	store, err := storage.GetStorageByConfig()
	if err != nil {
		log.Fatalf("failed to init store: %v", err)
	}

	defer storage.CloseDB()

	go handlers.StartDeleteWorker(store)

	if err := http.ListenAndServe(config.Options.FlagServiceAddress, Router(store)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func Router(store storage.Storage) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware, logger.LoggingWrapper)

	r.With(middleware.GiveAuthTokenToUser).Post("/", handlers.PostRequestHandler(store))
	r.With(middleware.GiveAuthTokenToUser).Post("/api/shorten", handlers.APIPostHandler(store))
	r.With(middleware.GiveAuthTokenToUser).Post("/api/shorten/batch", handlers.APIPostBatchHandler(store))

	r.Get("/{id}", handlers.GetRequestHandler(store))
	r.Get("/ping", handlers.DatabasePing())
	r.With(middleware.CheckAuthToken).Get("/api/user/urls", handlers.GetURLsByUserHandler(store))
	r.With(middleware.CheckAuthToken).Delete("/api/user/urls", handlers.APIDeleteUrlsHandler(store))

	return r
}
