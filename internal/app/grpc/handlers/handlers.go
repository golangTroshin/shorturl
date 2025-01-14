package grpc

import (
	"context"
	"log"

	"github.com/golangTroshin/shorturl/internal/app/config"
	shortener "github.com/golangTroshin/shorturl/internal/app/grpc/proto"
	"github.com/golangTroshin/shorturl/internal/app/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ShortenerServer implements the gRPC service for URL shortening.
//
// This server provides methods for shortening URLs, retrieving original URLs,
// deleting user URLs, fetching statistics, and checking the health of the database.
// It integrates with the `storage.Storage` interface to perform persistence operations.
type ShortenerServer struct {
	shortener.UnimplementedShortenerServer
	svc service.Service
}

// NewShortenerServer initializes a new ShortenerServer.
func NewShortenerServer(svc service.Service) *ShortenerServer {
	return &ShortenerServer{svc: svc}
}

// ShortenURL creates a shortened URL for the given original URL.
//
// This method processes a `ShortenURLRequest` containing the original URL,
// stores the URL mapping in the underlying storage, and returns the shortened URL.
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *shortener.ShortenURLRequest) (*shortener.ShortenURLResponse, error) {
	URL, err := s.svc.ShortenURL(ctx, req.Url)
	if err != nil {
		return nil, err
	}
	return &shortener.ShortenURLResponse{ShortUrl: URL.ShortURL}, nil
}

// GetOriginalURL retrieves the original URL for a given shortened URL.
//
// This method processes a `GetOriginalURLRequest` containing the shortened URL key,
// queries the underlying storage for the corresponding original URL, and returns it.
func (s *ShortenerServer) GetOriginalURL(ctx context.Context, req *shortener.GetOriginalURLRequest) (*shortener.GetOriginalURLResponse, error) {
	originalURL, err := s.svc.GetOriginalURL(ctx, req.ShortUrl)
	if err != nil {
		return nil, err
	}
	return &shortener.GetOriginalURLResponse{OriginalUrl: originalURL}, nil
}

// GetURLsByUser retrieves URLs associated with a given user ID.
func (s *ShortenerServer) GetUserURLs(ctx context.Context, req *shortener.GetUserURLsRequest) (*shortener.GetUserURLsResponse, error) {
	urls, err := s.svc.GetUserURLs(ctx)
	if err != nil {
		log.Printf("getURLsByUser: Error fetching URLs for user %s", err)
		return nil, err
	}

	// Convert storage URLs to gRPC response format
	var responseURLs []*shortener.URL
	for _, url := range urls {
		responseURLs = append(responseURLs, &shortener.URL{
			ShortUrl:    config.Options.FlagBaseURL + "/" + url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &shortener.GetUserURLsResponse{
		Urls: responseURLs,
	}, nil
}

// DeleteUserURLs handles a gRPC request to delete a batch of URLs for a user.
//
// This method processes a `DeleteUserURLsRequest` containing a list of URL IDs to delete.
// The URLs are added to a deletion queue for asynchronous processing.
func (s *ShortenerServer) DeleteUserURLs(ctx context.Context, req *shortener.DeleteUserURLsRequest) (*shortener.DeleteUserURLsResponse, error) {
	err := s.svc.DeleteUserURLs(ctx, req.ShortUrls)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	return &shortener.DeleteUserURLsResponse{Success: true}, nil
}

// GetStats handles a gRPC request to retrieve URL and user statistics.
//
// This method fetches statistical information about the total number of URLs
// and unique users from the storage layer.
func (s *ShortenerServer) GetStats(ctx context.Context, req *shortener.GetStatsRequest) (*shortener.GetStatsResponse, error) {
	stats, err := s.svc.GetStats(ctx)
	if err != nil {
		log.Printf("Error fetching stats: %v", err)
		return nil, status.Errorf(codes.Internal, "Unable to fetch statistics")
	}

	log.Printf("Stats retrieved: URLs: %d, Users: %d", stats.Urls, stats.Users)
	return &shortener.GetStatsResponse{
		Urls:  int32(stats.Urls),
		Users: int32(stats.Users),
	}, nil
}

// Ping handles a gRPC request to check the health of the database connection.
//
// This method attempts to establish a connection to the database and perform a health check.
// It responds with a success message if the database is reachable.
func (s *ShortenerServer) Ping(ctx context.Context, req *shortener.PingRequest) (*shortener.PingResponse, error) {
	err := s.svc.PingDatabase(ctx)

	if err != nil {
		log.Printf("Database ping error: %v", err)
		return nil, status.Errorf(codes.Unavailable, "%s", err.Error())
	}

	log.Println("Database is healthy")
	return &shortener.PingResponse{Status: "OK"}, nil
}
