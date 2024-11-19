// Package middleware provides reusable HTTP middleware components for the URL shortening service.
// These middleware functions extend the functionality of HTTP handlers by adding features such
// as authentication, gzip compression, and logging.
//
// # Middleware Components
//
// The middleware package includes the following key components:
//
//   - `GiveAuthTokenToUser`: Ensures that each user has a valid authentication token
//     (JWT) in their cookies. If a token is not present, a new token is generated and
//     added to the user's cookies.
//
//   - `CheckAuthToken`: Validates the presence of a valid authentication token in the
//     user's cookies. Requests without a valid token are rejected with a `401 Unauthorized`
//     status.
//
//   - `GzipMiddleware`: Compresses HTTP responses using Gzip for clients that support it.
//     It also decompresses incoming Gzip-encoded request bodies to make them accessible
//     to handlers.
//
// # Usage
//
// Middleware components are designed to be used with HTTP routers like `chi.Router`.
// They wrap HTTP handlers to add functionality.
//
// # Example Usage
//
// Applying middleware to an HTTP router:
//
//	package main
//
//	import (
//	    "net/http"
//	    "github.com/go-chi/chi"
//	    "github.com/golangTroshin/shorturl/internal/app/middleware"
//	)
//
//	func main() {
//	    r := chi.NewRouter()
//
//	    // Apply middleware
//	    r.Use(middleware.GzipMiddleware)
//	    r.Use(middleware.GiveAuthTokenToUser)
//
//	    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
//	        w.Write([]byte("Hello, World!"))
//	    })
//
//	    http.ListenAndServe(":8080", r)
//	}
//
// # Middleware Details
//
// ## GiveAuthTokenToUser
// This middleware checks if the user's request contains a valid authentication token (JWT) in their cookies.
// If the token is missing, a new one is generated, added to the cookies, and associated with the request's context.
//
// ## CheckAuthToken
// This middleware ensures that requests include a valid authentication token. If the token is missing or invalid,
// the request is rejected with a `401 Unauthorized` status.
//
// ## GzipMiddleware
// This middleware compresses HTTP responses using Gzip when supported by the client. It also handles
// decompression of incoming request bodies that are Gzip-encoded.
//
// # Extensibility
//
// The `middleware` package is designed to be extensible. Developers can add new middleware functions to
// enhance the functionality of the service.
//
// This package is essential for implementing cross-cutting concerns in the URL shortener service, such as
// authentication, request compression, and response logging.
package middleware
