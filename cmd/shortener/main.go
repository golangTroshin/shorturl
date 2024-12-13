// Package main is responsible for initializing storage, router, and configuration
// for the URL shortener service.
//
// It sets up the application entry point, routing, middleware, and server startup.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/handlers"
	"github.com/golangTroshin/shorturl/internal/app/logger"
	"github.com/golangTroshin/shorturl/internal/app/middleware"
	"github.com/golangTroshin/shorturl/internal/app/storage"
)

var (
	buildVersion = "N/A" // build version, set during build time
	buildDate    = "N/A" // build date, set during build time
	buildCommit  = "N/A" // build commit hash, set during build time
)

// main is the entry point of the application.
//
// It performs the following tasks:
//   - Parses configuration values from flags and environment variables using `config.ParseFlags`.
//   - Initializes the storage system based on the provided configuration using `storage.GetStorageByConfig`.
//   - Sets up a background worker for URL deletions using `handlers.StartDeleteWorker`.
//   - Starts the HTTP server with routes defined in the `Router` function.
//
// Logs fatal errors if configuration parsing, storage initialization, or server startup fails.
func main() {
	// Print build information
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := config.ParseFlags(); err != nil {
		log.Printf("error ocured while parsing flags: %v", err)
	}

	store, err := storage.GetStorageByConfig()
	if err != nil {
		log.Printf("failed to init store: %v", err)
	}

	defer storage.CloseDB()

	go handlers.StartDeleteWorker(store)

	if err := http.ListenAndServe(config.Options.FlagServiceAddress, Router(store)); err != nil {
		log.Printf("failed to start server: %v", err)
	}
}

// Router sets up and returns a Chi router instance for the application.
//
// Routes:
//   - POST "/"              : Shortens a URL using `handlers.PostRequestHandler`.
//   - POST "/api/shorten"   : Shortens a URL via API using `handlers.APIPostHandler`.
//   - POST "/api/shorten/batch" : Shortens multiple URLs in a batch via API using `handlers.APIPostBatchHandler`.
//   - GET "/{id}"           : Retrieves the original URL by its short ID using `handlers.GetRequestHandler`.
//   - GET "/ping"           : Performs a database health check using `handlers.DatabasePing`.
//   - GET "/api/user/urls"  : Retrieves URLs created by the authenticated user using `handlers.GetURLsByUserHandler`.
//   - DELETE "/api/user/urls": Deletes multiple URLs created by the authenticated user using `handlers.APIDeleteUrlsHandler`.
//
// Middleware:
//   - Applies gzip compression using `middleware.GzipMiddleware`.
//   - Logs incoming requests using `logger.LoggingWrapper`.
//   - Validates and provides authentication tokens for certain routes using `middleware.GiveAuthTokenToUser` and `middleware.CheckAuthToken`.
//
// Parameters:
//   - store: An implementation of the `storage.Storage` interface used for storing and retrieving URL data.
//
// Returns:
//   - A configured `chi.Router` instance.
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
