package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golangTroshin/shorturl/internal/app/http/handlers"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	"github.com/golangTroshin/shorturl/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAPIShortenURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := handlers.APIShortenURL(mockService)

	t.Run("Successful URL shortening", func(t *testing.T) {
		mockService.EXPECT().ShortenURL(gomock.Any(), "http://example.com").Return(
			storage.URL{ShortURL: "short123"}, nil,
		)

		body := `{"url": "http://example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader([]byte(body)))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var response storage.ResponseShortURL
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost/short123", "http://localhost"+response.ShortURL)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader([]byte("invalid body")))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAPIPostBatchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := handlers.APIPostBatchHandler(mockService)

	t.Run("Successful batch shortening", func(t *testing.T) {
		mockService.EXPECT().BatchShortenURLs(gomock.Any(), gomock.Any()).Return(
			[]storage.URL{
				{UUID: "id1", ShortURL: "short1", OriginalURL: "http://example.com/1"},
				{UUID: "id2", ShortURL: "short2", OriginalURL: "http://example.com/2"},
			}, nil,
		)

		body := `[{"correlation_id": "id1", "url": "http://example.com/1"}, {"correlation_id": "id2", "url": "http://example.com/2"}]`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader([]byte(body)))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var response []map[string]string
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, "http://localhost/short1", "http://localhost"+response[0]["short_url"])
		assert.Equal(t, "id1", response[0]["correlation_id"])
		assert.Equal(t, "http://localhost/short2", "http://localhost"+response[1]["short_url"])
		assert.Equal(t, "id2", response[1]["correlation_id"])
	})

	t.Run("Invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader([]byte("invalid body")))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAPIDeleteUrlsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := handlers.APIDeleteUrlsHandler(mockService)

	t.Run("Successful deletion", func(t *testing.T) {
		mockService.EXPECT().DeleteUserURLs(gomock.Any(), []string{"short1", "short2"}).Return(nil)

		body := `["short1", "short2"]`
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader([]byte(body)))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader([]byte("invalid body")))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAPIInternalGetStatsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := handlers.APIInternalGetStatsHandler(mockService)

	t.Run("Successful stats retrieval", func(t *testing.T) {
		mockService.EXPECT().GetStats(gomock.Any()).Return(storage.Stats{
			Urls:  100,
			Users: 10,
		}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var response map[string]int
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, 100, response["urls"])
		assert.Equal(t, 10, response["users"])
	})

	t.Run("No stats available", func(t *testing.T) {
		mockService.EXPECT().GetStats(gomock.Any()).Return(storage.Stats{}, errors.New("no stats available"))

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}
