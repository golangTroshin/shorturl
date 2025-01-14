package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/golangTroshin/shorturl/internal/app/service"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	"github.com/golangTroshin/shorturl/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestShortenURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	svc := service.NewURLService(mockStorage)

	t.Run("Shorten URL successfully", func(t *testing.T) {
		mockStorage.EXPECT().Set(gomock.Any(), "http://example.com").Return(
			storage.URL{ShortURL: "short123", OriginalURL: "http://example.com"}, nil,
		)

		result, err := svc.ShortenURL(context.Background(), "http://example.com")

		assert.NoError(t, err)
		assert.Equal(t, "short123", result.ShortURL)
		assert.Equal(t, "http://example.com", result.OriginalURL)
	})

	t.Run("Error shortening URL", func(t *testing.T) {
		mockStorage.EXPECT().Set(gomock.Any(), "http://example.com").Return(storage.URL{}, errors.New("storage error"))

		result, err := svc.ShortenURL(context.Background(), "http://example.com")

		assert.Error(t, err)
		assert.Equal(t, storage.URL{}, result)
	})
}

func TestGetOriginalURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	svc := service.NewURLService(mockStorage)

	t.Run("Get original URL successfully", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "short123").Return("http://example.com", nil)

		result, err := svc.GetOriginalURL(context.Background(), "short123")

		assert.NoError(t, err)
		assert.Equal(t, "http://example.com", result)
	})

	t.Run("Error retrieving original URL", func(t *testing.T) {
		mockStorage.EXPECT().Get(gomock.Any(), "short123").Return("", errors.New("not found"))

		result, err := svc.GetOriginalURL(context.Background(), "short123")

		assert.Error(t, err)
		assert.Equal(t, "", result)
	})
}

func TestBatchShortenURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	svc := service.NewURLService(mockStorage)

	batch := []storage.RequestBodyBanch{
		{CorrelationID: "id1", OriginalURL: "http://example.com/1"},
		{CorrelationID: "id2", OriginalURL: "http://example.com/2"},
	}

	t.Run("Batch shorten URLs successfully", func(t *testing.T) {
		mockStorage.EXPECT().SetBatch(gomock.Any(), batch).Return([]storage.URL{
			{UUID: "id1", ShortURL: "short1", OriginalURL: "http://example.com/1"},
			{UUID: "id2", ShortURL: "short2", OriginalURL: "http://example.com/2"},
		}, nil)

		result, err := svc.BatchShortenURLs(context.Background(), batch)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "short1", result[0].ShortURL)
		assert.Equal(t, "http://example.com/1", result[0].OriginalURL)
	})

	t.Run("Error in batch shortening URLs", func(t *testing.T) {
		mockStorage.EXPECT().SetBatch(gomock.Any(), batch).Return(nil, errors.New("batch error"))

		result, err := svc.BatchShortenURLs(context.Background(), batch)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	svc := service.NewURLService(mockStorage)

	t.Run("Get user URLs successfully", func(t *testing.T) {
		mockStorage.EXPECT().GetByUserID(gomock.Any(), "user123").Return([]storage.URL{
			{ShortURL: "short1", OriginalURL: "http://example.com/1"},
			{ShortURL: "short2", OriginalURL: "http://example.com/2"},
		}, nil)

		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user123")
		result, err := svc.GetUserURLs(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "short1", result[0].ShortURL)
	})

	t.Run("Error retrieving user URLs", func(t *testing.T) {
		mockStorage.EXPECT().GetByUserID(gomock.Any(), "user123").Return(nil, errors.New("storage error"))

		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user123")
		result, err := svc.GetUserURLs(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDeleteUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	svc := service.NewURLService(mockStorage)

	t.Run("Delete user URLs successfully", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user123")
		err := svc.DeleteUserURLs(ctx, []string{"short1", "short2"})

		assert.NoError(t, err)
	})

	t.Run("Error deleting user URLs (empty list)", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user123")
		err := svc.DeleteUserURLs(ctx, []string{})

		assert.Error(t, err)
		assert.Equal(t, "no URLs provided for deletion", err.Error())
	})
}

func TestGetStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	svc := service.NewURLService(mockStorage)

	t.Run("Get stats successfully", func(t *testing.T) {
		mockStorage.EXPECT().GetStats(gomock.Any()).Return(storage.Stats{
			Urls:  100,
			Users: 10,
		}, nil)

		result, err := svc.GetStats(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 100, result.Urls)
		assert.Equal(t, 10, result.Users)
	})

	t.Run("Error getting stats", func(t *testing.T) {
		mockStorage.EXPECT().GetStats(gomock.Any()).Return(storage.Stats{}, errors.New("storage error"))

		result, err := svc.GetStats(context.Background())

		assert.Error(t, err)
		assert.Equal(t, storage.Stats{}, result)
	})
}
