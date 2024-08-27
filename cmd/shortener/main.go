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

	r.Post("/", middleware.ChainMiddlewares(handlers.PostRequestHandler(store),
		middleware.GiveAuthTokenToUser,
		middleware.GzipMiddleware,
		logger.LoggingWrapper,
	))

	r.Post("/api/shorten", middleware.ChainMiddlewares(handlers.APIPostHandler(store),
		middleware.GiveAuthTokenToUser,
		middleware.GzipMiddleware,
		logger.LoggingWrapper,
	))

	r.Post("/api/shorten/batch", middleware.ChainMiddlewares(handlers.APIPostBatchHandler(store),
		middleware.GiveAuthTokenToUser,
		middleware.GzipMiddleware,
		logger.LoggingWrapper,
	))

	r.Get("/{id}", logger.LoggingWrapper(middleware.GzipMiddleware(handlers.GetRequestHandler(store))))
	r.Get("/api/user/urls", logger.LoggingWrapper(middleware.GzipMiddleware(handlers.GetURLsByUserHandler(store))))
	r.Get("/ping", logger.LoggingWrapper(handlers.DatabasePing()))

	r.Delete("/api/user/urls", middleware.ChainMiddlewares(handlers.APIDeleteUrlsHandler(store),
		middleware.GzipMiddleware,
		logger.LoggingWrapper,
	))

	return r
}
