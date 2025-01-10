package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
	"github.com/golangTroshin/shorturl/internal/app/service"
	"github.com/golangTroshin/shorturl/internal/app/storage"
	"github.com/golangTroshin/shorturl/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestShortenURL(t *testing.T) {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)

	handler := ShortenURL(svc)

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("Successful URL Shortening", func(t *testing.T) {
		reqBody := []byte("https://example.com")
		resp, err := http.Post(server.URL, "text/plain", bytes.NewReader(reqBody))

		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "/")
	})

	t.Run("Empty Body", func(t *testing.T) {
		resp, err := http.Post(server.URL, "text/plain", nil)

		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetOriginalURL(t *testing.T) {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "test-user")

	// Store a URL
	url, err := store.Set(ctx, "https://example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, url.ShortURL)

	// Simulate routing with chi
	router := chi.NewRouter()
	router.Get("/{id}", GetOriginalURL(svc))

	// Create a request to the short URL
	req := httptest.NewRequest(http.MethodGet, "/"+url.ShortURL, nil)
	recorder := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(recorder, req)

	// Assert the response
	assert.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	assert.Equal(t, "https://example.com", recorder.Header().Get("Location"))
}
func TestGetUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := GetUserURLs(mockService)

	t.Run("Successful retrieval of user URLs", func(t *testing.T) {
		// Mock service to return a list of URLs
		mockService.EXPECT().GetUserURLs(gomock.Any()).Return([]storage.URL{
			{ShortURL: "short1", OriginalURL: "http://example.com/1"},
			{ShortURL: "short2", OriginalURL: "http://example.com/2"},
		}, nil)

		// Simulate a request
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Validate response
		assert.Equal(t, http.StatusOK, rec.Code)

		var response []storage.URL
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Len(t, response, 2)
		assert.Equal(t, "short1", response[0].ShortURL)
		assert.Equal(t, "http://example.com/1", response[0].OriginalURL)
		assert.Equal(t, "short2", response[1].ShortURL)
		assert.Equal(t, "http://example.com/2", response[1].OriginalURL)
	})

	t.Run("No URLs for user", func(t *testing.T) {
		// Mock service to return an empty list
		mockService.EXPECT().GetUserURLs(gomock.Any()).Return([]storage.URL{}, nil)

		// Simulate a request
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Validate response
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("Service error", func(t *testing.T) {
		// Mock service to return an error
		mockService.EXPECT().GetUserURLs(gomock.Any()).Return(nil, context.DeadlineExceeded)

		// Simulate a request
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Validate response
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}

func TestPing(t *testing.T) {
	store := storage.NewMemoryStore()
	svc := service.NewURLService(store)
	handler := Ping(svc)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
