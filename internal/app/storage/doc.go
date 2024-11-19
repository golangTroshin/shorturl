// Package storage provides an abstraction layer for storing, retrieving, and managing
// URL mappings in a URL shortening service. It supports multiple storage backends,
// including memory-based, file-based, and database-backed implementations.
//
// The storage system is designed to handle single URL operations, batch operations,
// and user-specific URL management. It also supports marking URLs as deleted.
//
// # Storage Implementations
//
// The `Storage` interface defines the core operations for managing URLs. The package
// provides the following implementations:
//   - MemoryStore: An in-memory storage suitable for development or small-scale use.
//   - FileStore: A file-based storage that persists URLs to a file on disk.
//   - DatabaseStore: A database-backed storage for scalable and reliable use cases.
//
// The storage backend can be selected at runtime based on configuration values.
//
// # Core Types
//
//   - URL: Represents a mapping between a short URL and its original URL, including
//     metadata such as user ownership and deletion status.
//   - RequestURL: Represents an incoming request to shorten a single URL.
//   - ResponseShortURL: Represents the response for a successfully shortened URL.
//   - RequestBodyBanch: Represents a batch request to shorten multiple URLs.
//
// # Key Functions
//
//   - GetStorageByConfig: Initializes the appropriate storage backend based on
//     configuration values (e.g., database DSN, file path).
//   - generateShortURL: Generates a unique short URL key for a given input string using
//     SHA-256 hashing and Base64 encoding.
//
// # Example Usage
//
// Selecting a storage backend and performing operations:
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "log"
//	    "github.com/golangTroshin/shorturl/internal/app/storage"
//	)
//
//	func main() {
//	    // Initialize storage based on configuration
//	    store, err := storage.GetStorageByConfig()
//	    if err != nil {
//	        log.Fatalf("Failed to initialize storage: %v", err)
//	    }
//
//	    // Save a new URL
//	    ctx := context.Background()
//	    url, err := store.Set(ctx, "https://example.com")
//	    if err != nil {
//	        log.Fatalf("Failed to save URL: %v", err)
//	    }
//	    fmt.Println("Shortened URL:", url.ShortURL)
//
//	    // Retrieve the original URL
//	    originalURL, err := store.Get(ctx, url.ShortURL)
//	    if err != nil {
//	        log.Fatalf("Failed to retrieve URL: %v", err)
//	    }
//	    fmt.Println("Original URL:", originalURL)
//	}
//
// # Extensibility
//
// The `Storage` interface can be extended to support new storage backends by
// implementing its methods. This allows the service to adapt to changing requirements
// or integrate with new data storage systems.
//
// This package is a critical component of the URL shortener service, providing
// a flexible and extensible foundation for managing URL mappings.
package storage
