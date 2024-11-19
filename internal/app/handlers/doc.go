// Package handlers provides HTTP handler functions for the URL shortening service.
// These handlers process incoming HTTP requests, interact with the underlying storage layer,
// and generate appropriate HTTP responses.
//
// # Handler Functions
//
// The `handlers` package includes the following key handler functions:
//
// - `APIPostHandler`: Handles requests for creating a shortened URL from an original URL.
// - `APIPostBatchHandler`: Handles batch requests for creating multiple shortened URLs.
// - `APIDeleteUrlsHandler`: Handles requests to mark specific shortened URLs as deleted.
// - `GetRequestHandler`: Handles requests for retrieving the original URL from a shortened URL.
// - `GetURLsByUserHandler`: Handles requests to retrieve all URLs associated with a user.
// - `DatabasePing`: Provides a health-check endpoint for verifying database connectivity.
//
// # Usage
//
// Handlers are typically mapped to specific endpoints in an HTTP router such as `chi.Router`.
// These handlers rely on the storage layer for persisting and retrieving URL data.
//
// # Example Usage
//
// Mapping handlers to endpoints:
//
//	package main
//
//	import (
//	    "net/http"
//	    "github.com/go-chi/chi"
//	    "github.com/golangTroshin/shorturl/internal/app/handlers"
//	    "github.com/golangTroshin/shorturl/internal/app/storage"
//	)
//
//	func main() {
//	    store := storage.NewMemoryStore()
//	    r := chi.NewRouter()
//
//	    r.Post("/api/shorten", handlers.APIPostHandler(store))
//	    r.Post("/api/shorten/batch", handlers.APIPostBatchHandler(store))
//	    r.Get("/{id}", handlers.GetRequestHandler(store))
//
//	    http.ListenAndServe(":8080", r)
//	}
//
// # Handler Details
//
// ## APIPostHandler
// Accepts a JSON payload with an original URL and returns a shortened URL. If the URL already exists in the database,
// the existing shortened URL is returned with a `409 Conflict` status.
//
// ## APIPostBatchHandler
// Accepts a batch of JSON payloads, each containing an original URL, and returns a batch of shortened URLs.
// Useful for bulk URL shortening requests.
//
// ## APIDeleteUrlsHandler
// Marks specified shortened URLs as deleted for a given user. Deleted URLs are not retrievable via the API.
//
// ## GetRequestHandler
// Retrieves the original URL corresponding to a given shortened URL. Responds with a `307 Temporary Redirect`
// if the URL is found or a `410 Gone` status if the URL has been marked as deleted.
//
// ## GetURLsByUserHandler
// Retrieves all URLs associated with the authenticated user. Returns `204 No Content` if no URLs are found.
//
// ## DatabasePing
// Provides a health-check endpoint for verifying connectivity to the database. Responds with a `500 Internal Server Error`
// if the database is unreachable.
//
// # Extensibility
//
// The `handlers` package is designed to be extensible, allowing developers to add new handlers for additional
// API endpoints or modify existing ones as needed.
//
// This package serves as the primary interface between the client-facing API and the underlying storage layer,
// ensuring a clear separation of concerns and ease of maintenance.
package handlers
