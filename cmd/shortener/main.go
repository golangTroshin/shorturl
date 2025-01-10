// Package main is responsible for initializing storage, router, and configuration
// for the URL shortener service.
//
// It sets up the application entry point, routing, middleware, and server startup.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/golangTroshin/shorturl/internal/app/config"
	grpcServer "github.com/golangTroshin/shorturl/internal/app/grpc/handlers"
	interceptor "github.com/golangTroshin/shorturl/internal/app/grpc/interceptor"
	shortener "github.com/golangTroshin/shorturl/internal/app/grpc/proto"
	"github.com/golangTroshin/shorturl/internal/app/service"

	"github.com/golangTroshin/shorturl/internal/app/helpers"
	"github.com/golangTroshin/shorturl/internal/app/http/handlers"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/golangTroshin/shorturl/internal/app/logger"
	storageSvc "github.com/golangTroshin/shorturl/internal/app/storage"
	"google.golang.org/grpc"
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
//   - Initializes the storage system based on the provided configuration using `storageSvc.GetStorageByConfig`.
//   - Sets up a background worker for URL deletions using `service.StartDeleteWorker`.
//   - Starts the HTTP server with routes defined in the `Router` function.
//
// Logs errors if configuration parsing, storage initialization, or server startup fails.
func main() {
	// Print build information
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := config.ParseFlags(); err != nil {
		log.Printf("error occurred while parsing flags: %v", err)
	}

	storage, err := storageSvc.GetStorageByConfig()
	svc := service.NewURLService(storage)

	if err != nil {
		log.Printf("failed to initialize storage: %v", err)
	}
	defer storageSvc.CloseDB()

	go service.StartDeleteWorker(storage)

	// Create context with cancellation
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Create a gRPC server with interceptors
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.GiveAuthTokenToUserInterceptor, // Generates the token
			interceptor.CheckAuthTokenInterceptor,      // Validates the token
		),
	)
	shortener.RegisterShortenerServer(grpcSrv, grpcServer.NewShortenerServer(svc))
	go func() {
		log.Println("gRPC server is running on :50051")
		if err := grpcSrv.Serve(grpcListener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server
	srv := &http.Server{
		Addr:    config.Options.FlagServiceAddress,
		Handler: Router(svc),
	}

	go func() {
		var err error
		if config.Options.EnableHTTPS {
			crt, key := helpers.GetTLSCertificate()
			helpers.SaveToFile("crt", crt)
			helpers.SaveToFile("key", key)
			err = srv.ListenAndServeTLS("crt", "key")
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	log.Println("Server is running...")

	// Wait for termination signal
	<-ctx.Done()
	log.Println("Shutdown signal received")

	// Create context for server shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
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
//   - storage: An implementation of the `storageSvc.Storage` interface used for storing and retrieving URL data.
//
// Returns:
//   - A configured `chi.Router` instance.
func Router(svc service.Service) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware, logger.LoggingWrapper)

	r.With(middleware.GiveAuthTokenToUser).Post("/", handlers.ShortenURL(svc))
	r.With(middleware.GiveAuthTokenToUser).Post("/api/shorten", handlers.APIShortenURL(svc))
	r.With(middleware.GiveAuthTokenToUser).Post("/api/shorten/batch", handlers.APIPostBatchHandler(svc))
	r.With(middleware.IPTrustedMiddleware).Get("/api/internal/stats", handlers.APIInternalGetStatsHandler(svc))

	r.Get("/{id}", handlers.GetOriginalURL(svc))
	r.Get("/ping", handlers.Ping(svc))
	r.With(middleware.CheckAuthToken).Get("/api/user/urls", handlers.GetUserURLs(svc))
	r.With(middleware.CheckAuthToken).Delete("/api/user/urls", handlers.APIDeleteUrlsHandler(svc))

	return r
}
