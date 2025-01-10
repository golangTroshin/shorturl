package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	grpc "github.com/golangTroshin/shorturl/internal/app/grpc/handlers"
	shortener "github.com/golangTroshin/shorturl/internal/app/grpc/proto"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	"github.com/golangTroshin/shorturl/internal/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestShortenerServer_ShortenURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := grpc.NewShortenerServer(mockService)

	t.Run("Successful URL shortening", func(t *testing.T) {
		mockService.EXPECT().ShortenURL(gomock.Any(), "http://example.com").Return(
			storage.URL{ShortURL: "short123"}, nil,
		)

		req := &shortener.ShortenURLRequest{Url: "http://example.com"}
		resp, err := server.ShortenURL(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "short123", resp.ShortUrl)
	})

	t.Run("Error during URL shortening", func(t *testing.T) {
		mockService.EXPECT().ShortenURL(gomock.Any(), "http://example.com").Return(storage.URL{}, errors.New("internal error"))

		req := &shortener.ShortenURLRequest{Url: "http://example.com"}
		resp, err := server.ShortenURL(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestShortenerServer_GetOriginalURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := grpc.NewShortenerServer(mockService)

	t.Run("Successful retrieval of original URL", func(t *testing.T) {
		mockService.EXPECT().GetOriginalURL(gomock.Any(), "short123").Return("http://example.com", nil)

		req := &shortener.GetOriginalURLRequest{ShortUrl: "short123"}
		resp, err := server.GetOriginalURL(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "http://example.com", resp.OriginalUrl)
	})

	t.Run("Error retrieving original URL", func(t *testing.T) {
		mockService.EXPECT().GetOriginalURL(gomock.Any(), "short123").Return("", errors.New("not found"))

		req := &shortener.GetOriginalURLRequest{ShortUrl: "short123"}
		resp, err := server.GetOriginalURL(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestShortenerServer_GetUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := grpc.NewShortenerServer(mockService)

	t.Run("Successful retrieval of user URLs", func(t *testing.T) {
		mockService.EXPECT().GetUserURLs(gomock.Any()).Return([]storage.URL{
			{ShortURL: "short1", OriginalURL: "http://example.com/1"},
			{ShortURL: "short2", OriginalURL: "http://example.com/2"},
		}, nil)

		req := &shortener.GetUserURLsRequest{}
		resp, err := server.GetUserURLs(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Urls, 2)
		assert.Equal(t, "http://localhost/short1", "http://localhost"+resp.Urls[0].ShortUrl)
		assert.Equal(t, "http://example.com/1", resp.Urls[0].OriginalUrl)
	})

	t.Run("Error retrieving user URLs", func(t *testing.T) {
		mockService.EXPECT().GetUserURLs(gomock.Any()).Return(nil, errors.New("internal error"))

		req := &shortener.GetUserURLsRequest{}
		resp, err := server.GetUserURLs(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestShortenerServer_DeleteUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := grpc.NewShortenerServer(mockService)

	t.Run("Successful deletion of user URLs", func(t *testing.T) {
		// Mock the service to return nil for a successful deletion
		mockService.EXPECT().DeleteUserURLs(gomock.Any(), []string{"short1", "short2"}).Return(nil)

		// Call the method
		req := &shortener.DeleteUserURLsRequest{ShortUrls: []string{"short1", "short2"}}
		resp, err := server.DeleteUserURLs(context.Background(), req)

		// Assert the response
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("Error during URL deletion", func(t *testing.T) {
		// Mock the service to return an error
		mockService.EXPECT().DeleteUserURLs(gomock.Any(), []string{"short1", "short2"}).Return(errors.New("failed to delete URLs"))

		// Call the method
		req := &shortener.DeleteUserURLsRequest{ShortUrls: []string{"short1", "short2"}}
		resp, err := server.DeleteUserURLs(context.Background(), req)

		// Assert the response
		assert.Error(t, err)
		assert.Nil(t, resp)

		// Check gRPC error details
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Equal(t, "failed to delete URLs", st.Message())
	})
}

func TestShortenerServer_GetStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	server := grpc.NewShortenerServer(mockService)

	t.Run("Successful stats retrieval", func(t *testing.T) {
		mockService.EXPECT().GetStats(gomock.Any()).Return(storage.Stats{Urls: 100, Users: 10}, nil)

		req := &shortener.GetStatsRequest{}
		resp, err := server.GetStats(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(100), resp.Urls)
		assert.Equal(t, int32(10), resp.Users)
	})

	t.Run("Error fetching stats", func(t *testing.T) {
		mockService.EXPECT().GetStats(gomock.Any()).Return(storage.Stats{}, errors.New("internal error"))

		req := &shortener.GetStatsRequest{}
		resp, err := server.GetStats(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
